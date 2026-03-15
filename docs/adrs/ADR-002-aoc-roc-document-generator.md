# ADR-002: AOC/ROC Document Generator Architecture

## Status

Accepted

## Date

2026-03-15

## Context

PCI DSS v4.0.1 Requirement 12.4 mandates the generation of two specific document types as output of a compliance assessment:

**AOC (Attestation of Compliance)** — A standardized PCI SSC form that a merchant or service provider signs to attest they have completed a SAQ or ROC. Different SAQ variants apply to different merchant types:
- SAQ A: Card-not-present merchants, fully outsourced
- SAQ A-EP: E-commerce with payment page redirect
- SAQ B: Imprint machines or standalone dial-out terminals
- SAQ B-IP: IP-connected payment terminals
- SAQ C-VT: Virtual payment terminals
- SAQ C: Payment application systems
- SAQ D: Merchants and service providers not covered by other SAQs

**ROC (Report on Compliance)** — A detailed assessment report produced by a QSA (Qualified Security Assessor) for Level 1 merchants. Structured into sections mirroring the 12 PCI DSS requirement domains, with testing procedures and findings for each.

Current state: Raisin Protect produces generic PDF exports from the reporting module. These are not PCI Council-mandated formats. A QSA receiving a generic PDF compliance report cannot use it to fulfill assessment obligations.

The AOC/ROC generator must:
1. Aggregate data from controls, evidence, requirements, audit findings, and the new CDE scoping module (ADR-001)
2. Render structured, PCI SSC-compliant document templates
3. Support a review and approval workflow before document finalization
4. Capture QSA signature and attestation fields
5. Store generated documents with versioning (compliance documents must be retained for reference and renewal)

---

## Decision

Implement the AOC/ROC Document Generator using a **server-side Go template rendering pipeline with MinIO-backed storage**. The approach:

1. Go `html/template` renders a structured HTML document from aggregated compliance data
2. Headless Chromium (via `chromedp`) converts HTML to PDF server-side
3. Generated PDF is stored in MinIO under the existing file storage service
4. A new `compliance_documents` domain tracks document lifecycle (draft → under_review → approved → final)
5. QSA attestation fields are captured in a structured `document_attestations` table before final generation

---

## Template Rendering Approach

### Rationale: Go Templates + chromedp vs. Alternatives

| Approach | Pros | Cons | Decision |
|---|---|---|---|
| **Go `html/template` + chromedp** | Full layout control, pixel-perfect PDF, reuses existing Go stack, CSS for PCI form styling | Requires Chromium in container, slightly higher memory | **Selected** |
| `gofpdf` / `gopdf` | Pure Go, no binary dependency | Manual layout (no CSS), extremely tedious for complex multi-page forms | Rejected |
| LaTeX | Professional typesetting | No Go expertise, complex dependency chain, overkill | Rejected |
| External service (Puppeteer microservice) | Decoupled | New service to operate, network hop, adds infra complexity | Rejected for Sprint 11; revisit if needed |
| WeasyPrint (Python) | Good CSS-to-PDF | Requires Python runtime alongside Go app | Rejected |

**chromedp** runs headless Chrome as a subprocess. The Docker image for Raisin Protect's API service will be extended to include a Chromium installation. This is the same approach used by many production Go services (e.g., reporting tools, invoice generators).

### Template Structure

Templates are stored as embedded Go files (`//go:embed templates/documents/*`) and organized by document type:

```
api/internal/templates/documents/
├── base_layout.html          # Shared header/footer, CSS, page setup
├── aoc/
│   ├── saq_a.html            # SAQ A AOC template
│   ├── saq_a_ep.html
│   ├── saq_b.html
│   ├── saq_b_ip.html
│   ├── saq_c_vt.html
│   ├── saq_c.html
│   └── saq_d.html            # Includes service provider variant
├── roc/
│   ├── roc_cover.html        # Title page, QSA details, merchant info
│   ├── roc_section_executive.html
│   ├── roc_section_req1.html # Req 1: Install/maintain network security controls
│   ├── roc_section_req2.html # ... through req 12
│   └── roc_section_appendix.html
└── shared/
    ├── control_table.html    # Reusable control status table
    ├── evidence_list.html    # Reusable evidence listing
    └── cde_summary.html      # CDE scope summary (from ADR-001 data)
```

Templates receive a strongly-typed `DocumentData` struct populated by the data aggregation service (described below).

---

## Data Model

### New Tables (Sprint 11-12 migrations)

#### `compliance_documents` (migration 056)

The primary record for each generated compliance document.

```sql
CREATE TABLE compliance_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    -- Document classification
    document_type   TEXT NOT NULL CHECK (document_type IN (
                        'aoc_saq_a', 'aoc_saq_a_ep', 'aoc_saq_b', 'aoc_saq_b_ip',
                        'aoc_saq_c_vt', 'aoc_saq_c', 'aoc_saq_d',
                        'aoc_saq_d_sp',   -- Service Provider variant
                        'roc'
                    )),
    title           TEXT NOT NULL,
    -- Assessment period
    assessment_period_start DATE NOT NULL,
    assessment_period_end   DATE NOT NULL,
    pci_dss_version TEXT NOT NULL DEFAULT '4.0.1',
    -- Merchant/SP details (captured at generation time, not referenced from org)
    merchant_name   TEXT,
    merchant_dba    TEXT,
    merchant_url    TEXT,
    business_type   TEXT,
    -- Document lifecycle
    doc_status      TEXT NOT NULL DEFAULT 'draft' CHECK (doc_status IN (
                        'draft',          -- Being configured, not yet generated
                        'generating',     -- PDF generation in progress
                        'under_review',   -- Generated, awaiting internal approval
                        'approved',       -- Internally approved, ready for QSA
                        'qsa_review',     -- With QSA for review/signature
                        'final',          -- QSA-signed, locked, final version
                        'superseded',     -- Replaced by a newer version
                        'cancelled'
                    )),
    -- Versioning
    version         INTEGER NOT NULL DEFAULT 1,
    parent_id       UUID REFERENCES compliance_documents(id),  -- NULL for v1, set for revisions
    -- Storage
    storage_key     TEXT,               -- MinIO object key for generated PDF
    generated_at    TIMESTAMPTZ,
    generation_error TEXT,              -- Set if generation failed
    file_size_bytes BIGINT,
    -- Audit
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_compliance_docs_org ON compliance_documents (org_id);
CREATE INDEX idx_compliance_docs_org_status ON compliance_documents (org_id, doc_status);
CREATE INDEX idx_compliance_docs_org_type ON compliance_documents (org_id, document_type, doc_status);
```

#### `document_attestations` (migration 057)

Captures structured attestation and signature fields for both merchant and QSA signatories. These fields feed into the template rendering.

```sql
CREATE TABLE document_attestations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID NOT NULL REFERENCES compliance_documents(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    attestation_role TEXT NOT NULL CHECK (attestation_role IN (
                        'merchant_signatory',   -- Officer signing the AOC on behalf of merchant
                        'qsa_signatory',        -- QSA signing/validating the assessment
                        'isac_signatory',       -- Internal Security Assessor (for SAQ)
                        'sp_signatory'          -- Service provider representative
                    )),
    -- Signatory details
    full_name       TEXT NOT NULL,
    title           TEXT NOT NULL,
    company_name    TEXT,
    company_address TEXT,
    company_url     TEXT,
    email           TEXT,
    phone           TEXT,
    -- For QSA: assessor credentials
    qsa_company     TEXT,               -- QSA company name
    qsa_number      TEXT,               -- PCI SSC QSA listing number
    -- Signature capture
    signed_at       TIMESTAMPTZ,
    signature_method TEXT CHECK (signature_method IN (
                        'digital_signature',    -- Electronic signature captured
                        'manual',               -- Will be printed and signed by hand
                        'docusign_ref'          -- External e-sign reference
                    )),
    signature_ref   TEXT,               -- DocuSign envelope ID or digital cert ref
    -- Audit
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_attestation_role UNIQUE (document_id, attestation_role)
);

CREATE INDEX idx_document_attestations_doc ON document_attestations (document_id);
```

#### `document_approvals` (migration 057, same file)

Internal approval workflow before the document goes to QSA. This is the organization's internal review step.

```sql
CREATE TABLE document_approvals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID NOT NULL REFERENCES compliance_documents(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    approver_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    approval_status TEXT NOT NULL DEFAULT 'pending' CHECK (approval_status IN (
                        'pending', 'approved', 'rejected', 'withdrawn'
                    )),
    comments        TEXT,
    responded_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_doc_approver UNIQUE (document_id, approver_id)
);

CREATE INDEX idx_doc_approvals_document ON document_approvals (document_id);
CREATE INDEX idx_doc_approvals_approver ON document_approvals (approver_id, approval_status);
```

#### `document_requirement_snapshots` (migration 058)

Captures the compliance posture snapshot at document generation time. This is critical for audit integrity — the document must reflect the state at a point in time, not the live state.

```sql
CREATE TABLE document_requirement_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID NOT NULL REFERENCES compliance_documents(id) ON DELETE CASCADE,
    requirement_id  UUID NOT NULL REFERENCES requirements(id),
    requirement_code TEXT NOT NULL,     -- Denormalized for snapshot integrity
    requirement_title TEXT NOT NULL,
    in_scope        BOOLEAN NOT NULL,
    control_count   INTEGER NOT NULL DEFAULT 0,
    passing_controls INTEGER NOT NULL DEFAULT 0,
    evidence_count  INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL CHECK (status IN (
                        'compliant', 'non_compliant', 'partially_compliant',
                        'not_applicable', 'compensating_control'
                    )),
    notes           TEXT,
    snapshotted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_doc_req_snapshots_document ON document_requirement_snapshots (document_id);
```

---

## Data Aggregation Architecture

Before generating the PDF, a `DocumentAggregator` service collects and denormalizes all required data into a `DocumentData` struct. This is a read-only operation across multiple tables.

```go
// api/internal/services/document_aggregator.go

type DocumentData struct {
    Document      *models.ComplianceDocument
    Organization  *models.Organization
    Attestations  map[string]*models.DocumentAttestation // keyed by role
    PCIDSSVersion string

    // CDE Scope (from ADR-001 tables)
    CDESummary    *CDEScopeSummary
    CDEAssets     []CDEAssetSummary

    // Controls & Requirements
    Requirements  []RequirementWithStatus
    Controls      []ControlSummary
    ControlCount  int
    PassingCount  int

    // Evidence
    EvidenceSummary EvidenceSummary
    KeyEvidence     []EvidenceItem // top evidence items per requirement domain

    // Findings (from Audit Hub)
    OpenFindings   []AuditFindingSummary
    ClosedFindings []AuditFindingSummary

    // Monitoring
    MonitoringSummary MonitoringSummary

    // Generation metadata
    GeneratedAt   time.Time
    GeneratedBy   string
    DocumentID    string
}
```

The aggregator runs within a serializable transaction to ensure consistent point-in-time data. After aggregation, requirement snapshots are written to `document_requirement_snapshots` before PDF rendering begins.

---

## Document Generation Flow

```
User configures document
(type, period, attestations)
        │
        ▼
POST /api/v1/documents
        │
        ▼
document created (status: draft)
        │
        ▼
User adds attestations
POST /api/v1/documents/:id/attestations
        │
        ▼
User requests generation
POST /api/v1/documents/:id/generate
        │
        ▼
Validation (all required attestation fields populated?)
        │ fail → 422 Unprocessable Entity
        │ pass ↓
        ▼
Status → generating
(async job queued via Go goroutine + monitoring worker pattern)
        │
        ▼
DocumentAggregator.Collect()
(serializable transaction, snapshots written)
        │
        ▼
Template.Execute(documentData) → HTML string
        │
        ▼
chromedp.CaptureScreenshot... → PDF bytes
        │
        ▼
MinIO.PutObject(pdfBytes, storageKey)
        │
        ▼
Status → under_review
document.storage_key set, document.generated_at set
Notification sent to approvers
        │
        ▼
Approvers review PDF
POST /api/v1/documents/:id/approvals/:approver_id/respond
(approved / rejected)
        │ all approved ↓       │ any rejected → back to draft
        ▼                      ▼
Status → approved         Status → draft
(ready for QSA)           (regeneration required)
        │
        ▼
QSA reviews and signs
(offline or DocuSign integration in future sprint)
        │
        ▼
Status → final
(locked, no further modifications)
```

**Async generation implementation:** The generation job runs in a goroutine within the existing monitoring worker pattern. For Sprint 11, a simple goroutine with a context timeout is sufficient. If generation volume grows, a task queue (e.g., Redis-backed) can be introduced without changing the API surface.

---

## Versioning and Document Lifecycle

- Each document has an integer `version` starting at 1
- Regenerating an approved or final document creates a new `compliance_documents` record with `parent_id` referencing the original and `version = parent.version + 1`
- The previous document is marked `superseded`
- Document history is queryable via `GET /api/v1/documents?parentId=<id>` or by following the `parent_id` chain
- `final` documents are immutable: no regeneration, no field updates (HTTP 409 if attempted)

---

## API Endpoint Design

```
// Compliance Document Generator (Sprint 11-12)
docs := protected.Group("/documents")

// Document CRUD
docs.GET("",                    middleware.RequireRoles(models.DocumentViewRoles...), handlers.ListDocuments)
docs.POST("",                   middleware.RequireRoles(models.DocumentCreateRoles...), handlers.CreateDocument)
docs.GET("/:id",                middleware.RequireRoles(models.DocumentViewRoles...), handlers.GetDocument)
docs.PUT("/:id",                middleware.RequireRoles(models.DocumentCreateRoles...), handlers.UpdateDocument)
docs.DELETE("/:id",             middleware.RequireRoles(models.DocumentCreateRoles...), handlers.CancelDocument)

// Generation lifecycle
docs.POST("/:id/generate",      middleware.RequireRoles(models.DocumentCreateRoles...), handlers.GenerateDocument)
docs.GET("/:id/download",       middleware.RequireRoles(models.DocumentViewRoles...), handlers.DownloadDocument)
docs.GET("/:id/status",         middleware.RequireRoles(models.DocumentViewRoles...), handlers.GetDocumentStatus)

// Attestations
docs.GET("/:id/attestations",   middleware.RequireRoles(models.DocumentViewRoles...), handlers.ListAttestations)
docs.PUT("/:id/attestations/:role", middleware.RequireRoles(models.DocumentCreateRoles...), handlers.UpsertAttestation)

// Approval workflow
docs.GET("/:id/approvals",      middleware.RequireRoles(models.DocumentViewRoles...), handlers.ListApprovals)
docs.POST("/:id/approvals",     middleware.RequireRoles(models.DocumentCreateRoles...), handlers.AddApprover)
docs.PUT("/:id/approvals/:approver_id/respond", handlers.RespondToApproval) // approver check in handler

// Requirement snapshots (read-only, for audit trail)
docs.GET("/:id/snapshots",      middleware.RequireRoles(models.DocumentViewRoles...), handlers.ListRequirementSnapshots)

// Finalization
docs.POST("/:id/finalize",      middleware.RequireRoles(models.DocumentFinalizeRoles...), handlers.FinalizeDocument)
```

**RBAC roles:**
- `DocumentViewRoles`: `auditor`, `compliance_manager`, `admin`, `owner`
- `DocumentCreateRoles`: `compliance_manager`, `admin`, `owner`
- `DocumentFinalizeRoles`: `admin`, `owner` (only principals can mark a document final)

---

## QSA Signature and Attestation Fields

The `document_attestations` table captures structured signatory data. For Sprint 11, the flow is:

1. Merchant configures `merchant_signatory` attestation (name, title, company)
2. QSA details are entered via `qsa_signatory` attestation (name, QSA number, company)
3. When `signature_method = 'manual'`, the PDF is generated with blank signature fields — the document is printed and physically signed
4. When `signature_method = 'docusign_ref'`, a DocuSign envelope ID is stored in `signature_ref` and can be verified externally

Electronic signature integration (DocuSign or PandaDoc native workflow) is deferred to Sprint 12. Sprint 11 delivers manual-signature support, which covers the majority of current use cases.

---

## Alternatives Considered

### Option A: Third-Party Document Generation Service (e.g., Carbone, DocRaptor)

**Pros:** No chromedp dependency, SaaS handles rendering complexity.
**Cons:** CHD-adjacent data (compliance status, org details, control posture) must be transmitted to a third party, creating a PCI DSS scope concern. A document generator for a PCI compliance tool cannot itself be a PCI scope risk. Rejected on this basis.

### Option B: Pre-built PCI SSC PDF Forms with Field Population

The PCI SSC publishes PDF forms. These could theoretically be populated via `pdf` field-filling libraries.
**Pros:** Guaranteed format compliance.
**Cons:** PCI SSC PDF forms use proprietary form field structures that are difficult to automate reliably. Text overflow, pagination, and custom sections are uncontrollable. Template maintenance (when PCI SSC updates forms) is painful. Rejected.

### Option C: Client-Side PDF Generation (React PDF / jsPDF)

Generate the document in the browser using React PDF or similar.
**Pros:** No server-side PDF dependency.
**Cons:** Compliance documents must be generated server-side for integrity — the server-generated artifact is the authoritative record. Client-side generation cannot be verified as unmodified. Rejected for compliance integrity reasons.

### Option D: Separate Document Generation Microservice

Extract PDF generation into a dedicated service.
**Pros:** Isolated failure domain, independently scalable.
**Cons:** Premature decomposition. PCI DSS compliance document generation frequency is low (once or twice per year per org). A microservice would add operational overhead (new container, service mesh config, health checks) for a low-frequency operation. Rejected as over-engineering.

---

## Consequences

**Positive:**
- Produces PCI SSC-mandated document formats — closes the critical Gap 4 identified in the architecture review
- Server-side generation ensures document integrity and creates an auditable artifact in MinIO
- The approval workflow gives organizations an internal control over document release
- Template system supports all SAQ variants — covers merchant tier diversity without code changes per variant
- Point-in-time snapshots ensure documents are accurate to the assessment period even as live data changes

**Negative / Trade-offs:**
- Chromium in the container increases Docker image size by ~250MB; mitigated by using a separate builder stage or a slim chromium-only layer
- `chromedp` is a CGO dependency — cross-compilation for non-Linux targets requires care (deploy target is Linux, so not a practical concern)
- Document generation is synchronous to the goroutine but async to the HTTP caller — the `generating` status requires polling or a future webhook/SSE notification
- PCI SSC updates AOC/SAQ forms periodically; templates will require maintenance with each new version

**Implementation order:**
1. Migrations 056-058 (Dana)
2. `DocumentAggregator` service (Logan — data layer)
3. Go templates for SAQ A/D AOC first (Logan — template engine)
4. chromedp integration + MinIO storage (Logan)
5. Approval workflow endpoints (Logan)
6. Frontend document creation and download UI (Alex)
7. ROC template sections (Logan — Sprint 12)

---

## Dependency on ADR-001 (CDE Scoping Module)

The AOC/ROC generator directly consumes data from the CDE Scoping Module:
- CDE scope summary is embedded in all AOC documents (merchant's scoping statement)
- ROC sections for Req 1 and 11.4 reference CDE asset inventory and segmentation test results

ADR-001 must be implemented first (or in parallel with a stub). The `DocumentAggregator` can tolerate missing CDE data with a graceful `"CDE inventory not yet configured"` warning in the document, rather than blocking generation entirely.

---

## References

- PCI DSS v4.0.1 Requirements 12.4.1, 12.4.2
- PCI SSC AOC templates: [pcistandards.org](https://www.pcisecuritystandards.org/document_library/)
- chromedp library: [github.com/chromedp/chromedp](https://github.com/chromedp/chromedp)
- Existing MinIO service: `api/internal/services/minio.go`
- Existing monitoring worker pattern: `api/internal/workers/`
- CDE Scoping Module ADR: `docs/adrs/ADR-001-cde-scoping-module.md`

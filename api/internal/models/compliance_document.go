package models

import "time"

// --- Document template type constants ---

const (
	DocTypeAOCSAQA   = "aoc_saq_a"
	DocTypeAOCSAQAEP = "aoc_saq_a_ep"
	DocTypeAOCSAQB   = "aoc_saq_b"
	DocTypeAOCSAQBIP = "aoc_saq_b_ip"
	DocTypeAOCSAQCVT = "aoc_saq_c_vt"
	DocTypeAOCSAQC   = "aoc_saq_c"
	DocTypeAOCSAQD   = "aoc_saq_d"
	DocTypeAOCSAQDSP = "aoc_saq_d_sp"
	DocTypeROC       = "roc"
)

var ValidDocumentTypes = []string{
	DocTypeAOCSAQA, DocTypeAOCSAQAEP, DocTypeAOCSAQB, DocTypeAOCSAQBIP,
	DocTypeAOCSAQCVT, DocTypeAOCSAQC, DocTypeAOCSAQD, DocTypeAOCSAQDSP,
	DocTypeROC,
}

func IsValidDocumentType(t string) bool {
	for _, v := range ValidDocumentTypes {
		if v == t {
			return true
		}
	}
	return false
}

// --- Document status constants ---

const (
	DocStatusDraft      = "draft"
	DocStatusGenerating = "generating"
	DocStatusReview     = "review"
	DocStatusApproved   = "approved"
	DocStatusFinal      = "final"
	DocStatusSigned     = "signed"
	DocStatusSuperseded = "superseded"
	DocStatusCancelled  = "cancelled"
)

var ValidDocumentStatuses = []string{
	DocStatusDraft, DocStatusGenerating, DocStatusReview,
	DocStatusApproved, DocStatusFinal, DocStatusSigned,
	DocStatusSuperseded, DocStatusCancelled,
}

// ValidDocStatusTransitions defines allowed document status transitions.
var ValidDocStatusTransitions = map[string][]string{
	DocStatusDraft:      {DocStatusGenerating, DocStatusCancelled},
	DocStatusGenerating: {DocStatusReview, DocStatusDraft},          // draft on failure
	DocStatusReview:     {DocStatusApproved, DocStatusDraft},        // draft on rejection
	DocStatusApproved:   {DocStatusFinal, DocStatusSigned, DocStatusDraft},
	DocStatusFinal:      {},                                         // terminal
	DocStatusSigned:     {},                                         // terminal
	DocStatusCancelled:  {},                                         // terminal
}

func IsValidDocStatusTransition(from, to string) bool {
	allowed, ok := ValidDocStatusTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

// RBAC roles for document management.
var (
	DocumentViewRoles     = []string{RoleAuditor, RoleComplianceManager, RoleCISO}
	DocumentCreateRoles   = []string{RoleComplianceManager, RoleCISO}
	DocumentFinalizeRoles = []string{RoleCISO}
)

// --- Template type constants (for document_templates table) ---

const (
	TemplateTypeAOC    = "aoc"
	TemplateTypeROC    = "roc"
	TemplateTypeSAQA   = "saq_a"
	TemplateTypeSAQAEP = "saq_a_ep"
	TemplateTypeSAQB   = "saq_b"
	TemplateTypeSAQBIP = "saq_b_ip"
	TemplateTypeSAQCVT = "saq_c_vt"
	TemplateTypeSAQC   = "saq_c"
	TemplateTypeSAQD   = "saq_d"
	TemplateTypeSAQDSP = "saq_d_sp"
)

// --- DocumentTemplate ---

// DocumentTemplate represents a PCI DSS AOC/ROC document template definition.
type DocumentTemplate struct {
	ID             string    `json:"id"`
	OrgID          *string   `json:"org_id"` // nil = system/global template
	TemplateType   string    `json:"template_type"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	PCIDSSVersion  string    `json:"pci_dss_version"`
	Version        string    `json:"version"`
	IsActive       bool      `json:"is_active"`
	Sections       string    `json:"sections"` // raw JSONB
	CreatedBy      *string   `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// --- ComplianceDocument (task AC: "generated_documents") ---

// ComplianceDocument represents a generated AOC/ROC compliance document.
type ComplianceDocument struct {
	ID                     string     `json:"id"`
	OrgID                  string     `json:"org_id"`
	TemplateID             *string    `json:"template_id"`
	DocumentType           string     `json:"document_type"`
	Title                  string     `json:"title"`
	AssessmentPeriodStart  time.Time  `json:"assessment_period_start"`
	AssessmentPeriodEnd    time.Time  `json:"assessment_period_end"`
	PCIDSSVersion          string     `json:"pci_dss_version"`
	MerchantName           *string    `json:"merchant_name"`
	MerchantDBA            *string    `json:"merchant_dba"`
	MerchantURL            *string    `json:"merchant_url"`
	BusinessType           *string    `json:"business_type"`
	QSAName                *string    `json:"qsa_name"`
	QSACompany             *string    `json:"qsa_company"`
	QSASignatureDate       *time.Time `json:"qsa_signature_date"`
	DocStatus              string     `json:"doc_status"`
	GeneratedBy            *string    `json:"generated_by"`
	DataSnapshot           *string    `json:"data_snapshot"` // raw JSONB
	PDFPath                *string    `json:"pdf_path"`
	FileSizeBytes          *int64     `json:"file_size_bytes"`
	GeneratedAt            *time.Time `json:"generated_at"`
	GenerationError        *string    `json:"generation_error,omitempty"`
	Version                int        `json:"version"`
	ParentID               *string    `json:"parent_id"`
	CreatedBy              *string    `json:"created_by"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// CreateComplianceDocumentRequest is the request body for POST /documents.
type CreateComplianceDocumentRequest struct {
	TemplateID            *string `json:"template_id"`
	DocumentType          string  `json:"document_type"  binding:"required"`
	Title                 string  `json:"title"          binding:"required"`
	AssessmentPeriodStart string  `json:"assessment_period_start" binding:"required"` // RFC3339 date
	AssessmentPeriodEnd   string  `json:"assessment_period_end"   binding:"required"` // RFC3339 date
	PCIDSSVersion         *string `json:"pci_dss_version"`
	MerchantName          *string `json:"merchant_name"`
	MerchantDBA           *string `json:"merchant_dba"`
	MerchantURL           *string `json:"merchant_url"`
	BusinessType          *string `json:"business_type"`
	QSAName               *string `json:"qsa_name"`
	QSACompany            *string `json:"qsa_company"`
	QSASignatureDate      *string `json:"qsa_signature_date"` // RFC3339 date
}

// UpdateComplianceDocumentRequest is the request body for PUT /documents/:id.
type UpdateComplianceDocumentRequest struct {
	Title                 *string `json:"title"`
	MerchantName          *string `json:"merchant_name"`
	MerchantDBA           *string `json:"merchant_dba"`
	MerchantURL           *string `json:"merchant_url"`
	BusinessType          *string `json:"business_type"`
	QSAName               *string `json:"qsa_name"`
	QSACompany            *string `json:"qsa_company"`
	QSASignatureDate      *string `json:"qsa_signature_date"`
}

// --- DocumentSection ---

// DocumentSection represents a content section within a compliance document.
type DocumentSection struct {
	ID               string    `json:"id"`
	DocumentID       string    `json:"document_id"`
	OrgID            string    `json:"org_id"`
	SectionKey       string    `json:"section_key"`
	Title            string    `json:"title"`
	Content          *string   `json:"content"`
	ComplianceStatus *string   `json:"compliance_status"`
	EvidenceIDs      []string  `json:"evidence_ids"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UpsertDocumentSectionRequest is the request body for PUT /documents/:id/sections/:key.
type UpsertDocumentSectionRequest struct {
	Title            string   `json:"title"       binding:"required"`
	Content          *string  `json:"content"`
	ComplianceStatus *string  `json:"compliance_status"`
	EvidenceIDs      []string `json:"evidence_ids"`
	SortOrder        *int     `json:"sort_order"`
}

// --- DocumentAttestation ---

// AttestationRole constants.
const (
	AttestationRoleMerchant = "merchant_signatory"
	AttestationRoleQSA      = "qsa_signatory"
	AttestationRoleISAC     = "isac_signatory"
	AttestationRoleSP       = "sp_signatory"
)

var ValidAttestationRoles = []string{
	AttestationRoleMerchant, AttestationRoleQSA,
	AttestationRoleISAC, AttestationRoleSP,
}

// DocumentAttestation represents a signatory's attestation for a compliance document.
type DocumentAttestation struct {
	ID              string     `json:"id"`
	DocumentID      string     `json:"document_id"`
	OrgID           string     `json:"org_id"`
	AttestationRole string     `json:"attestation_role"`
	FullName        string     `json:"full_name"`
	Title           string     `json:"title"`
	CompanyName     *string    `json:"company_name"`
	CompanyAddress  *string    `json:"company_address"`
	CompanyURL      *string    `json:"company_url"`
	Email           *string    `json:"email"`
	Phone           *string    `json:"phone"`
	QSACompany      *string    `json:"qsa_company"`
	QSANumber       *string    `json:"qsa_number"`
	SignedAt        *time.Time `json:"signed_at"`
	SignatureMethod *string    `json:"signature_method"`
	SignatureRef    *string    `json:"signature_ref"`
	CreatedBy       *string    `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UpsertAttestationRequest is the request body for PUT /documents/:id/attestations/:role.
type UpsertAttestationRequest struct {
	FullName        string  `json:"full_name"    binding:"required"`
	Title           string  `json:"title"        binding:"required"`
	CompanyName     *string `json:"company_name"`
	CompanyAddress  *string `json:"company_address"`
	CompanyURL      *string `json:"company_url"`
	Email           *string `json:"email"`
	Phone           *string `json:"phone"`
	QSACompany      *string `json:"qsa_company"`
	QSANumber       *string `json:"qsa_number"`
	SignedAt        *string `json:"signed_at"`        // RFC3339 datetime
	SignatureMethod *string `json:"signature_method"`
	SignatureRef    *string `json:"signature_ref"`
}

// --- DocumentApproval ---

// DocumentApproval represents an internal approval record for a compliance document.
type DocumentApproval struct {
	ID             string     `json:"id"`
	DocumentID     string     `json:"document_id"`
	OrgID          string     `json:"org_id"`
	ApproverID     string     `json:"approver_id"`
	ApprovalStatus string     `json:"approval_status"`
	Comments       *string    `json:"comments"`
	RespondedAt    *time.Time `json:"responded_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ApprovalResponseRequest is the request body for PUT /documents/:id/approvals/:approver_id/respond.
type ApprovalResponseRequest struct {
	Status   string  `json:"status"   binding:"required"` // approved | rejected
	Comments *string `json:"comments"`
}

// --- DocumentRequirementSnapshot ---

// RequirementSnapshotStatus constants (mirrors section compliance_status).
const (
	SnapStatusCompliant            = "compliant"
	SnapStatusNonCompliant         = "non_compliant"
	SnapStatusPartiallyCompliant   = "partially_compliant"
	SnapStatusNotApplicable        = "not_applicable"
	SnapStatusCompensatingControl  = "compensating_control"
	SnapStatusCustomizedApproach   = "customized_approach"
)

// DocumentRequirementSnapshot captures compliance posture per requirement at generation time.
type DocumentRequirementSnapshot struct {
	ID               string    `json:"id"`
	DocumentID       string    `json:"document_id"`
	RequirementID    *string   `json:"requirement_id"`
	RequirementCode  string    `json:"requirement_code"`
	RequirementTitle string    `json:"requirement_title"`
	InScope          bool      `json:"in_scope"`
	ControlCount     int       `json:"control_count"`
	PassingControls  int       `json:"passing_controls"`
	EvidenceCount    int       `json:"evidence_count"`
	Status           string    `json:"status"`
	Notes            *string   `json:"notes"`
	SnapshottedAt    time.Time `json:"snapshotted_at"`
}

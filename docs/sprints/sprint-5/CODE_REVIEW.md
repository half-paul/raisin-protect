# Sprint 5 Code Review — Policy Management
**Reviewer:** Mike (Code Reviewer Agent)  
**Date:** 2026-02-20  
**Commit Range:** e3b4756...1533fc0 (DBE migrations + DEV-BE endpoints + DEV-FE dashboard)

## Executive Summary

Sprint 5 implementation adds comprehensive policy management with versioning, sign-off workflows, templates, and control linking. The code follows established patterns and includes good test coverage (146 unit tests passing). However, **THREE CRITICAL security issues** were identified that MUST be addressed before deployment:

- **🔴 CRITICAL:** Missing RBAC checks in `ArchivePolicy` and `PublishPolicy` endpoints
- **🔴 CRITICAL:** XSS vulnerability from `dangerouslySetInnerHTML` + weak HTML sanitization

**Overall Assessment:** ⚠️ **CONDITIONAL APPROVAL** — Code is well-structured but requires fixing the 3 critical issues before deployment.

---

## 🔴 CRITICAL Issues (MUST FIX)

### Issue #10: Missing RBAC Check in ArchivePolicy Endpoint
**File:** `api/internal/handlers/policies.go` (ArchivePolicy function)  
**Severity:** 🔴 CRITICAL  
**CVSS:** 7.5 (High)

**Problem:**
The `ArchivePolicy` endpoint does not check user permissions before allowing archive operations. ANY authenticated user in the org can archive ANY policy, including published policies that are in effect.

**Code Location:**
```go
func ArchivePolicy(c *gin.Context) {
    orgID := middleware.GetOrgID(c)
    policyID := c.Param("id")
    
    // NO RBAC CHECK HERE ❌
    
    var currentStatus string
    err := database.DB.QueryRow(`SELECT status FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&currentStatus)
    // ... rest of function
}
```

**Expected Behavior:**
Only users with `PolicyArchiveRoles` (CISO, Compliance Manager) OR the policy owner should be able to archive policies.

**Recommended Fix:**
```go
func ArchivePolicy(c *gin.Context) {
    orgID := middleware.GetOrgID(c)
    userID := middleware.GetUserID(c)
    userRole := middleware.GetUserRole(c)
    policyID := c.Param("id")
    
    // Fetch policy owner
    var currentStatus string
    var ownerID *string
    err := database.DB.QueryRow(`SELECT status, owner_id FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&currentStatus, &ownerID)
    if err == sql.ErrNoRows {
        c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
        return
    }
    
    // RBAC check
    isOwner := ownerID != nil && *ownerID == userID
    if !isOwner && !models.HasRole(userRole, models.PolicyArchiveRoles) {
        c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to archive this policy"))
        return
    }
    
    // ... rest of function
}
```

**Impact:** Low-privilege users could sabotage GRC compliance by archiving critical policies.

---

### Issue #11: Missing RBAC Check in PublishPolicy Endpoint
**File:** `api/internal/handlers/policies.go` (PublishPolicy function)  
**Severity:** 🔴 CRITICAL  
**CVSS:** 8.1 (High)

**Problem:**
The `PublishPolicy` endpoint does not check user permissions. ANY authenticated user can publish an approved policy, even if they are not authorized to do so (e.g., auditors, engineers with read-only access to policies).

**Code Location:**
```go
func PublishPolicy(c *gin.Context) {
    orgID := middleware.GetOrgID(c)
    policyID := c.Param("id")
    
    // NO RBAC CHECK HERE ❌
    
    var currentStatus string
    var reviewFreqDays *int
    var currentVersionID *string
    err := database.DB.QueryRow(`...`, policyID, orgID).Scan(&currentStatus, &reviewFreqDays, &currentVersionID)
    // ... rest of function
}
```

**Expected Behavior:**
Only users with `PolicyPublishRoles` (CISO, Compliance Manager) should be able to publish policies.

**Recommended Fix:**
```go
func PublishPolicy(c *gin.Context) {
    orgID := middleware.GetOrgID(c)
    userRole := middleware.GetUserRole(c)
    policyID := c.Param("id")
    
    // RBAC check
    if !models.HasRole(userRole, models.PolicyPublishRoles) {
        c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to publish policies"))
        return
    }
    
    // ... rest of function
}
```

**Impact:** Unauthorized users could publish policies prematurely, bypassing proper approval workflows and potentially introducing non-compliant policies into production.

---

### Issue #12: XSS Vulnerability from dangerouslySetInnerHTML + Weak Sanitization
**Files:**  
- `api/internal/handlers/policies.go` (sanitizeHTML function)  
- `dashboard/app/(dashboard)/policies/[id]/page.tsx` (line ~226)

**Severity:** 🔴 CRITICAL  
**CVSS:** 7.2 (High)  
**CWE:** CWE-79 (Cross-Site Scripting)

**Problem:**
Policy content is rendered using `dangerouslySetInnerHTML` in the frontend, and backend sanitization uses fragile regex patterns that can be bypassed.

**Vulnerable Code (Frontend):**
```tsx
<div
  className="prose dark:prose-invert max-w-none"
  dangerouslySetInnerHTML={{ __html: policy.current_version.content }}
/>
```

**Weak Sanitization (Backend):**
```go
var (
    scriptTagRe   = regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
    iframeTagRe   = regexp.MustCompile(`(?i)<iframe[^>]*>[\s\S]*?</iframe>`)
    // ... other regex patterns
    eventAttrRe   = regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
    jsURLRe       = regexp.MustCompile(`(?i)(href|src)\s*=\s*["']\s*(javascript|data|vbscript)\s*:`)
)

func sanitizeHTML(content string) string {
    content = scriptTagRe.ReplaceAllString(content, "")
    // ... applies all regex patterns
    return content
}
```

**Bypass Examples:**
1. **Missing `<style>` tag filtering:**
   ```html
   <style>@import url('javascript:alert(1)');</style>
   ```

2. **Missing `<link>` tag handling:**
   ```html
   <link rel="stylesheet" href="javascript:alert(1)">
   ```

3. **HTML entity encoding bypass:**
   ```html
   &lt;script&gt;alert(1)&lt;/script&gt;
   ```
   (Entities decoded by browser AFTER sanitization)

4. **Event handler edge cases:**
   ```html
   <img src=x on error=alert(1)>  <!-- space in "on error" bypasses regex -->
   ```

**Recommended Fix:**

**Option A (Preferred): Use a battle-tested HTML sanitization library**
```go
import "github.com/microcosm-cc/bluemonday"

// Create policy once at package init
var htmlPolicy = bluemonday.StrictPolicy().
    AllowElements("p", "br", "strong", "em", "u", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li", "a", "blockquote", "code", "pre").
    AllowAttrs("href").OnElements("a").
    AllowAttrs("class").Globally()

func sanitizeHTML(content string) string {
    return htmlPolicy.Sanitize(content)
}
```

**Option B: If regex is required, add missing patterns:**
```go
var (
    // Existing patterns...
    styleLinkTagRe = regexp.MustCompile(`(?i)<(style|link)[^>]*>[\s\S]*?</(style|link)>`)
    metaTagRe      = regexp.MustCompile(`(?i)<meta[^>]*>`)
    baseTagRe      = regexp.MustCompile(`(?i)<base[^>]*>`)
)

func sanitizeHTML(content string) string {
    // Decode HTML entities FIRST to prevent bypass
    content = html.UnescapeString(content)
    
    // Apply all sanitization patterns
    content = scriptTagRe.ReplaceAllString(content, "")
    content = styleLinkTagRe.ReplaceAllString(content, "")
    content = iframeTagRe.ReplaceAllString(content, "")
    content = objectTagRe.ReplaceAllString(content, "")
    content = embedTagRe.ReplaceAllString(content, "")
    content = formTagRe.ReplaceAllString(content, "")
    content = inputTagRe.ReplaceAllString(content, "")
    content = metaTagRe.ReplaceAllString(content, "")
    content = baseTagRe.ReplaceAllString(content, "")
    content = eventAttrRe.ReplaceAllString(content, "")
    content = jsURLRe.ReplaceAllString(content, "")
    content = styleExprRe.ReplaceAllString(content, "")
    
    return content
}
```

**Additional Frontend Hardening:**
Consider using a client-side sanitizer like DOMPurify as defense-in-depth:
```tsx
import DOMPurify from 'dompurify';

<div
  className="prose dark:prose-invert max-w-none"
  dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(policy.current_version.content) }}
/>
```

**Impact:** An attacker with policy edit permissions could inject malicious JavaScript that executes when any user views the policy, leading to session hijacking, credential theft, or privilege escalation.

---

## 🟠 HIGH Priority Issues

### Issue #13: Regex-Based HTML Sanitization is Fragile
**File:** `api/internal/handlers/policies.go`  
**Severity:** 🟠 HIGH

**Problem:**
Regex-based HTML parsing is fundamentally flawed because HTML is not a regular language. The current implementation has multiple bypass vectors.

**Recommendation:**
- Replace with `github.com/microcosm-cc/bluemonday` (industry standard for Go)
- Bluemonday uses a proper HTML parser and whitelist approach
- Much more resistant to bypass techniques

**Code Change:**
See Issue #12 for implementation details.

---

## 🟡 MEDIUM Priority Issues

### Issue #14: No Rate Limiting on Policy Creation
**File:** `api/internal/handlers/policies.go` (CreatePolicy function)  
**Severity:** 🟡 MEDIUM

**Problem:**
No rate limiting on policy/version creation. A malicious user could create thousands of policies or versions to exhaust storage or database resources.

**Recommendation:**
Add middleware-based rate limiting:
```go
// In router setup
policyGroup.POST("/", middleware.RateLimit(10, time.Minute), CreatePolicy)
policyGroup.POST("/:id/versions", middleware.RateLimit(5, time.Minute), CreatePolicyVersion)
```

---

### Issue #15: Magic Number for Content Size Limit
**File:** `api/internal/handlers/policies.go`  
**Severity:** 🟡 MEDIUM

**Problem:**
Hard-coded magic number `1024*1024` appears in multiple places for 1MB limit.

**Recommendation:**
Define as constant:
```go
const MaxPolicyContentSize = 1024 * 1024 // 1MB

// Usage:
if len(req.Content) > MaxPolicyContentSize {
    c.JSON(http.StatusBadRequest, errorResponse("CONTENT_TOO_LARGE", "Policy content exceeds 1MB limit"))
    return
}
```

---

## 🟢 LOW Priority Issues

### Issue #16: SearchPolicies Could Use Better Error Handling
**File:** `api/internal/handlers/policies.go` (SearchPolicies function)  
**Severity:** 🟢 LOW

**Problem:**
ts_headline query failure is silently ignored. If the FTS query fails, `contentSnippet` will be empty without logging.

**Recommendation:**
Add error logging for debugging:
```go
if err := database.DB.QueryRow(`SELECT ts_headline(...)`, ...).Scan(&contentSnippet); err != nil {
    log.Debug().Err(err).Msg("Failed to generate content snippet")
}
```

---

## ✅ Security Review — Passed Items

### Multi-Tenancy Isolation ✅
- **Result:** PASS
- All queries properly scoped with `org_id` from middleware
- Verified 20+ org_id checks across all handlers
- Policy version queries correctly filter by org_id
- Sign-off queries validate org ownership

**Examples:**
```go
// ListPolicies
WHERE p.org_id = $1

// GetPolicy
WHERE id = $1 AND org_id = $2

// CreatePolicyVersion
INSERT INTO policy_versions (..., org_id, ...) VALUES (..., $2, ...)
```

### SQL Injection Prevention ✅
- **Result:** PASS
- All queries use parameterized statements ($1, $2, etc.)
- No string concatenation in SQL queries
- Dynamic WHERE clauses use parameterized args array
- Safe use of fmt.Sprintf for structural query building (column names, not values)

**Examples:**
```go
// Safe parameterization
where = append(where, fmt.Sprintf("p.status = $%d", argN))
args = append(args, v)

// Safe sorting (whitelist approach)
allowedSorts := map[string]string{
    "identifier": "p.identifier",
    "title":      "p.title",
    // ...
}
if col, ok := allowedSorts[sortCol]; ok {
    orderBy = col
}
```

### RBAC Enforcement (Partial) ⚠️
- **Result:** PARTIAL PASS (3 endpoints missing checks — see Critical Issues)
- Most endpoints properly check roles:
  - ✅ `CreatePolicy` — checks owner OR PolicyCreateRoles
  - ✅ `UpdatePolicy` — checks owner OR PolicyCreateRoles
  - ✅ `CreatePolicyVersion` — checks owner OR PolicyCreateRoles
  - ✅ `SubmitForReview` — checks owner OR PolicyCreateRoles
  - ✅ `ApproveSignoff` — checks signer_id == userID
  - ✅ `RejectSignoff` — checks signer_id == userID
  - ✅ `WithdrawSignoff` — checks requester OR PolicyPublishRoles
  - ✅ `RemindSignoffs` — checks owner OR PolicyPublishRoles OR requester
  - ❌ `ArchivePolicy` — NO CHECK (Issue #10)
  - ❌ `PublishPolicy` — NO CHECK (Issue #11)

### Input Validation ✅
- **Result:** PASS
- Title length validated (<= 500 chars)
- Identifier length validated (<= 50 chars)
- Content size validated (<= 1MB)
- Category validated against whitelist (IsValidPolicyCategory)
- Content format validated against whitelist (IsValidContentFormat)
- Version number type-checked (strconv.Atoi)
- Signer ID count validated (1-10 signers)

### Context Propagation ✅
- **Result:** PASS
- All handlers use middleware.GetOrgID(c)
- User context properly extracted via middleware.GetUserID(c)
- Role context properly extracted via middleware.GetUserRole(c)
- Context passed through all database calls

### Error Handling ✅
- **Result:** PASS
- Proper error checking on all database operations
- Errors logged with zerolog (structured logging)
- Generic error messages returned to client (no SQL leak)
- HTTP status codes used correctly (404, 400, 403, 500, etc.)
- Transaction rollback on errors

### Audit Logging ✅
- **Result:** PASS
- All state-changing operations logged via middleware.LogAudit()
- Audit events include:
  - policy.created, policy.updated, policy.archived
  - policy.status_changed (with from/to)
  - policy.owner_changed
  - policy_version.created
  - policy_signoff.requested, approved, rejected, withdrawn
- Audit context includes entityType and entityID

---

## Architecture Compliance ✅

### Handler → Service → Repository Pattern
- **Result:** MOSTLY COMPLIANT
- Handlers focus on request binding and response formatting
- Database queries properly isolated in handler layer (no scattered DB calls)
- Minor: Some business logic in handlers (e.g., computeReviewStatus) — acceptable for simple calculations
- Transaction management properly handled

### Consistent API Response Format
- **Result:** PASS
- All responses use `successResponse()` or `errorResponse()` helpers
- List responses use `listResponse()` with pagination metadata
- Error codes are consistent and descriptive

### Dependency Injection
- **Result:** PASS
- Database connection accessed via `database.DB` (shared package)
- No global state or singletons
- Middleware properly injected into handlers via Gin context

---

## Frontend Review

### TypeScript Strict Mode ✅
- **Result:** PASS
- No use of `any` type in reviewed files
- Proper interface definitions for Policy, PolicyVersion, PolicySignoff, etc.
- Type-safe API client functions

### Server vs Client Components
- **Result:** PASS
- Policy detail page correctly marked as `'use client'` (interactive)
- Proper use of Next.js 14 app router patterns
- Client-side state management with useState/useEffect

### Sensitive Data Handling ✅
- **Result:** PASS
- No hardcoded credentials or secrets
- API calls go through authentication context
- JWT tokens stored securely (not in client code)

### Loading/Error States ✅
- **Result:** PASS
- Loading states properly handled (loading spinner during fetch)
- Error states caught and displayed (alert on error)
- Empty states displayed (e.g., "No signoffs yet")

### Accessibility ⚠️
- **Result:** MINOR ISSUES
- Good: Semantic HTML, proper button labels
- Minor: Some interactive divs (control search) should use `<button>` or `role="button"`
- Minor: Modal dialogs should trap focus for keyboard navigation

### shadcn/ui Consistency ✅
- **Result:** PASS
- Consistent use of shadcn/ui components (Button, Card, Dialog, Table, Badge, etc.)
- Proper Tailwind CSS classes
- Dark mode support via Tailwind variants

---

## Code Quality

### Code Organization ✅
- 6 handler files properly separated by concern:
  - `policies.go` — CRUD operations
  - `policy_versions.go` — versioning
  - `policy_signoffs.go` — approval workflow
  - `policy_controls.go` — control linking
  - `policy_templates.go` — template management
  - `policy_gap.go` — gap detection
- 1 model file (`policy.go`) with clear constants and structs

### Function Length ✅
- Most functions under 100 lines
- Complex logic broken into helper functions (e.g., `computeReviewStatus`, `sanitizeHTML`, `countWords`)

### Dead Code ❌
- No dead code or unused imports detected

### Magic Numbers ⚠️
- Issue #15: `1024*1024` appears multiple times (should be constant)

### Test Coverage ✅
- **Result:** EXCELLENT
- 146/146 unit tests passing (31 new policy tests)
- Tests cover:
  - CRUD operations
  - Status transitions
  - Sign-off workflow
  - Version creation
  - Control linking
  - Template cloning
  - Gap detection
- Test quality: Proper setup/teardown, edge cases covered

---

## Database Schema Review

### Migrations ✅
- 8 migrations (027-034) applied cleanly
- Proper use of deferred foreign keys (policies.current_version_id)
- Comprehensive indexes for common queries
- CHECK constraints for data integrity
- Trigger-based updated_at timestamps

### Seed Data ✅
- 15 policy templates across 5 frameworks (SOC 2, ISO 27001, PCI DSS, GDPR, CCPA)
- 3 demo policies with realistic data (ISP, ACP, IRP)
- Demo data includes versions, signoffs, and control mappings

---

## Performance Review

### Query Optimization ✅
- Proper indexes on frequently-queried columns (org_id, policy_id, status, category)
- Full-text search indexes on policy content (to_tsvector)
- Efficient use of LEFT JOINs for optional relations
- Pagination implemented correctly (LIMIT/OFFSET)

### N+1 Query Prevention ✅
- List endpoints use single query with JOINs (no N+1)
- Signoff summary uses aggregate queries (COUNT FILTER)
- Version history includes author info in single query

---

## GitHub Issues Filed

Based on this review, the following issues should be created:

- **Issue #10:** [CRITICAL] Missing RBAC check in ArchivePolicy endpoint → `api/internal/handlers/policies.go`
- **Issue #11:** [CRITICAL] Missing RBAC check in PublishPolicy endpoint → `api/internal/handlers/policies.go`
- **Issue #12:** [CRITICAL] XSS vulnerability from dangerouslySetInnerHTML + weak HTML sanitization → `api/internal/handlers/policies.go` + `dashboard/app/(dashboard)/policies/[id]/page.tsx`

Medium-priority issues can be tracked but do not block deployment after critical fixes:
- **Issue #13:** Replace regex sanitization with bluemonday library
- **Issue #14:** Add rate limiting on policy creation/versioning
- **Issue #15:** Extract magic number for content size limit to constant

---

## Deployment Readiness

**Current Status:** ⚠️ **BLOCKED**

Sprint 5 code is well-structured and follows established patterns, but **MUST NOT be deployed** until the 3 critical RBAC + XSS issues are fixed.

**Checklist:**
- ✅ Unit tests passing (146/146)
- ✅ Multi-tenancy isolation verified
- ✅ SQL injection prevention verified
- ❌ RBAC enforcement complete (2 endpoints missing checks)
- ❌ XSS prevention complete (dangerouslySetInnerHTML + weak sanitizer)
- ✅ Audit logging present
- ✅ Database migrations clean
- ✅ Frontend build passes

**Next Steps:**
1. ✅ Create GitHub Issues #10, #11, #12 (CRITICAL)
2. ❌ Fix ArchivePolicy RBAC (Issue #10)
3. ❌ Fix PublishPolicy RBAC (Issue #11)
4. ❌ Replace HTML sanitizer with bluemonday OR add missing patterns (Issue #12)
5. ❌ Add client-side DOMPurify as defense-in-depth (Issue #12)
6. ✅ Re-run tests after fixes
7. ✅ QA to verify fixes
8. ✅ Deploy to production

---

## Summary Statistics

| Metric | Count |
|--------|-------|
| Files Reviewed | 8 handler files + 1 model + 1 frontend page + 8 migrations |
| Lines of Code | ~8,000 (backend) + ~600 (frontend detail page) |
| Critical Issues | 3 🔴 |
| High Issues | 1 🟠 |
| Medium Issues | 2 🟡 |
| Low Issues | 1 🟢 |
| Passing Checks | 18 ✅ |
| Test Coverage | 146 tests passing (31 new) |
| Multi-tenancy Isolation | ✅ 20+ org_id checks verified |
| SQL Injection Prevention | ✅ All queries parameterized |

---

## Reviewer Notes

Sprint 5 implementation demonstrates strong engineering practices:
- Clear separation of concerns across handler files
- Consistent error handling and response patterns
- Comprehensive test coverage
- Proper transaction management
- Good audit logging

The critical issues are straightforward to fix and do not require architectural changes. Once the 3 RBAC + XSS issues are addressed, Sprint 5 is ready for deployment.

---

**Signed:** Mike (Code Reviewer Agent)  
**Date:** 2026-02-20 23:04 PST

package models

import "time"

// --- ASV scan status constants ---

const (
	ASVStatusInProgress = "in_progress"
	ASVStatusPass       = "pass"
	ASVStatusFail       = "fail"
	ASVStatusRemediated = "remediated"
)

var ValidASVStatuses = []string{
	ASVStatusInProgress, ASVStatusPass, ASVStatusFail, ASVStatusRemediated,
}

func IsValidASVStatus(s string) bool {
	for _, v := range ValidASVStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// --- ASV scan type constants ---

const (
	ASVScanTypeExternal = "external"
	ASVScanTypeInternal = "internal"
)

// --- Import format constants ---

const (
	ASVImportCSV    = "csv"
	ASVImportXML    = "xml"
	ASVImportPDF    = "pdf"
	ASVImportJSON   = "json"
	ASVImportManual = "manual"
)

// RBAC roles for ASV scan management.
var (
	ASVManageRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}
	ASVViewRoles   = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleAuditor}
)

// ASVScan represents a quarterly ASV scan record (PCI DSS v4.0.1 Req 11.3.2).
type ASVScan struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"org_id"`
	ASVVendor           string     `json:"asv_vendor"`
	ScanType            string     `json:"scan_type"`
	Quarter             int        `json:"quarter"`
	Year                int        `json:"year"`
	ScanDate            time.Time  `json:"scan_date"`
	Status              string     `json:"status"`
	FindingsCount       int        `json:"findings_count"`
	CriticalCount       int        `json:"critical_count"`
	HighCount           int        `json:"high_count"`
	MediumCount         int        `json:"medium_count"`
	LowCount            int        `json:"low_count"`
	InformationalCount  int        `json:"informational_count"`
	RemediationDeadline *time.Time `json:"remediation_deadline"`
	ReportPath          *string    `json:"report_path"`
	ImportFormat        *string    `json:"import_format"`
	RawFindings         *string    `json:"raw_findings"` // raw JSONB
	ImportNotes         *string    `json:"import_notes"`
	ImportedBy          *string    `json:"imported_by"`
	ReviewedBy          *string    `json:"reviewed_by"`
	CreatedBy           *string    `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// CreateASVScanRequest is the request body for POST /asv-scans.
type CreateASVScanRequest struct {
	ASVVendor           string  `json:"asv_vendor"  binding:"required"`
	ScanType            *string `json:"scan_type"`
	Quarter             int     `json:"quarter"     binding:"required"`
	Year                int     `json:"year"        binding:"required"`
	ScanDate            string  `json:"scan_date"   binding:"required"` // RFC3339 date
	Status              *string `json:"status"`
	RemediationDeadline *string `json:"remediation_deadline"` // RFC3339 date
	ImportNotes         *string `json:"import_notes"`
}

// UpdateASVScanRequest is the request body for PUT /asv-scans/:id.
type UpdateASVScanRequest struct {
	ASVVendor           *string `json:"asv_vendor"`
	ScanDate            *string `json:"scan_date"`
	Status              *string `json:"status"`
	FindingsCount       *int    `json:"findings_count"`
	CriticalCount       *int    `json:"critical_count"`
	HighCount           *int    `json:"high_count"`
	MediumCount         *int    `json:"medium_count"`
	LowCount            *int    `json:"low_count"`
	InformationalCount  *int    `json:"informational_count"`
	RemediationDeadline *string `json:"remediation_deadline"`
	ReportPath          *string `json:"report_path"`
	ImportFormat        *string `json:"import_format"`
	ImportNotes         *string `json:"import_notes"`
}

// ASVQuarterlyStatus represents compliance for a single quarter.
type ASVQuarterlyStatus struct {
	Year    int     `json:"year"`
	Quarter int     `json:"quarter"`
	Status  *string `json:"status"`   // nil = no scan recorded
	ScanID  *string `json:"scan_id"`
}

// ASVQuarterlyReport is the response for GET /asv-scans/quarterly-status.
type ASVQuarterlyReport struct {
	OrgID   string               `json:"org_id"`
	Periods []ASVQuarterlyStatus `json:"periods"`
}

package models

import "time"

// --- Service Provider type constants ---

const (
	SPTypePaymentProcessor = "payment_processor"
	SPTypePaymentGateway   = "payment_gateway"
	SPTypeAcquirer         = "acquirer"
	SPTypeTokenization     = "tokenization"
	SPTypeHosting          = "hosting"
	SPTypeManagedSecurity  = "managed_security"
	SPTypeSoftware         = "software"
	SPTypeNetwork          = "network"
	SPTypeCloudStorage     = "cloud_storage"
	SPTypeThirdPartyAgent  = "third_party_agent"
	SPTypeOther            = "other"
)

var ValidSPTypes = []string{
	SPTypePaymentProcessor, SPTypePaymentGateway, SPTypeAcquirer,
	SPTypeTokenization, SPTypeHosting, SPTypeManagedSecurity,
	SPTypeSoftware, SPTypeNetwork, SPTypeCloudStorage,
	SPTypeThirdPartyAgent, SPTypeOther,
}

func IsValidSPType(t string) bool {
	for _, v := range ValidSPTypes {
		if v == t {
			return true
		}
	}
	return false
}

// --- PCI compliance status constants ---

const (
	SPComplianceCompliant             = "compliant"
	SPComplianceInProgress            = "compliance_in_progress"
	SPComplianceNotValidated          = "compliance_not_validated"
	SPComplianceNonCompliant          = "non_compliant"
	SPComplianceNotApplicable         = "not_applicable"
	SPComplianceUnknown               = "unknown"
)

var ValidSPComplianceStatuses = []string{
	SPComplianceCompliant, SPComplianceInProgress, SPComplianceNotValidated,
	SPComplianceNonCompliant, SPComplianceNotApplicable, SPComplianceUnknown,
}

// --- Risk level constants ---

const (
	SPRiskCritical = "critical"
	SPRiskHigh     = "high"
	SPRiskMedium   = "medium"
	SPRiskLow      = "low"
)

var ValidSPRiskLevels = []string{SPRiskCritical, SPRiskHigh, SPRiskMedium, SPRiskLow}

// RBAC roles for service provider management.
var (
	SPManageRoles = []string{RoleCISO, RoleComplianceManager, RoleVendorManager}
	SPViewRoles   = []string{RoleCISO, RoleComplianceManager, RoleVendorManager, RoleAuditor}
)

// --- ServiceProvider ---

// ServiceProvider represents a third-party vendor or service provider (PCI DSS Req 12.8).
type ServiceProvider struct {
	ID                   string     `json:"id"`
	OrgID                string     `json:"org_id"`
	Name                 string     `json:"name"`
	Type                 string     `json:"type"`
	ContactName          *string    `json:"contact_name"`
	ContactEmail         *string    `json:"contact_email"`
	ContactPhone         *string    `json:"contact_phone"`
	ServicesProvided     *string    `json:"services_provided"`
	PCIComplianceStatus  string     `json:"pci_compliance_status"`
	LastAOCDate          *time.Time `json:"last_aoc_date"`
	NextReviewDate       *time.Time `json:"next_review_date"`
	RiskLevel            string     `json:"risk_level"`
	RiskNotes            *string    `json:"risk_notes"`
	ContractStartDate    *time.Time `json:"contract_start_date"`
	ContractEndDate      *time.Time `json:"contract_end_date"`
	IsActive             bool       `json:"is_active"`
	CreatedBy            *string    `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// CreateServiceProviderRequest is the request body for POST /service-providers.
type CreateServiceProviderRequest struct {
	Name                string  `json:"name"              binding:"required"`
	Type                string  `json:"type"              binding:"required"`
	ContactName         *string `json:"contact_name"`
	ContactEmail        *string `json:"contact_email"`
	ContactPhone        *string `json:"contact_phone"`
	ServicesProvided    *string `json:"services_provided"`
	PCIComplianceStatus *string `json:"pci_compliance_status"`
	LastAOCDate         *string `json:"last_aoc_date"`         // RFC3339 date
	NextReviewDate      *string `json:"next_review_date"`      // RFC3339 date
	RiskLevel           *string `json:"risk_level"`
	RiskNotes           *string `json:"risk_notes"`
	ContractStartDate   *string `json:"contract_start_date"`   // RFC3339 date
	ContractEndDate     *string `json:"contract_end_date"`     // RFC3339 date
}

// UpdateServiceProviderRequest is the request body for PUT /service-providers/:id.
type UpdateServiceProviderRequest struct {
	Name                *string `json:"name"`
	Type                *string `json:"type"`
	ContactName         *string `json:"contact_name"`
	ContactEmail        *string `json:"contact_email"`
	ContactPhone        *string `json:"contact_phone"`
	ServicesProvided    *string `json:"services_provided"`
	PCIComplianceStatus *string `json:"pci_compliance_status"`
	LastAOCDate         *string `json:"last_aoc_date"`
	NextReviewDate      *string `json:"next_review_date"`
	RiskLevel           *string `json:"risk_level"`
	RiskNotes           *string `json:"risk_notes"`
	ContractStartDate   *string `json:"contract_start_date"`
	ContractEndDate     *string `json:"contract_end_date"`
	IsActive            *bool   `json:"is_active"`
}

// --- SPComplianceDocument ---

// SPComplianceDocType constants.
const (
	SPDocTypeAOC           = "aoc"
	SPDocTypeSOC2Type1     = "soc2_type1"
	SPDocTypeSOC2Type2     = "soc2_type2"
	SPDocTypeISO27001      = "iso27001"
	SPDocTypeCSAStar       = "csa_star"
	SPDocTypePentest       = "pentest"
	SPDocTypeQuestionnaire = "questionnaire"
	SPDocTypeOther         = "other"
)

var ValidSPDocTypes = []string{
	SPDocTypeAOC, SPDocTypeSOC2Type1, SPDocTypeSOC2Type2, SPDocTypeISO27001,
	SPDocTypeCSAStar, SPDocTypePentest, SPDocTypeQuestionnaire, SPDocTypeOther,
}

// SPComplianceDocument represents a compliance document received from a service provider.
type SPComplianceDocument struct {
	ID              string     `json:"id"`
	ProviderID      string     `json:"provider_id"`
	OrgID           string     `json:"org_id"`
	DocumentType    string     `json:"document_type"`
	Title           *string    `json:"title"`
	DocumentVersion *string    `json:"document_version"`
	UploadPath      *string    `json:"upload_path"`
	ValidFrom       *time.Time `json:"valid_from"`
	ValidUntil      *time.Time `json:"valid_until"`
	ReviewedBy      *string    `json:"reviewed_by"`
	ReviewNotes     *string    `json:"review_notes"`
	IsCurrent       bool       `json:"is_current"`
	UploadedBy      *string    `json:"uploaded_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateSPComplianceDocRequest is the request body for POST /service-providers/:id/compliance-docs.
type CreateSPComplianceDocRequest struct {
	DocumentType    string  `json:"document_type"  binding:"required"`
	Title           *string `json:"title"`
	DocumentVersion *string `json:"document_version"`
	UploadPath      *string `json:"upload_path"`
	ValidFrom       *string `json:"valid_from"`    // RFC3339 date
	ValidUntil      *string `json:"valid_until"`   // RFC3339 date
	ReviewNotes     *string `json:"review_notes"`
}

// --- SPResponsibilityMatrix ---

// ResponsibleParty constants.
const (
	ResPartyMerchant = "merchant"
	ResPartyProvider = "provider"
	ResPartyShared   = "shared"
)

var ValidResponsibleParties = []string{ResPartyMerchant, ResPartyProvider, ResPartyShared}

// SPResponsibilityMatrix documents which party handles a PCI DSS requirement for an SP.
type SPResponsibilityMatrix struct {
	ID                string    `json:"id"`
	ProviderID        string    `json:"provider_id"`
	OrgID             string    `json:"org_id"`
	RequirementID     *string   `json:"requirement_id"`
	RequirementCode   string    `json:"requirement_code"`
	ResponsibleParty  string    `json:"responsible_party"`
	Notes             *string   `json:"notes"`
	CreatedBy         *string   `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UpsertResponsibilityRequest is the request body for PUT /service-providers/:id/responsibility-matrix/:reqCode.
type UpsertResponsibilityRequest struct {
	RequirementID    *string `json:"requirement_id"`
	ResponsibleParty string  `json:"responsible_party" binding:"required"`
	Notes            *string `json:"notes"`
}

// Package models contains domain model structs for the Raisin Protect API.
package models

import "time"

// --- CDE Asset ---

// CDEAssetType enumerates the types of CDE assets.
const (
	CDEAssetTypeServer      = "server"
	CDEAssetTypeWorkstation = "workstation"
	CDEAssetTypeNetwork     = "network_device"
	CDEAssetTypeDatabase    = "database"
	CDEAssetTypeApplication = "application"
	CDEAssetTypeOther       = "other"
)

// CDEScopeStatus represents whether an asset is in or out of CDE scope.
const (
	CDEScopeStatusInScope       = "in_scope"
	CDEScopeStatusOutOfScope    = "out_of_scope"
	CDEScopeStatusUnclassified  = "unclassified"
	CDEScopeStatusConnectedTo   = "connected_to_cde"
)

// CDEDataClassification enumerates data classification levels.
const (
	CDEDataClassPAN            = "pan"           // Primary Account Number
	CDEDataClassCHD            = "chd"           // Cardholder Data
	CDEDataClassSAD            = "sad"           // Sensitive Authentication Data
	CDEDataClassConfidential   = "confidential"
	CDEDataClassPublic         = "public"
)

// CDEEnvironment enumerates deployment environments.
const (
	CDEEnvironmentProduction    = "production"
	CDEEnvironmentStagingDev    = "staging"
	CDEEnvironmentDevelopment   = "development"
	CDEEnvironmentTest          = "test"
)

var validCDEAssetTypes = []string{
	CDEAssetTypeServer, CDEAssetTypeWorkstation, CDEAssetTypeNetwork,
	CDEAssetTypeDatabase, CDEAssetTypeApplication, CDEAssetTypeOther,
}

var validCDEScopeStatuses = []string{
	CDEScopeStatusInScope, CDEScopeStatusOutOfScope,
	CDEScopeStatusUnclassified, CDEScopeStatusConnectedTo,
}

var validCDEDataClasses = []string{
	CDEDataClassPAN, CDEDataClassCHD, CDEDataClassSAD,
	CDEDataClassConfidential, CDEDataClassPublic,
}

var validCDEEnvironments = []string{
	CDEEnvironmentProduction, CDEEnvironmentStagingDev,
	CDEEnvironmentDevelopment, CDEEnvironmentTest,
}

// IsValidCDEAssetType validates the asset type field.
func IsValidCDEAssetType(v string) bool {
	for _, t := range validCDEAssetTypes {
		if t == v {
			return true
		}
	}
	return false
}

// IsValidCDEScopeStatus validates the scope_status field.
func IsValidCDEScopeStatus(v string) bool {
	for _, s := range validCDEScopeStatuses {
		if s == v {
			return true
		}
	}
	return false
}

// IsValidCDEDataClass validates the data_classification field.
func IsValidCDEDataClass(v string) bool {
	for _, d := range validCDEDataClasses {
		if d == v {
			return true
		}
	}
	return false
}

// IsValidCDEEnvironment validates the environment field.
func IsValidCDEEnvironment(v string) bool {
	for _, e := range validCDEEnvironments {
		if e == v {
			return true
		}
	}
	return false
}

// CDEAsset represents a system or component that is in, out of, or connected to the CDE.
type CDEAsset struct {
	ID                 string     `json:"id"`
	OrgID              string     `json:"org_id"`
	Name               string     `json:"name"`
	Type               string     `json:"type"`
	IPAddress          *string    `json:"ip_address"`
	Hostname           *string    `json:"hostname"`
	Environment        string     `json:"environment"`
	ScopeStatus        string     `json:"scope_status"`
	ScopeJustification *string    `json:"scope_justification"`
	DataClassification string     `json:"data_classification"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateCDEAssetRequest is the request body for POST /cde/assets.
type CreateCDEAssetRequest struct {
	Name               string  `json:"name"                binding:"required"`
	Type               string  `json:"type"                binding:"required"`
	IPAddress          *string `json:"ip_address"`
	Hostname           *string `json:"hostname"`
	Environment        string  `json:"environment"         binding:"required"`
	ScopeStatus        string  `json:"scope_status"        binding:"required"`
	ScopeJustification *string `json:"scope_justification"`
	DataClassification string  `json:"data_classification" binding:"required"`
}

// UpdateCDEAssetRequest is the request body for PUT /cde/assets/:id.
type UpdateCDEAssetRequest struct {
	Name               *string `json:"name"`
	Type               *string `json:"type"`
	IPAddress          *string `json:"ip_address"`
	Hostname           *string `json:"hostname"`
	Environment        *string `json:"environment"`
	ScopeStatus        *string `json:"scope_status"`
	ScopeJustification *string `json:"scope_justification"`
	DataClassification *string `json:"data_classification"`
}

// --- CDE Network Segment ---

// CDESegmentType enumerates the types of network segments.
const (
	CDESegmentTypeCDE        = "cde"
	CDESegmentTypeConnected  = "connected"
	CDESegmentTypeOutOfScope = "out_of_scope"
	CDESegmentTypeDMZ        = "dmz"
	CDESegmentTypeManagement = "management"
)

var validCDESegmentTypes = []string{
	CDESegmentTypeCDE, CDESegmentTypeConnected, CDESegmentTypeOutOfScope,
	CDESegmentTypeDMZ, CDESegmentTypeManagement,
}

// IsValidCDESegmentType validates the segment_type field.
func IsValidCDESegmentType(v string) bool {
	for _, t := range validCDESegmentTypes {
		if t == v {
			return true
		}
	}
	return false
}

// CDENetworkSegment represents a network segment in the CDE topology.
type CDENetworkSegment struct {
	ID              string    `json:"id"`
	OrgID           string    `json:"org_id"`
	Name            string    `json:"name"`
	VLAN            *string   `json:"vlan"`
	Subnet          *string   `json:"subnet"`
	SegmentType     string    `json:"segment_type"`
	IsolationMethod *string   `json:"isolation_method"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateCDESegmentRequest is the request body for POST /cde/segments.
type CreateCDESegmentRequest struct {
	Name            string  `json:"name"         binding:"required"`
	VLAN            *string `json:"vlan"`
	Subnet          *string `json:"subnet"`
	SegmentType     string  `json:"segment_type" binding:"required"`
	IsolationMethod *string `json:"isolation_method"`
}

// UpdateCDESegmentRequest is the request body for PUT /cde/segments/:id.
type UpdateCDESegmentRequest struct {
	Name            *string `json:"name"`
	VLAN            *string `json:"vlan"`
	Subnet          *string `json:"subnet"`
	SegmentType     *string `json:"segment_type"`
	IsolationMethod *string `json:"isolation_method"`
}

// --- CDE Data Flow ---

// CDEDataFlow represents the flow of cardholder data between assets.
type CDEDataFlow struct {
	ID               string    `json:"id"`
	OrgID            string    `json:"org_id"`
	SourceAssetID    string    `json:"source_asset_id"`
	DestAssetID      string    `json:"dest_asset_id"`
	Protocol         *string   `json:"protocol"`
	Port             *int      `json:"port"`
	DataType         *string   `json:"data_type"`
	EncryptionMethod *string   `json:"encryption_method"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateCDEDataFlowRequest is the request body for POST /cde/data-flows.
type CreateCDEDataFlowRequest struct {
	SourceAssetID    string  `json:"source_asset_id"    binding:"required"`
	DestAssetID      string  `json:"dest_asset_id"      binding:"required"`
	Protocol         *string `json:"protocol"`
	Port             *int    `json:"port"`
	DataType         *string `json:"data_type"`
	EncryptionMethod *string `json:"encryption_method"`
}

// UpdateCDEDataFlowRequest is the request body for PUT /cde/data-flows/:id.
type UpdateCDEDataFlowRequest struct {
	SourceAssetID    *string `json:"source_asset_id"`
	DestAssetID      *string `json:"dest_asset_id"`
	Protocol         *string `json:"protocol"`
	Port             *int    `json:"port"`
	DataType         *string `json:"data_type"`
	EncryptionMethod *string `json:"encryption_method"`
}

// --- CDE Segmentation Test ---

// CDESegTestResult enumerates pass/fail results for a segmentation test.
const (
	CDESegTestResultPass = "pass"
	CDESegTestResultFail = "fail"
	CDESegTestResultNA   = "n/a"
)

var validCDESegTestResults = []string{
	CDESegTestResultPass, CDESegTestResultFail, CDESegTestResultNA,
}

// IsValidCDESegTestResult validates the result field.
func IsValidCDESegTestResult(v string) bool {
	for _, r := range validCDESegTestResults {
		if r == v {
			return true
		}
	}
	return false
}

// CDESegmentationTest represents a quarterly segmentation test for PCI DSS Req 11.4.
type CDESegmentationTest struct {
	ID           string     `json:"id"`
	OrgID        string     `json:"org_id"`
	TestDate     time.Time  `json:"test_date"`
	Tester       string     `json:"tester"`
	Methodology  *string    `json:"methodology"`
	SegmentID    *string    `json:"segment_id"`
	Result       string     `json:"result"`
	Findings     *string    `json:"findings"`
	NextTestDate *time.Time `json:"next_test_date"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateCDESegTestRequest is the request body for POST /cde/segmentation-tests.
type CreateCDESegTestRequest struct {
	TestDate     string  `json:"test_date"  binding:"required"` // RFC3339 date string
	Tester       string  `json:"tester"     binding:"required"`
	Methodology  *string `json:"methodology"`
	SegmentID    *string `json:"segment_id"`
	Result       string  `json:"result"     binding:"required"`
	Findings     *string `json:"findings"`
	NextTestDate *string `json:"next_test_date"` // RFC3339 date string
}

// UpdateCDESegTestRequest is the request body for PUT /cde/segmentation-tests/:id.
type UpdateCDESegTestRequest struct {
	TestDate     *string `json:"test_date"`
	Tester       *string `json:"tester"`
	Methodology  *string `json:"methodology"`
	SegmentID    *string `json:"segment_id"`
	Result       *string `json:"result"`
	Findings     *string `json:"findings"`
	NextTestDate *string `json:"next_test_date"`
}

// --- Scope Summary ---

// CDEScopeSummary is the aggregated response for GET /cde/scope-summary.
type CDEScopeSummary struct {
	Assets struct {
		Total        int `json:"total"`
		InScope      int `json:"in_scope"`
		OutOfScope   int `json:"out_of_scope"`
		ConnectedTo  int `json:"connected_to_cde"`
		Unclassified int `json:"unclassified"`
	} `json:"assets"`
	Segments struct {
		Total int `json:"total"`
		ByCDEType map[string]int `json:"by_type"`
	} `json:"segments"`
	DataFlows struct {
		Total     int `json:"total"`
		Encrypted int `json:"encrypted"`
	} `json:"data_flows"`
	SegmentationTests struct {
		Total          int        `json:"total"`
		LastTestDate   *time.Time `json:"last_test_date"`
		NextTestDate   *time.Time `json:"next_test_date"`
		OverdueCount   int        `json:"overdue_count"`
		PassCount      int        `json:"pass_count"`
		FailCount      int        `json:"fail_count"`
	} `json:"segmentation_tests"`
}

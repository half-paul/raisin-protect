package models

import "time"

// Risk category constants (from risk_category enum).
const (
	RiskCategoryOperational   = "operational"
	RiskCategoryFinancial     = "financial"
	RiskCategoryStrategic     = "strategic"
	RiskCategoryCompliance    = "compliance"
	RiskCategoryTechnology    = "technology"
	RiskCategoryLegal         = "legal"
	RiskCategoryReputational  = "reputational"
	RiskCategoryThirdParty    = "third_party"
	RiskCategoryPhysical      = "physical"
	RiskCategoryDataPrivacy   = "data_privacy"
	RiskCategoryCyberSecurity = "cyber_security"
	RiskCategoryHumanRes      = "human_resources"
	RiskCategoryEnvironmental = "environmental"
	RiskCategoryCustom        = "custom"
)

// ValidRiskCategories is the complete list.
var ValidRiskCategories = []string{
	RiskCategoryOperational, RiskCategoryFinancial, RiskCategoryStrategic,
	RiskCategoryCompliance, RiskCategoryTechnology, RiskCategoryLegal,
	RiskCategoryReputational, RiskCategoryThirdParty, RiskCategoryPhysical,
	RiskCategoryDataPrivacy, RiskCategoryCyberSecurity, RiskCategoryHumanRes,
	RiskCategoryEnvironmental, RiskCategoryCustom,
}

// IsValidRiskCategory checks if a category is valid.
func IsValidRiskCategory(cat string) bool {
	for _, c := range ValidRiskCategories {
		if c == cat {
			return true
		}
	}
	return false
}

// Risk status constants.
const (
	RiskStatusIdentified = "identified"
	RiskStatusOpen       = "open"
	RiskStatusAssessing  = "assessing"
	RiskStatusTreating   = "treating"
	RiskStatusMonitoring = "monitoring"
	RiskStatusAccepted   = "accepted"
	RiskStatusClosed     = "closed"
	RiskStatusArchived   = "archived"
)

// ValidRiskStatuses is the complete list.
var ValidRiskStatuses = []string{
	RiskStatusIdentified, RiskStatusOpen, RiskStatusAssessing,
	RiskStatusTreating, RiskStatusMonitoring, RiskStatusAccepted,
	RiskStatusClosed, RiskStatusArchived,
}

// IsValidRiskStatus checks if a status is valid.
func IsValidRiskStatus(s string) bool {
	for _, v := range ValidRiskStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// RiskStatusTransitions defines allowed transitions.
var RiskStatusTransitions = map[string][]string{
	RiskStatusIdentified: {RiskStatusOpen, RiskStatusAssessing},
	RiskStatusOpen:       {RiskStatusAssessing, RiskStatusTreating, RiskStatusAccepted, RiskStatusClosed},
	RiskStatusAssessing:  {RiskStatusOpen, RiskStatusTreating, RiskStatusAccepted},
	RiskStatusTreating:   {RiskStatusMonitoring, RiskStatusAccepted, RiskStatusClosed},
	RiskStatusMonitoring: {RiskStatusTreating, RiskStatusAccepted, RiskStatusClosed},
	RiskStatusAccepted:   {RiskStatusOpen, RiskStatusAssessing, RiskStatusTreating},
	RiskStatusClosed:     {RiskStatusOpen},
}

// IsValidRiskStatusTransition checks if a transition is allowed.
func IsValidRiskStatusTransition(from, to string) bool {
	allowed, ok := RiskStatusTransitions[from]
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

// Likelihood level constants.
const (
	LikelihoodRare          = "rare"
	LikelihoodUnlikely      = "unlikely"
	LikelihoodPossible      = "possible"
	LikelihoodLikely        = "likely"
	LikelihoodAlmostCertain = "almost_certain"
)

// ValidLikelihoodLevels lists valid likelihood levels.
var ValidLikelihoodLevels = []string{
	LikelihoodRare, LikelihoodUnlikely, LikelihoodPossible,
	LikelihoodLikely, LikelihoodAlmostCertain,
}

// IsValidLikelihood checks if a likelihood is valid.
func IsValidLikelihood(l string) bool {
	for _, v := range ValidLikelihoodLevels {
		if v == l {
			return true
		}
	}
	return false
}

// LikelihoodScore maps likelihood to numeric score.
func LikelihoodScore(l string) int {
	switch l {
	case LikelihoodRare:
		return 1
	case LikelihoodUnlikely:
		return 2
	case LikelihoodPossible:
		return 3
	case LikelihoodLikely:
		return 4
	case LikelihoodAlmostCertain:
		return 5
	default:
		return 0
	}
}

// Impact level constants.
const (
	ImpactNegligible = "negligible"
	ImpactMinor      = "minor"
	ImpactModerate   = "moderate"
	ImpactMajor      = "major"
	ImpactSevere     = "severe"
)

// ValidImpactLevels lists valid impact levels.
var ValidImpactLevels = []string{
	ImpactNegligible, ImpactMinor, ImpactModerate,
	ImpactMajor, ImpactSevere,
}

// IsValidImpact checks if an impact is valid.
func IsValidImpact(i string) bool {
	for _, v := range ValidImpactLevels {
		if v == i {
			return true
		}
	}
	return false
}

// ImpactScore maps impact to numeric score.
func ImpactScore(i string) int {
	switch i {
	case ImpactNegligible:
		return 1
	case ImpactMinor:
		return 2
	case ImpactModerate:
		return 3
	case ImpactMajor:
		return 4
	case ImpactSevere:
		return 5
	default:
		return 0
	}
}

// ScoreSeverity returns the severity band for a given score.
func ScoreSeverity(score float64) string {
	switch {
	case score >= 20:
		return "critical"
	case score >= 12:
		return "high"
	case score >= 6:
		return "medium"
	default:
		return "low"
	}
}

// Treatment type constants.
const (
	TreatmentMitigate = "mitigate"
	TreatmentAccept   = "accept"
	TreatmentTransfer = "transfer"
	TreatmentAvoid    = "avoid"
)

// ValidTreatmentTypes lists valid treatment types.
var ValidTreatmentTypes = []string{
	TreatmentMitigate, TreatmentAccept, TreatmentTransfer, TreatmentAvoid,
}

// IsValidTreatmentType checks if a treatment type is valid.
func IsValidTreatmentType(t string) bool {
	for _, v := range ValidTreatmentTypes {
		if v == t {
			return true
		}
	}
	return false
}

// Treatment status constants.
const (
	TreatmentStatusPlanned      = "planned"
	TreatmentStatusInProgress   = "in_progress"
	TreatmentStatusImplemented  = "implemented"
	TreatmentStatusVerified     = "verified"
	TreatmentStatusIneffective  = "ineffective"
	TreatmentStatusCancelled    = "cancelled"
)

// ValidTreatmentStatuses lists valid treatment statuses.
var ValidTreatmentStatuses = []string{
	TreatmentStatusPlanned, TreatmentStatusInProgress, TreatmentStatusImplemented,
	TreatmentStatusVerified, TreatmentStatusIneffective, TreatmentStatusCancelled,
}

// IsValidTreatmentStatus checks if a treatment status is valid.
func IsValidTreatmentStatus(s string) bool {
	for _, v := range ValidTreatmentStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// TreatmentStatusTransitions defines allowed treatment status transitions.
var TreatmentStatusTransitions = map[string][]string{
	TreatmentStatusPlanned:     {TreatmentStatusInProgress, TreatmentStatusCancelled},
	TreatmentStatusInProgress:  {TreatmentStatusImplemented, TreatmentStatusCancelled},
	TreatmentStatusImplemented: {TreatmentStatusVerified, TreatmentStatusIneffective},
	TreatmentStatusVerified:    {TreatmentStatusIneffective},
}

// IsValidTreatmentStatusTransition checks if a treatment status transition is allowed.
func IsValidTreatmentStatusTransition(from, to string) bool {
	allowed, ok := TreatmentStatusTransitions[from]
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

// Assessment type constants.
const (
	AssessmentTypeInherent = "inherent"
	AssessmentTypeResidual = "residual"
	AssessmentTypeTarget   = "target"
)

// ValidAssessmentTypes lists valid assessment types.
var ValidAssessmentTypes = []string{
	AssessmentTypeInherent, AssessmentTypeResidual, AssessmentTypeTarget,
}

// IsValidAssessmentType checks if an assessment type is valid.
func IsValidAssessmentType(t string) bool {
	for _, v := range ValidAssessmentTypes {
		if v == t {
			return true
		}
	}
	return false
}

// Control effectiveness constants.
const (
	EffectivenessEffective          = "effective"
	EffectivenessPartiallyEffective = "partially_effective"
	EffectivenessIneffective        = "ineffective"
	EffectivenessNotAssessed        = "not_assessed"
)

// ValidEffectiveness lists valid effectiveness values.
var ValidEffectiveness = []string{
	EffectivenessEffective, EffectivenessPartiallyEffective,
	EffectivenessIneffective, EffectivenessNotAssessed,
}

// IsValidEffectiveness checks if an effectiveness value is valid.
func IsValidEffectiveness(e string) bool {
	for _, v := range ValidEffectiveness {
		if v == e {
			return true
		}
	}
	return false
}

// Priority constants for treatments.
var ValidPriorities = []string{"critical", "high", "medium", "low"}

// IsValidPriority checks if a priority is valid.
func IsValidPriority(p string) bool {
	for _, v := range ValidPriorities {
		if v == p {
			return true
		}
	}
	return false
}

// Effectiveness rating constants for treatment completion.
var ValidEffectivenessRatings = []string{"highly_effective", "effective", "partially_effective", "ineffective"}

// IsValidEffectivenessRating checks if an effectiveness rating is valid.
func IsValidEffectivenessRating(r string) bool {
	for _, v := range ValidEffectivenessRatings {
		if v == r {
			return true
		}
	}
	return false
}

// --- Role sets for risk management ---

// RiskCreateRoles can create risks.
var RiskCreateRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// RiskAcceptRoles can accept risks (requires CISO or compliance_manager).
var RiskAcceptRoles = []string{RoleCISO, RoleComplianceManager}

// RiskArchiveRoles can archive risks.
var RiskArchiveRoles = []string{RoleCISO, RoleComplianceManager}

// RiskAssessRoles can create assessments.
var RiskAssessRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// RiskRecalcRoles can recalculate scores.
var RiskRecalcRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// RiskGapRoles can view gap detection.
var RiskGapRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleAuditor}

// RiskControlLinkRoles can link/unlink controls to risks.
var RiskControlLinkRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// --- Structs ---

// Risk represents a risk in the database.
type Risk struct {
	ID                      string     `json:"id"`
	OrgID                   string     `json:"org_id"`
	Identifier              string     `json:"identifier"`
	Title                   string     `json:"title"`
	Description             *string    `json:"description"`
	Category                string     `json:"category"`
	Status                  string     `json:"status"`
	OwnerID                 *string    `json:"owner_id"`
	SecondaryOwnerID        *string    `json:"secondary_owner_id"`
	InherentLikelihood      *string    `json:"inherent_likelihood"`
	InherentImpact          *string    `json:"inherent_impact"`
	InherentScore           *float64   `json:"inherent_score"`
	ResidualLikelihood      *string    `json:"residual_likelihood"`
	ResidualImpact          *string    `json:"residual_impact"`
	ResidualScore           *float64   `json:"residual_score"`
	RiskAppetiteThreshold   *float64   `json:"risk_appetite_threshold"`
	AcceptedAt              *time.Time `json:"accepted_at"`
	AcceptedBy              *string    `json:"accepted_by"`
	AcceptanceExpiry        *time.Time `json:"acceptance_expiry"`
	AcceptanceJustification *string    `json:"acceptance_justification"`
	AssessmentFrequencyDays *int       `json:"assessment_frequency_days"`
	NextAssessmentAt        *time.Time `json:"next_assessment_at"`
	LastAssessedAt          *time.Time `json:"last_assessed_at"`
	Source                  *string    `json:"source"`
	AffectedAssets          []string   `json:"affected_assets"`
	IsTemplate              bool       `json:"is_template"`
	TemplateSourceID        *string    `json:"template_source_id"`
	Tags                    []string   `json:"tags"`
	Metadata                string     `json:"metadata"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// RiskAssessment represents a point-in-time risk assessment.
type RiskAssessment struct {
	ID              string     `json:"id"`
	OrgID           string     `json:"org_id"`
	RiskID          string     `json:"risk_id"`
	AssessmentType  string     `json:"assessment_type"`
	Likelihood      string     `json:"likelihood"`
	Impact          string     `json:"impact"`
	LikelihoodScore int        `json:"likelihood_score"`
	ImpactScore     int        `json:"impact_score"`
	OverallScore    float64    `json:"overall_score"`
	ScoringFormula  string     `json:"scoring_formula"`
	Severity        string     `json:"severity"`
	Justification   *string    `json:"justification"`
	Assumptions     *string    `json:"assumptions"`
	DataSources     []string   `json:"data_sources"`
	AssessedBy      string     `json:"assessed_by"`
	AssessmentDate  time.Time  `json:"assessment_date"`
	ValidUntil      *time.Time `json:"valid_until"`
	IsCurrent       bool       `json:"is_current"`
	SupersededBy    *string    `json:"superseded_by"`
	CreatedAt       time.Time  `json:"created_at"`
}

// RiskTreatment represents a treatment plan for a risk.
type RiskTreatment struct {
	ID                        string     `json:"id"`
	OrgID                     string     `json:"org_id"`
	RiskID                    string     `json:"risk_id"`
	TreatmentType             string     `json:"treatment_type"`
	Title                     string     `json:"title"`
	Description               *string    `json:"description"`
	Status                    string     `json:"status"`
	OwnerID                   *string    `json:"owner_id"`
	Priority                  string     `json:"priority"`
	DueDate                   *time.Time `json:"due_date"`
	StartedAt                 *time.Time `json:"started_at"`
	CompletedAt               *time.Time `json:"completed_at"`
	EstimatedEffortHours      *float64   `json:"estimated_effort_hours"`
	ActualEffortHours         *float64   `json:"actual_effort_hours"`
	ExpectedResidualLikelihood *string   `json:"expected_residual_likelihood"`
	ExpectedResidualImpact    *string    `json:"expected_residual_impact"`
	ExpectedResidualScore     *float64   `json:"expected_residual_score"`
	EffectivenessRating       *string    `json:"effectiveness_rating"`
	EffectivenessNotes        *string    `json:"effectiveness_notes"`
	EffectivenessReviewedAt   *time.Time `json:"effectiveness_reviewed_at"`
	EffectivenessReviewedBy   *string    `json:"effectiveness_reviewed_by"`
	TargetControlID           *string    `json:"target_control_id"`
	Notes                     *string    `json:"notes"`
	CreatedBy                 string     `json:"created_by"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// RiskControl represents a risk-to-control linkage.
type RiskControl struct {
	ID                      string     `json:"id"`
	OrgID                   string     `json:"org_id"`
	RiskID                  string     `json:"risk_id"`
	ControlID               string     `json:"control_id"`
	Effectiveness           string     `json:"effectiveness"`
	MitigationPercentage    *int       `json:"mitigation_percentage"`
	Notes                   *string    `json:"notes"`
	LinkedBy                string     `json:"linked_by"`
	LastEffectivenessReview *time.Time `json:"last_effectiveness_review"`
	ReviewedBy              *string    `json:"reviewed_by"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// --- Request types ---

// CreateRiskRequest is the request for creating a risk.
type CreateRiskRequest struct {
	Identifier              string               `json:"identifier" binding:"required"`
	Title                   string               `json:"title" binding:"required"`
	Description             *string              `json:"description"`
	Category                string               `json:"category" binding:"required"`
	OwnerID                 *string              `json:"owner_id"`
	SecondaryOwnerID        *string              `json:"secondary_owner_id"`
	RiskAppetiteThreshold   *float64             `json:"risk_appetite_threshold"`
	AssessmentFrequencyDays *int                 `json:"assessment_frequency_days"`
	Source                  *string              `json:"source"`
	AffectedAssets          []string             `json:"affected_assets"`
	Tags                    []string             `json:"tags"`
	InitialAssessment       *InitialAssessment   `json:"initial_assessment"`
}

// InitialAssessment is an optional inline assessment when creating a risk.
type InitialAssessment struct {
	InherentLikelihood string  `json:"inherent_likelihood" binding:"required"`
	InherentImpact     string  `json:"inherent_impact" binding:"required"`
	ResidualLikelihood *string `json:"residual_likelihood"`
	ResidualImpact     *string `json:"residual_impact"`
	Justification      *string `json:"justification"`
}

// UpdateRiskRequest is the request for updating a risk.
type UpdateRiskRequest struct {
	Title                   *string  `json:"title"`
	Description             *string  `json:"description"`
	Category                *string  `json:"category"`
	OwnerID                 *string  `json:"owner_id"`
	SecondaryOwnerID        *string  `json:"secondary_owner_id"`
	RiskAppetiteThreshold   *float64 `json:"risk_appetite_threshold"`
	AssessmentFrequencyDays *int     `json:"assessment_frequency_days"`
	NextAssessmentAt        *string  `json:"next_assessment_at"`
	Source                  *string  `json:"source"`
	AffectedAssets          []string `json:"affected_assets"`
	Tags                    []string `json:"tags"`
}

// ChangeRiskStatusRequest is the request for changing risk status.
type ChangeRiskStatusRequest struct {
	Status             string  `json:"status" binding:"required"`
	Justification      *string `json:"justification"`
	AcceptanceExpiry   *string `json:"acceptance_expiry"`
}

// CreateAssessmentRequest is the request for creating a risk assessment.
type CreateAssessmentRequest struct {
	AssessmentType string   `json:"assessment_type" binding:"required"`
	Likelihood     string   `json:"likelihood" binding:"required"`
	Impact         string   `json:"impact" binding:"required"`
	ScoringFormula *string  `json:"scoring_formula"`
	Justification  *string  `json:"justification"`
	Assumptions    *string  `json:"assumptions"`
	DataSources    []string `json:"data_sources"`
	ValidUntil     *string  `json:"valid_until"`
}

// CreateTreatmentRequest is the request for creating a treatment plan.
type CreateTreatmentRequest struct {
	TreatmentType              string   `json:"treatment_type" binding:"required"`
	Title                      string   `json:"title" binding:"required"`
	Description                *string  `json:"description"`
	OwnerID                    *string  `json:"owner_id"`
	Priority                   *string  `json:"priority"`
	DueDate                    *string  `json:"due_date"`
	EstimatedEffortHours       *float64 `json:"estimated_effort_hours"`
	ExpectedResidualLikelihood *string  `json:"expected_residual_likelihood"`
	ExpectedResidualImpact     *string  `json:"expected_residual_impact"`
	TargetControlID            *string  `json:"target_control_id"`
	Notes                      *string  `json:"notes"`
}

// UpdateTreatmentRequest is the request for updating a treatment plan.
type UpdateTreatmentRequest struct {
	Title                *string  `json:"title"`
	Description          *string  `json:"description"`
	Status               *string  `json:"status"`
	OwnerID              *string  `json:"owner_id"`
	Priority             *string  `json:"priority"`
	DueDate              *string  `json:"due_date"`
	StartedAt            *string  `json:"started_at"`
	EstimatedEffortHours *float64 `json:"estimated_effort_hours"`
	ActualEffortHours    *float64 `json:"actual_effort_hours"`
	Notes                *string  `json:"notes"`
}

// CompleteTreatmentRequest is the request for marking a treatment complete.
type CompleteTreatmentRequest struct {
	ActualEffortHours   *float64 `json:"actual_effort_hours"`
	EffectivenessRating *string  `json:"effectiveness_rating"`
	EffectivenessNotes  *string  `json:"effectiveness_notes"`
}

// LinkRiskControlRequest is the request for linking a control to a risk.
type LinkRiskControlRequest struct {
	ControlID            string  `json:"control_id" binding:"required"`
	Effectiveness        *string `json:"effectiveness"`
	MitigationPercentage *int    `json:"mitigation_percentage"`
	Notes                *string `json:"notes"`
}

// UpdateRiskControlRequest is the request for updating risk-control effectiveness.
type UpdateRiskControlRequest struct {
	Effectiveness        *string `json:"effectiveness"`
	MitigationPercentage *int    `json:"mitigation_percentage"`
	Notes                *string `json:"notes"`
}

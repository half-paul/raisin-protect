package models

import "time"

// Evidence types from spec §3.4.1.
var ValidEvidenceTypes = []string{
	"screenshot", "api_response", "configuration_export", "log_sample",
	"policy_document", "access_list", "vulnerability_report", "certificate",
	"training_record", "penetration_test", "audit_report", "other",
}

// Evidence statuses.
var ValidEvidenceStatuses = []string{
	"draft", "pending_review", "approved", "rejected", "expired", "superseded",
}

// Collection methods.
var ValidCollectionMethods = []string{
	"manual_upload", "automated_pull", "api_ingestion", "screenshot_capture", "system_export",
}

// Link strength values.
var ValidLinkStrengths = []string{"primary", "supporting", "supplementary"}

// Link target types.
var ValidLinkTargetTypes = []string{"control", "requirement"}

// Allowed MIME types for evidence uploads.
var AllowedMIMETypes = []string{
	"application/pdf", "image/png", "image/jpeg", "image/gif",
	"application/json", "text/csv", "text/plain", "application/xml",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

// MaxFileSize is 100MB.
const MaxFileSize = 104857600

// EvidenceUploadRoles can create/upload evidence.
var EvidenceUploadRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleITAdmin, RoleDevOpsEngineer}

// EvidenceLinkRoles can create/delete evidence links.
var EvidenceLinkRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// EvidenceEvalRoles can evaluate evidence.
var EvidenceEvalRoles = []string{RoleCISO, RoleComplianceManager, RoleAuditor}

// EvidenceStatusRoles can change evidence status.
var EvidenceStatusRoles = []string{RoleCISO, RoleComplianceManager}

// Evidence status transitions.
var ValidEvidenceStatusTransitions = map[string][]string{
	"draft":          {"pending_review"},
	"pending_review": {"approved", "rejected"},
	"rejected":       {"pending_review"},
	"approved":       {"expired"},
	"expired":        {"pending_review"},
}

// IsValidEvidenceType checks if the evidence type is valid.
func IsValidEvidenceType(t string) bool {
	for _, v := range ValidEvidenceTypes {
		if v == t {
			return true
		}
	}
	return false
}

// IsValidEvidenceStatus checks if the evidence status is valid.
func IsValidEvidenceStatus(s string) bool {
	for _, v := range ValidEvidenceStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidCollectionMethod checks if the collection method is valid.
func IsValidCollectionMethod(m string) bool {
	for _, v := range ValidCollectionMethods {
		if v == m {
			return true
		}
	}
	return false
}

// IsValidMIMEType checks if the MIME type is allowed.
func IsValidMIMEType(m string) bool {
	for _, v := range AllowedMIMETypes {
		if v == m {
			return true
		}
	}
	return false
}

// IsValidLinkStrength checks if the link strength is valid.
func IsValidLinkStrength(s string) bool {
	for _, v := range ValidLinkStrengths {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidLinkTargetType checks if the link target type is valid.
func IsValidLinkTargetType(t string) bool {
	for _, v := range ValidLinkTargetTypes {
		if v == t {
			return true
		}
	}
	return false
}

// IsValidEvidenceStatusTransition checks if a status transition is allowed.
func IsValidEvidenceStatusTransition(from, to string) bool {
	allowed, ok := ValidEvidenceStatusTransitions[from]
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

// EvidenceArtifact represents an evidence artifact in the database.
type EvidenceArtifact struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"org_id"`
	Title               string     `json:"title"`
	Description         *string    `json:"description"`
	EvidenceType        string     `json:"evidence_type"`
	Status              string     `json:"status"`
	CollectionMethod    string     `json:"collection_method"`
	FileName            string     `json:"file_name"`
	FileSize            int64      `json:"file_size"`
	MIMEType            string     `json:"mime_type"`
	ObjectKey           string     `json:"object_key"`
	ChecksumSHA256      *string    `json:"checksum_sha256"`
	ParentArtifactID    *string    `json:"parent_artifact_id"`
	Version             int        `json:"version"`
	IsCurrent           bool       `json:"is_current"`
	CollectionDate      string     `json:"collection_date"`
	ExpiresAt           *time.Time `json:"expires_at"`
	FreshnessPeriodDays *int       `json:"freshness_period_days"`
	SourceSystem        *string    `json:"source_system"`
	UploadedBy          *string    `json:"uploaded_by"`
	Tags                []string   `json:"tags"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// CreateEvidenceRequest is the request for creating an evidence artifact.
type CreateEvidenceRequest struct {
	Title               string   `json:"title" binding:"required"`
	Description         *string  `json:"description"`
	EvidenceType        string   `json:"evidence_type" binding:"required"`
	CollectionMethod    *string  `json:"collection_method"`
	FileName            string   `json:"file_name" binding:"required"`
	FileSize            int64    `json:"file_size" binding:"required"`
	MIMEType            string   `json:"mime_type" binding:"required"`
	CollectionDate      string   `json:"collection_date" binding:"required"`
	FreshnessPeriodDays *int     `json:"freshness_period_days"`
	SourceSystem        *string  `json:"source_system"`
	Tags                []string `json:"tags"`
}

// UpdateEvidenceRequest is the request for updating evidence metadata.
type UpdateEvidenceRequest struct {
	Title               *string  `json:"title"`
	Description         *string  `json:"description"`
	EvidenceType        *string  `json:"evidence_type"`
	CollectionDate      *string  `json:"collection_date"`
	FreshnessPeriodDays *int     `json:"freshness_period_days"`
	SourceSystem        *string  `json:"source_system"`
	Tags                []string `json:"tags"`
}

// ChangeEvidenceStatusRequest is the request for changing evidence status.
type ChangeEvidenceStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ConfirmUploadRequest is the request for confirming an upload.
type ConfirmUploadRequest struct {
	ChecksumSHA256 *string `json:"checksum_sha256"`
}

// CreateVersionRequest is the request for creating a new version.
type CreateVersionRequest struct {
	Title               *string  `json:"title"`
	Description         *string  `json:"description"`
	EvidenceType        *string  `json:"evidence_type"`
	CollectionMethod    *string  `json:"collection_method"`
	FileName            string   `json:"file_name" binding:"required"`
	FileSize            int64    `json:"file_size" binding:"required"`
	MIMEType            string   `json:"mime_type" binding:"required"`
	CollectionDate      string   `json:"collection_date" binding:"required"`
	FreshnessPeriodDays *int     `json:"freshness_period_days"`
	SourceSystem        *string  `json:"source_system"`
	Tags                []string `json:"tags"`
}

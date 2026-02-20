package models

import "time"

// Valid evaluation verdicts.
var ValidEvalVerdicts = []string{"sufficient", "partial", "insufficient", "needs_update"}

// Valid confidence levels.
var ValidConfidenceLevels = []string{"high", "medium", "low"}

// IsValidEvalVerdict checks if the verdict is valid.
func IsValidEvalVerdict(v string) bool {
	for _, valid := range ValidEvalVerdicts {
		if v == valid {
			return true
		}
	}
	return false
}

// IsValidConfidenceLevel checks if the confidence level is valid.
func IsValidConfidenceLevel(c string) bool {
	for _, valid := range ValidConfidenceLevels {
		if c == valid {
			return true
		}
	}
	return false
}

// EvidenceEvaluation represents an evaluation of an evidence artifact.
type EvidenceEvaluation struct {
	ID               string    `json:"id"`
	OrgID            string    `json:"org_id"`
	ArtifactID       string    `json:"artifact_id"`
	EvidenceLinkID   *string   `json:"evidence_link_id"`
	Verdict          string    `json:"verdict"`
	Confidence       string    `json:"confidence"`
	Comments         string    `json:"comments"`
	MissingElements  []string  `json:"missing_elements"`
	RemediationNotes *string   `json:"remediation_notes"`
	EvaluatedBy      string    `json:"evaluated_by"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateEvaluationRequest is the request for creating an evaluation.
type CreateEvaluationRequest struct {
	EvidenceLinkID   *string  `json:"evidence_link_id"`
	Verdict          string   `json:"verdict" binding:"required"`
	Confidence       *string  `json:"confidence"`
	Comments         string   `json:"comments" binding:"required"`
	MissingElements  []string `json:"missing_elements"`
	RemediationNotes *string  `json:"remediation_notes"`
}

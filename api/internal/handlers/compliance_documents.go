package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListComplianceDocuments handles GET /api/v1/documents.
// Returns paginated AOC/ROC compliance documents for the org.
func ListComplianceDocuments(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// CreateComplianceDocument handles POST /api/v1/documents.
// Creates a new AOC/ROC document in draft status.
// Initial status is always "draft". Validates document_type and assessment period dates.
func CreateComplianceDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// GetComplianceDocument handles GET /api/v1/documents/:id.
// Returns a single compliance document scoped to the requesting org.
func GetComplianceDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// UpdateComplianceDocument handles PUT /api/v1/documents/:id.
// Updates editable fields on a document. Blocked for terminal states (final, signed, cancelled).
func UpdateComplianceDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// GenerateDocument handles POST /api/v1/documents/:id/generate.
// Transitions document from draft → generating and queues async generation job.
// Returns 202 Accepted. Aggregates data from controls, evidence, requirements, CDE scope.
func GenerateDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// FinalizeDocument handles POST /api/v1/documents/:id/finalize.
// Transitions document from approved → final and triggers PDF rendering.
// CISO-only (DocumentFinalizeRoles). Returns 202 Accepted.
func FinalizeDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// GetDocumentSection handles GET /api/v1/documents/:id/sections/:key.
// Returns a single section within a compliance document.
func GetDocumentSection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// UpsertDocumentSection handles PUT /api/v1/documents/:id/sections/:key.
// Creates or updates a section. Blocked for finalized/signed/cancelled documents.
func UpsertDocumentSection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// GetDocumentAttestations handles GET /api/v1/documents/:id/attestations.
// Returns all signatory attestations for a document.
func GetDocumentAttestations(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// UpsertDocumentAttestation handles PUT /api/v1/documents/:id/attestations/:role.
// Creates or updates an attestation. Validates attestation_role against ValidAttestationRoles.
func UpsertDocumentAttestation(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

// GetDocumentRequirements handles GET /api/v1/documents/:id/requirements.
// Returns the requirement compliance snapshot captured at document generation time.
func GetDocumentRequirements(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errorResponse("NOT_IMPLEMENTED", "Compliance document handler not yet implemented"))
}

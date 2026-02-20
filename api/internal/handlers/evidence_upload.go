package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

var checksumRegex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// ConfirmEvidenceUpload confirms that a file upload to MinIO is complete.
func ConfirmEvidenceUpload(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	callerID := middleware.GetUserID(c)
	callerRole := middleware.GetUserRole(c)
	artifactID := c.Param("id")

	var status, objectKey string
	var uploadedBy *string
	var fileSize int64
	err := database.QueryRow(`
		SELECT status, object_key, uploaded_by, file_size
		FROM evidence_artifacts WHERE id = $1 AND org_id = $2
	`, artifactID, orgID).Scan(&status, &objectKey, &uploadedBy, &fileSize)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get evidence for confirm")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Auth: uploader or admin
	isAdmin := models.HasRole(callerRole, models.EvidenceUploadRoles)
	isUploader := uploadedBy != nil && *uploadedBy == callerID
	if !isAdmin && !isUploader {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized"))
		return
	}

	if status != "draft" {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Upload already confirmed"))
		return
	}

	var req models.ConfirmUploadRequest
	c.ShouldBindJSON(&req) // optional body

	// Verify file in MinIO
	var actualSize int64
	if minioService != nil {
		actual, err := minioService.VerifyObjectExists(objectKey)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "File not found in storage. Upload may have failed."))
			return
		}
		actualSize = actual
	} else {
		actualSize = fileSize // no MinIO in dev/test
	}

	// Store checksum if provided
	if req.ChecksumSHA256 != nil {
		if !checksumRegex.MatchString(*req.ChecksumSHA256) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "checksum_sha256 must be a 64-character hex string"))
			return
		}
		database.Exec("UPDATE evidence_artifacts SET checksum_sha256 = $1 WHERE id = $2", *req.ChecksumSHA256, artifactID)
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":               artifactID,
		"status":           "draft",
		"file_verified":    true,
		"file_size_actual": actualSize,
		"checksum_sha256":  req.ChecksumSHA256,
		"message":          "Upload confirmed. Artifact ready for review.",
	}))
}

// GetUploadURL generates a fresh presigned upload URL for an artifact.
func GetUploadURL(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	callerID := middleware.GetUserID(c)
	callerRole := middleware.GetUserRole(c)
	artifactID := c.Param("id")

	var status, objectKey, mimeType string
	var uploadedBy *string
	err := database.QueryRow(`
		SELECT status, object_key, mime_type, uploaded_by
		FROM evidence_artifacts WHERE id = $1 AND org_id = $2
	`, artifactID, orgID).Scan(&status, &objectKey, &mimeType, &uploadedBy)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	isAdmin := models.HasRole(callerRole, models.EvidenceUploadRoles)
	isUploader := uploadedBy != nil && *uploadedBy == callerID
	if !isAdmin && !isUploader {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized"))
		return
	}

	// Check if already confirmed by verifying file exists
	if minioService != nil {
		if _, err := minioService.VerifyObjectExists(objectKey); err == nil {
			c.JSON(http.StatusConflict, errorResponse("CONFLICT", "File already uploaded"))
			return
		}
	}

	if minioService == nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse("SERVICE_UNAVAILABLE", "Storage service not available"))
		return
	}

	url, err := minioService.GenerateUploadURL(objectKey, mimeType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate upload URL")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to generate upload URL"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": artifactID,
		"upload": gin.H{
			"presigned_url": url,
			"method":        "PUT",
			"expires_in":    minioService.UploadTTLSeconds(),
			"max_size":      models.MaxFileSize,
			"content_type":  mimeType,
		},
	}))
}

// GetDownloadURL generates a presigned download URL for an evidence artifact.
func GetDownloadURL(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	artifactID := c.Param("id")

	// Check for specific version request
	versionStr := c.Query("version")
	var objectKey, fileName, mimeType, status string
	var fSize int64

	if versionStr != "" {
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid version number"))
			return
		}
		// Find the specific version
		err = database.QueryRow(`
			SELECT object_key, file_name, mime_type, file_size, status
			FROM evidence_artifacts
			WHERE (id = $1 OR parent_artifact_id = $1) AND org_id = $2 AND version = $3
		`, artifactID, orgID, version).Scan(&objectKey, &fileName, &mimeType, &fSize, &status)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Version not found"))
			return
		}
		if err != nil {
			log.Error().Err(err).Msg("Failed to get version for download")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
			return
		}
	} else {
		err := database.QueryRow(`
			SELECT object_key, file_name, mime_type, file_size, status
			FROM evidence_artifacts WHERE id = $1 AND org_id = $2
		`, artifactID, orgID).Scan(&objectKey, &fileName, &mimeType, &fSize, &status)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
			return
		}
		if err != nil {
			log.Error().Err(err).Msg("Failed to get evidence for download")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
			return
		}
	}

	if status == "draft" {
		// Check if file was actually uploaded
		if minioService != nil {
			if _, err := minioService.VerifyObjectExists(objectKey); err != nil {
				c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "File not yet uploaded"))
				return
			}
		}
	}

	if minioService == nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse("SERVICE_UNAVAILABLE", "Storage service not available"))
		return
	}

	url, err := minioService.GenerateDownloadURL(objectKey, fileName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate download URL")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to generate download URL"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":        artifactID,
		"file_name": fileName,
		"file_size": fSize,
		"mime_type": mimeType,
		"download": gin.H{
			"presigned_url": url,
			"method":        "GET",
			"expires_in":    minioService.DownloadTTLSeconds(),
		},
	}))
}

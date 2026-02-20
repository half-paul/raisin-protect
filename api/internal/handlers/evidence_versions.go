package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// CreateEvidenceVersion creates a new version of an existing evidence artifact.
func CreateEvidenceVersion(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	artifactID := c.Param("id")

	// Get the parent artifact
	var parentTitle, parentType, parentMethod string
	var parentSourceSystem *string
	var parentFreshDays *int
	var currentVersion int
	err := database.QueryRow(`
		SELECT title, evidence_type, collection_method, source_system,
			   freshness_period_days, version
		FROM evidence_artifacts WHERE id = $1 AND org_id = $2 AND is_current = TRUE
	`, artifactID, orgID).Scan(&parentTitle, &parentType, &parentMethod,
		&parentSourceSystem, &parentFreshDays, &currentVersion)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get parent artifact")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	var req models.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate file fields
	if len(req.FileName) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "File name must be at most 255 characters"))
		return
	}
	req.FileName = sanitizeFileName(req.FileName)
	if req.FileSize <= 0 || req.FileSize > models.MaxFileSize {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "File size must be between 1 byte and 100MB"))
		return
	}
	if !models.IsValidMIMEType(req.MIMEType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "MIME type not allowed"))
		return
	}
	collDate, err := time.Parse("2006-01-02", req.CollectionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid collection_date format"))
		return
	}
	if collDate.After(time.Now()) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "collection_date cannot be in the future"))
		return
	}

	// Inherit or override
	title := parentTitle
	if req.Title != nil {
		title = *req.Title
	}
	eType := parentType
	if req.EvidenceType != nil {
		if !models.IsValidEvidenceType(*req.EvidenceType) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid evidence type"))
			return
		}
		eType = *req.EvidenceType
	}
	collMethod := parentMethod
	if req.CollectionMethod != nil {
		if !models.IsValidCollectionMethod(*req.CollectionMethod) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid collection method"))
			return
		}
		collMethod = *req.CollectionMethod
	}
	sourceSys := parentSourceSystem
	if req.SourceSystem != nil {
		sourceSys = req.SourceSystem
	}
	freshDays := parentFreshDays
	if req.FreshnessPeriodDays != nil {
		freshDays = req.FreshnessPeriodDays
	}

	var expiresAt *time.Time
	if freshDays != nil {
		exp := collDate.AddDate(0, 0, *freshDays)
		expiresAt = &exp
	}

	newVersion := currentVersion + 1
	newID := uuid.New().String()

	// Determine parent_artifact_id (use original root)
	var parentArtifactID string
	var existingParent *string
	database.QueryRow("SELECT parent_artifact_id FROM evidence_artifacts WHERE id = $1", artifactID).Scan(&existingParent)
	if existingParent != nil {
		parentArtifactID = *existingParent
	} else {
		parentArtifactID = artifactID
	}

	objectKey := fmt.Sprintf("%s/%s/%d/%s", orgID, parentArtifactID, newVersion, req.FileName)

	// Mark old version as superseded
	database.Exec("UPDATE evidence_artifacts SET is_current = FALSE, status = 'superseded' WHERE id = $1", artifactID)

	// Create new version
	_, err = database.Exec(`
		INSERT INTO evidence_artifacts (id, org_id, title, description, evidence_type, status,
			collection_method, file_name, file_size, mime_type, object_key,
			parent_artifact_id, version, is_current,
			collection_date, expires_at, freshness_period_days,
			source_system, uploaded_by, tags)
		VALUES ($1, $2, $3, $4, $5, 'draft', $6, $7, $8, $9, $10,
			$11, $12, TRUE, $13, $14, $15, $16, $17, $18)
	`, newID, orgID, title, req.Description, eType,
		collMethod, req.FileName, req.FileSize, req.MIMEType, objectKey,
		parentArtifactID, newVersion,
		req.CollectionDate, expiresAt, freshDays,
		sourceSys, userID, pq.Array(req.Tags))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create evidence version")
		// Revert supersede
		database.Exec("UPDATE evidence_artifacts SET is_current = TRUE, status = 'approved' WHERE id = $1", artifactID)
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Copy evidence links from old version to new version
	database.Exec(`
		INSERT INTO evidence_links (id, org_id, artifact_id, target_type, control_id, requirement_id, notes, strength, linked_by)
		SELECT gen_random_uuid(), org_id, $1, target_type, control_id, requirement_id, notes, strength, linked_by
		FROM evidence_links WHERE artifact_id = $2
	`, newID, artifactID)

	middleware.LogAudit(c, "evidence.version_created", "evidence", &newID, map[string]interface{}{
		"parent_id": parentArtifactID, "version": newVersion,
	})

	resp := gin.H{
		"id":                  newID,
		"parent_artifact_id":  parentArtifactID,
		"version":             newVersion,
		"is_current":          true,
		"status":              "draft",
		"title":               title,
		"file_name":           req.FileName,
		"previous_version": gin.H{
			"id":      artifactID,
			"version": currentVersion,
			"status":  "superseded",
		},
		"created_at": time.Now(),
	}

	if minioService != nil {
		url, err := minioService.GenerateUploadURL(objectKey, req.MIMEType)
		if err == nil {
			resp["upload"] = gin.H{
				"presigned_url": url,
				"method":        "PUT",
				"expires_in":    minioService.UploadTTLSeconds(),
				"max_size":      models.MaxFileSize,
				"content_type":  req.MIMEType,
			}
		}
	}

	c.JSON(http.StatusCreated, successResponse(c, resp))
}

// ListEvidenceVersions returns the version history for an evidence artifact.
func ListEvidenceVersions(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	artifactID := c.Param("id")

	// Find the root of the version chain
	var rootID string
	var parentID *string
	err := database.QueryRow("SELECT id, parent_artifact_id FROM evidence_artifacts WHERE id = $1 AND org_id = $2",
		artifactID, orgID).Scan(&rootID, &parentID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	lookupID := rootID
	if parentID != nil {
		lookupID = *parentID
	}

	rows, err := database.Query(`
		SELECT ea.id, ea.version, ea.is_current, ea.title, ea.status,
			   ea.file_name, ea.file_size, ea.collection_date,
			   ea.uploaded_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   ea.created_at
		FROM evidence_artifacts ea
		LEFT JOIN users u ON u.id = ea.uploaded_by
		WHERE (ea.id = $1 OR ea.parent_artifact_id = $1) AND ea.org_id = $2
		ORDER BY ea.version DESC
	`, lookupID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list versions")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	versions := []gin.H{}
	var totalVersions, currentVersion int
	for rows.Next() {
		var vID, vTitle, vStatus, vFileName, vCollDate string
		var vVersion int
		var vIsCurrent bool
		var vFileSize int64
		var vUploadedBy *string
		var vUploaderName string
		var vCreatedAt time.Time

		if err := rows.Scan(&vID, &vVersion, &vIsCurrent, &vTitle, &vStatus,
			&vFileName, &vFileSize, &vCollDate,
			&vUploadedBy, &vUploaderName, &vCreatedAt); err != nil {
			continue
		}

		totalVersions++
		if vIsCurrent {
			currentVersion = vVersion
		}

		item := gin.H{
			"id": vID, "version": vVersion, "is_current": vIsCurrent,
			"title": vTitle, "status": vStatus,
			"file_name": vFileName, "file_size": vFileSize,
			"collection_date": vCollDate, "created_at": vCreatedAt,
		}
		if vUploadedBy != nil {
			item["uploaded_by"] = gin.H{"id": *vUploadedBy, "name": vUploaderName}
		} else {
			item["uploaded_by"] = nil
		}
		versions = append(versions, item)
	}

	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": versions,
		"meta": gin.H{
			"total_versions":  totalVersions,
			"current_version": currentVersion,
			"request_id":      reqID,
		},
	})
}

package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/half-paul/raisin-protect/api/internal/services"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var minioService *services.MinIOService

// SetMinIO sets the MinIO service for handlers.
func SetMinIO(s *services.MinIOService) {
	minioService = s
}

var fileNameSanitizer = regexp.MustCompile(`[^\w\-. ]`)

func sanitizeFileName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "\x00", "")
	return name
}

func computeFreshnessStatus(expiresAt *time.Time) string {
	if expiresAt == nil {
		return "fresh"
	}
	now := time.Now()
	if expiresAt.Before(now) {
		return "expired"
	}
	if expiresAt.Before(now.Add(30 * 24 * time.Hour)) {
		return "expiring_soon"
	}
	return "fresh"
}

func daysUntilExpiry(expiresAt *time.Time) *int {
	if expiresAt == nil {
		return nil
	}
	days := int(time.Until(*expiresAt).Hours() / 24)
	return &days
}

// ListEvidence lists evidence artifacts with filtering and pagination.
func ListEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"ea.org_id = $1", "ea.is_current = TRUE"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("include_versions"); v == "true" {
		where = where[:1] // remove is_current filter
	}

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("ea.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("evidence_type"); v != "" {
		where = append(where, fmt.Sprintf("ea.evidence_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("collection_method"); v != "" {
		where = append(where, fmt.Sprintf("ea.collection_method = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("uploaded_by"); v != "" {
		where = append(where, fmt.Sprintf("ea.uploaded_by = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("control_id"); v != "" {
		where = append(where, fmt.Sprintf("ea.id IN (SELECT artifact_id FROM evidence_links WHERE control_id = $%d AND org_id = $1)", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("requirement_id"); v != "" {
		where = append(where, fmt.Sprintf("ea.id IN (SELECT artifact_id FROM evidence_links WHERE requirement_id = $%d AND org_id = $1)", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("tags"); v != "" {
		tags := strings.Split(v, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				where = append(where, fmt.Sprintf("$%d = ANY(ea.tags)", argN))
				args = append(args, tag)
				argN++
			}
		}
	}
	if v := c.Query("freshness"); v != "" {
		switch v {
		case "fresh":
			where = append(where, "(ea.expires_at IS NULL OR ea.expires_at > NOW() + INTERVAL '30 days')")
		case "expiring_soon":
			where = append(where, "ea.expires_at IS NOT NULL AND ea.expires_at <= NOW() + INTERVAL '30 days' AND ea.expires_at > NOW()")
		case "expired":
			where = append(where, "ea.expires_at IS NOT NULL AND ea.expires_at <= NOW()")
		}
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf(
			"to_tsvector('english', ea.title || ' ' || COALESCE(ea.description, '')) @@ plainto_tsquery('english', $%d)", argN))
		args = append(args, v)
		argN++
	}

	allowedSort := map[string]string{
		"title":           "ea.title",
		"evidence_type":   "ea.evidence_type",
		"status":          "ea.status",
		"collection_date": "ea.collection_date",
		"expires_at":      "ea.expires_at",
		"created_at":      "ea.created_at",
		"file_size":       "ea.file_size",
	}
	sortField := c.DefaultQuery("sort", "created_at")
	sortCol, ok := allowedSort[sortField]
	if !ok {
		sortCol = "ea.created_at"
	}
	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM evidence_artifacts ea WHERE %s", whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ea.id, ea.title, ea.description, ea.evidence_type, ea.status,
			   ea.collection_method, ea.file_name, ea.file_size, ea.mime_type,
			   ea.version, ea.is_current, ea.collection_date, ea.expires_at,
			   ea.freshness_period_days, ea.source_system,
			   ea.uploaded_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   ea.tags, ea.created_at, ea.updated_at,
			   COALESCE((SELECT COUNT(*) FROM evidence_links el WHERE el.artifact_id = ea.id), 0),
			   COALESCE((SELECT COUNT(*) FROM evidence_evaluations ee WHERE ee.artifact_id = ea.id), 0)
		FROM evidence_artifacts ea
		LEFT JOIN users u ON u.id = ea.uploaded_by
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list evidence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, title, eType, status, method, fileName, mimeType string
			desc, sourceSystem, uploadedByID                     *string
			fileSize                                             int64
			version                                              int
			isCurrent                                            bool
			collDate                                             string
			expiresAt                                            *time.Time
			freshDays                                            *int
			uploaderName                                         string
			tags                                                 pq.StringArray
			createdAt, updatedAt                                 time.Time
			linksCount, evalsCount                               int
		)
		if err := rows.Scan(&id, &title, &desc, &eType, &status,
			&method, &fileName, &fileSize, &mimeType,
			&version, &isCurrent, &collDate, &expiresAt,
			&freshDays, &sourceSystem,
			&uploadedByID, &uploaderName,
			&tags, &createdAt, &updatedAt,
			&linksCount, &evalsCount); err != nil {
			log.Error().Err(err).Msg("Failed to scan evidence row")
			continue
		}

		item := gin.H{
			"id":                    id,
			"title":                 title,
			"description":           desc,
			"evidence_type":         eType,
			"status":                status,
			"collection_method":     method,
			"file_name":             fileName,
			"file_size":             fileSize,
			"mime_type":             mimeType,
			"version":               version,
			"is_current":            isCurrent,
			"collection_date":       collDate,
			"expires_at":            expiresAt,
			"freshness_period_days": freshDays,
			"freshness_status":      computeFreshnessStatus(expiresAt),
			"source_system":         sourceSystem,
			"tags":                  []string(tags),
			"links_count":           linksCount,
			"evaluations_count":     evalsCount,
			"created_at":            createdAt,
			"updated_at":            updatedAt,
		}

		if uploadedByID != nil {
			item["uploaded_by"] = gin.H{"id": *uploadedByID, "name": uploaderName}
		} else {
			item["uploaded_by"] = nil
		}

		// Get latest evaluation
		var evalVerdict, evalConfidence *string
		var evalAt *time.Time
		database.QueryRow(`
			SELECT verdict, confidence, created_at FROM evidence_evaluations
			WHERE artifact_id = $1 ORDER BY created_at DESC LIMIT 1
		`, id).Scan(&evalVerdict, &evalConfidence, &evalAt)
		if evalVerdict != nil {
			item["latest_evaluation"] = gin.H{
				"verdict":      *evalVerdict,
				"confidence":   *evalConfidence,
				"evaluated_at": evalAt,
			}
		} else {
			item["latest_evaluation"] = nil
		}

		results = append(results, item)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// CreateEvidence creates an evidence artifact and returns a presigned upload URL.
func CreateEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreateEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate
	if len(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must be at most 500 characters"))
		return
	}
	if req.Description != nil && len(*req.Description) > 10000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Description must be at most 10000 characters"))
		return
	}
	if !models.IsValidEvidenceType(req.EvidenceType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid evidence type"))
		return
	}
	collMethod := "manual_upload"
	if req.CollectionMethod != nil {
		if !models.IsValidCollectionMethod(*req.CollectionMethod) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid collection method"))
			return
		}
		collMethod = *req.CollectionMethod
	}
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

	// Validate collection date
	collDate, err := time.Parse("2006-01-02", req.CollectionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid collection_date format (use YYYY-MM-DD)"))
		return
	}
	if collDate.After(time.Now()) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "collection_date cannot be in the future"))
		return
	}

	if req.FreshnessPeriodDays != nil && (*req.FreshnessPeriodDays < 1 || *req.FreshnessPeriodDays > 3650) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "freshness_period_days must be between 1 and 3650"))
		return
	}
	if req.SourceSystem != nil && len(*req.SourceSystem) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "source_system must be at most 255 characters"))
		return
	}
	if len(req.Tags) > 20 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Maximum 20 tags allowed"))
		return
	}
	for _, tag := range req.Tags {
		if len(tag) > 50 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Each tag must be at most 50 characters"))
			return
		}
	}

	// Compute expires_at
	var expiresAt *time.Time
	if req.FreshnessPeriodDays != nil {
		exp := collDate.AddDate(0, 0, *req.FreshnessPeriodDays)
		expiresAt = &exp
	}

	artifactID := uuid.New().String()
	objectKey := fmt.Sprintf("%s/%s/1/%s", orgID, artifactID, req.FileName)

	_, err = database.Exec(`
		INSERT INTO evidence_artifacts (id, org_id, title, description, evidence_type, status,
			collection_method, file_name, file_size, mime_type, object_key,
			version, is_current, collection_date, expires_at, freshness_period_days,
			source_system, uploaded_by, tags)
		VALUES ($1, $2, $3, $4, $5, 'draft', $6, $7, $8, $9, $10,
			1, TRUE, $11, $12, $13, $14, $15, $16)
	`, artifactID, orgID, req.Title, req.Description, req.EvidenceType,
		collMethod, req.FileName, req.FileSize, req.MIMEType, objectKey,
		req.CollectionDate, expiresAt, req.FreshnessPeriodDays,
		req.SourceSystem, userID, pq.Array(req.Tags))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create evidence artifact")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "evidence.uploaded", "evidence", &artifactID, map[string]interface{}{
		"title": req.Title, "type": req.EvidenceType, "file_name": req.FileName,
	})

	resp := gin.H{
		"id":              artifactID,
		"title":           req.Title,
		"evidence_type":   req.EvidenceType,
		"status":          "draft",
		"file_name":       req.FileName,
		"object_key":      objectKey,
		"version":         1,
		"collection_date": req.CollectionDate,
		"expires_at":      expiresAt,
		"created_at":      time.Now(),
	}

	// Generate presigned upload URL if MinIO is available
	if minioService != nil {
		url, err := minioService.GenerateUploadURL(objectKey, req.MIMEType)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to generate presigned upload URL")
		} else {
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

// GetEvidence returns a single evidence artifact with full details.
func GetEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	artifactID := c.Param("id")

	var (
		id, title, eType, status, method, fileName, mimeType, objectKey string
		desc, checksum, parentID, sourceSystem, uploadedByID            *string
		fileSize                                                        int64
		version                                                         int
		isCurrent                                                       bool
		collDate                                                        string
		expiresAt                                                       *time.Time
		freshDays                                                       *int
		uploaderName, uploaderEmail                                     string
		tags                                                            pq.StringArray
		createdAt, updatedAt                                            time.Time
	)

	err := database.QueryRow(`
		SELECT ea.id, ea.title, ea.description, ea.evidence_type, ea.status,
			   ea.collection_method, ea.file_name, ea.file_size, ea.mime_type,
			   ea.object_key, ea.checksum_sha256,
			   ea.parent_artifact_id, ea.version, ea.is_current,
			   ea.collection_date, ea.expires_at, ea.freshness_period_days,
			   ea.source_system, ea.uploaded_by,
			   COALESCE(u.first_name || ' ' || u.last_name, ''),
			   COALESCE(u.email, ''),
			   ea.tags, ea.created_at, ea.updated_at
		FROM evidence_artifacts ea
		LEFT JOIN users u ON u.id = ea.uploaded_by
		WHERE ea.id = $1 AND ea.org_id = $2
	`, artifactID, orgID).Scan(
		&id, &title, &desc, &eType, &status,
		&method, &fileName, &fileSize, &mimeType,
		&objectKey, &checksum,
		&parentID, &version, &isCurrent,
		&collDate, &expiresAt, &freshDays,
		&sourceSystem, &uploadedByID,
		&uploaderName, &uploaderEmail,
		&tags, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get evidence artifact")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Count total versions
	var totalVersions int
	lookupID := id
	if parentID != nil {
		lookupID = *parentID
	}
	database.QueryRow(`
		SELECT COUNT(*) FROM evidence_artifacts
		WHERE (id = $1 OR parent_artifact_id = $1) AND org_id = $2
	`, lookupID, orgID).Scan(&totalVersions)

	resp := gin.H{
		"id":                    id,
		"title":                 title,
		"description":           desc,
		"evidence_type":         eType,
		"status":                status,
		"collection_method":     method,
		"file_name":             fileName,
		"file_size":             fileSize,
		"mime_type":             mimeType,
		"object_key":            objectKey,
		"checksum_sha256":       checksum,
		"version":               version,
		"is_current":            isCurrent,
		"total_versions":        totalVersions,
		"collection_date":       collDate,
		"expires_at":            expiresAt,
		"freshness_period_days": freshDays,
		"freshness_status":      computeFreshnessStatus(expiresAt),
		"days_until_expiry":     daysUntilExpiry(expiresAt),
		"source_system":         sourceSystem,
		"tags":                  []string(tags),
		"metadata":              gin.H{},
		"created_at":            createdAt,
		"updated_at":            updatedAt,
	}

	if uploadedByID != nil {
		resp["uploaded_by"] = gin.H{"id": *uploadedByID, "name": uploaderName, "email": uploaderEmail}
	} else {
		resp["uploaded_by"] = nil
	}

	// Get links
	linkRows, err := database.Query(`
		SELECT el.id, el.target_type, el.control_id, el.requirement_id,
			   el.strength, el.notes,
			   el.linked_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   el.created_at
		FROM evidence_links el
		LEFT JOIN users u ON u.id = el.linked_by
		WHERE el.artifact_id = $1 AND el.org_id = $2
		ORDER BY el.created_at
	`, id, orgID)
	if err == nil {
		defer linkRows.Close()
		links := []gin.H{}
		for linkRows.Next() {
			var lID, lType, lStrength string
			var lControlID, lRequirementID, lNotes, lLinkedByID *string
			var lLinkerName string
			var lCreatedAt time.Time
			if err := linkRows.Scan(&lID, &lType, &lControlID, &lRequirementID,
				&lStrength, &lNotes,
				&lLinkedByID, &lLinkerName,
				&lCreatedAt); err != nil {
				continue
			}
			link := gin.H{
				"id":          lID,
				"target_type": lType,
				"strength":    lStrength,
				"notes":       lNotes,
				"created_at":  lCreatedAt,
			}
			if lLinkedByID != nil {
				link["linked_by"] = gin.H{"id": *lLinkedByID, "name": lLinkerName}
			} else {
				link["linked_by"] = nil
			}

			// Fetch target details
			if lType == "control" && lControlID != nil {
				var cIdentifier, cTitle, cStatus string
				database.QueryRow("SELECT identifier, title, status FROM controls WHERE id = $1", *lControlID).
					Scan(&cIdentifier, &cTitle, &cStatus)
				link["control"] = gin.H{"id": *lControlID, "identifier": cIdentifier, "title": cTitle, "status": cStatus}
				link["requirement"] = nil
			} else if lType == "requirement" && lRequirementID != nil {
				var rIdentifier, rTitle, fName, fvVersion string
				database.QueryRow(`
					SELECT r.identifier, r.title, f.name, fv.version
					FROM requirements r
					JOIN framework_versions fv ON fv.id = r.framework_version_id
					JOIN frameworks f ON f.id = fv.framework_id
					WHERE r.id = $1
				`, *lRequirementID).Scan(&rIdentifier, &rTitle, &fName, &fvVersion)
				link["control"] = nil
				link["requirement"] = gin.H{
					"id": *lRequirementID, "identifier": rIdentifier,
					"title": rTitle, "framework": fName, "framework_version": fvVersion,
				}
			}

			links = append(links, link)
		}
		resp["links"] = links
	} else {
		resp["links"] = []gin.H{}
	}

	// Get latest evaluation
	var evalID, evalVerdict, evalConfidence, evalComments string
	var evalByID *string
	var evalByName string
	var evalCreatedAt time.Time
	err = database.QueryRow(`
		SELECT ee.id, ee.verdict, ee.confidence, ee.comments,
			   ee.evaluated_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   ee.created_at
		FROM evidence_evaluations ee
		LEFT JOIN users u ON u.id = ee.evaluated_by
		WHERE ee.artifact_id = $1 AND ee.org_id = $2
		ORDER BY ee.created_at DESC LIMIT 1
	`, id, orgID).Scan(&evalID, &evalVerdict, &evalConfidence, &evalComments,
		&evalByID, &evalByName, &evalCreatedAt)
	if err == nil {
		evalResp := gin.H{
			"id":         evalID,
			"verdict":    evalVerdict,
			"confidence": evalConfidence,
			"comments":   evalComments,
			"created_at": evalCreatedAt,
		}
		if evalByID != nil {
			evalResp["evaluated_by"] = gin.H{"id": *evalByID, "name": evalByName}
		}
		resp["latest_evaluation"] = evalResp
	} else {
		resp["latest_evaluation"] = nil
	}

	c.JSON(http.StatusOK, successResponse(c, resp))
}

// UpdateEvidence updates evidence artifact metadata.
func UpdateEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	callerID := middleware.GetUserID(c)
	callerRole := middleware.GetUserRole(c)
	artifactID := c.Param("id")

	var currentUploader *string
	err := database.QueryRow("SELECT uploaded_by FROM evidence_artifacts WHERE id = $1 AND org_id = $2",
		artifactID, orgID).Scan(&currentUploader)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	isAdmin := models.HasRole(callerRole, []string{models.RoleCISO, models.RoleComplianceManager, models.RoleSecurityEngineer})
	isUploader := currentUploader != nil && *currentUploader == callerID
	if !isAdmin && !isUploader {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to update this evidence"))
		return
	}

	var req models.UpdateEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	changes := map[string]interface{}{}

	if req.Title != nil {
		if len(*req.Title) > 500 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must be at most 500 characters"))
			return
		}
		database.Exec("UPDATE evidence_artifacts SET title = $1 WHERE id = $2", *req.Title, artifactID)
		changes["title"] = *req.Title
	}
	if req.Description != nil {
		if len(*req.Description) > 10000 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Description must be at most 10000 characters"))
			return
		}
		database.Exec("UPDATE evidence_artifacts SET description = $1 WHERE id = $2", *req.Description, artifactID)
		changes["description"] = "updated"
	}
	if req.EvidenceType != nil {
		if !models.IsValidEvidenceType(*req.EvidenceType) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid evidence type"))
			return
		}
		database.Exec("UPDATE evidence_artifacts SET evidence_type = $1 WHERE id = $2", *req.EvidenceType, artifactID)
		changes["evidence_type"] = *req.EvidenceType
	}
	if req.CollectionDate != nil {
		cd, err := time.Parse("2006-01-02", *req.CollectionDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid collection_date"))
			return
		}
		if cd.After(time.Now()) {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "collection_date cannot be in the future"))
			return
		}
		database.Exec("UPDATE evidence_artifacts SET collection_date = $1 WHERE id = $2", *req.CollectionDate, artifactID)
		changes["collection_date"] = *req.CollectionDate
	}
	if req.FreshnessPeriodDays != nil {
		if *req.FreshnessPeriodDays < 1 || *req.FreshnessPeriodDays > 3650 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "freshness_period_days must be between 1 and 3650"))
			return
		}
		// Recalculate expires_at
		var collDateStr string
		if req.CollectionDate != nil {
			collDateStr = *req.CollectionDate
		} else {
			database.QueryRow("SELECT collection_date FROM evidence_artifacts WHERE id = $1", artifactID).Scan(&collDateStr)
		}
		cd, _ := time.Parse("2006-01-02", collDateStr)
		exp := cd.AddDate(0, 0, *req.FreshnessPeriodDays)
		database.Exec("UPDATE evidence_artifacts SET freshness_period_days = $1, expires_at = $2 WHERE id = $3",
			*req.FreshnessPeriodDays, exp, artifactID)
		changes["freshness_period_days"] = *req.FreshnessPeriodDays
	}
	if req.SourceSystem != nil {
		database.Exec("UPDATE evidence_artifacts SET source_system = $1 WHERE id = $2", *req.SourceSystem, artifactID)
		changes["source_system"] = *req.SourceSystem
	}
	if req.Tags != nil {
		database.Exec("UPDATE evidence_artifacts SET tags = $1 WHERE id = $2", pq.Array(req.Tags), artifactID)
		changes["tags"] = "updated"
	}

	if len(changes) > 0 {
		middleware.LogAudit(c, "evidence.updated", "evidence", &artifactID, changes)
	}

	// Return updated artifact
	GetEvidence(c)
}

// ChangeEvidenceStatus changes an evidence artifact's lifecycle status.
func ChangeEvidenceStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	artifactID := c.Param("id")

	var currentStatus, currentTitle string
	err := database.QueryRow(
		"SELECT status, title FROM evidence_artifacts WHERE id = $1 AND org_id = $2",
		artifactID, orgID).Scan(&currentStatus, &currentTitle)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	var req models.ChangeEvidenceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if !models.IsValidEvidenceStatus(req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid evidence status"))
		return
	}
	if !models.IsValidEvidenceStatusTransition(currentStatus, req.Status) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
			fmt.Sprintf("Invalid status transition from '%s' to '%s'", currentStatus, req.Status)))
		return
	}

	_, err = database.Exec("UPDATE evidence_artifacts SET status = $1 WHERE id = $2", req.Status, artifactID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to change evidence status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "evidence.status_changed", "evidence", &artifactID, map[string]interface{}{
		"old_status": currentStatus, "new_status": req.Status,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":              artifactID,
		"status":          req.Status,
		"previous_status": currentStatus,
		"message":         "Status updated",
	}))
}

// DeleteEvidence soft-deletes an evidence artifact.
func DeleteEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	artifactID := c.Param("id")

	var currentTitle string
	err := database.QueryRow(
		"SELECT title FROM evidence_artifacts WHERE id = $1 AND org_id = $2",
		artifactID, orgID).Scan(&currentTitle)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	_, err = database.Exec("UPDATE evidence_artifacts SET status = 'superseded', is_current = FALSE WHERE id = $1", artifactID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete evidence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "evidence.deleted", "evidence", &artifactID, map[string]interface{}{
		"title": currentTitle, "artifact_id": artifactID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":      artifactID,
		"status":  "superseded",
		"message": "Evidence artifact removed from active view. File retained for audit trail.",
	}))
}

// SearchEvidence provides advanced search across evidence artifacts.
func SearchEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"ea.org_id = $1", "ea.is_current = TRUE"}
	args := []interface{}{orgID}
	argN := 2

	searchQuery := c.Query("q")
	if searchQuery != "" {
		where = append(where, fmt.Sprintf(
			"to_tsvector('english', ea.title || ' ' || COALESCE(ea.description, '') || ' ' || COALESCE(ea.source_system, '') || ' ' || array_to_string(ea.tags, ' ')) @@ plainto_tsquery('english', $%d)", argN))
		args = append(args, searchQuery)
		argN++
	}

	// Multiple evidence_type values
	if v := c.Query("evidence_type"); v != "" {
		types := strings.Split(v, ",")
		placeholders := []string{}
		for _, t := range types {
			t = strings.TrimSpace(t)
			if t != "" {
				placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
				args = append(args, t)
				argN++
			}
		}
		if len(placeholders) > 0 {
			where = append(where, fmt.Sprintf("ea.evidence_type IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	// Multiple statuses
	if v := c.Query("status"); v != "" {
		statuses := strings.Split(v, ",")
		placeholders := []string{}
		for _, s := range statuses {
			s = strings.TrimSpace(s)
			if s != "" {
				placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
				args = append(args, s)
				argN++
			}
		}
		if len(placeholders) > 0 {
			where = append(where, fmt.Sprintf("ea.status IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if v := c.Query("freshness"); v != "" {
		switch v {
		case "fresh":
			where = append(where, "(ea.expires_at IS NULL OR ea.expires_at > NOW() + INTERVAL '30 days')")
		case "expiring_soon":
			where = append(where, "ea.expires_at IS NOT NULL AND ea.expires_at <= NOW() + INTERVAL '30 days' AND ea.expires_at > NOW()")
		case "expired":
			where = append(where, "ea.expires_at IS NOT NULL AND ea.expires_at <= NOW()")
		}
	}
	if v := c.Query("date_from"); v != "" {
		where = append(where, fmt.Sprintf("ea.collection_date >= $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_to"); v != "" {
		where = append(where, fmt.Sprintf("ea.collection_date <= $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("uploaded_by"); v != "" {
		where = append(where, fmt.Sprintf("ea.uploaded_by = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("has_links"); v == "true" {
		where = append(where, "EXISTS (SELECT 1 FROM evidence_links el WHERE el.artifact_id = ea.id)")
	} else if v == "false" {
		where = append(where, "NOT EXISTS (SELECT 1 FROM evidence_links el WHERE el.artifact_id = ea.id)")
	}
	if v := c.Query("has_evaluations"); v == "true" {
		where = append(where, "EXISTS (SELECT 1 FROM evidence_evaluations ee WHERE ee.artifact_id = ea.id)")
	} else if v == "false" {
		where = append(where, "NOT EXISTS (SELECT 1 FROM evidence_evaluations ee WHERE ee.artifact_id = ea.id)")
	}
	if v := c.Query("control_id"); v != "" {
		where = append(where, fmt.Sprintf("ea.id IN (SELECT artifact_id FROM evidence_links WHERE control_id = $%d)", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("framework_id"); v != "" {
		where = append(where, fmt.Sprintf(`ea.id IN (
			SELECT el.artifact_id FROM evidence_links el
			JOIN controls c ON c.id = el.control_id
			JOIN control_mappings cm ON cm.control_id = c.id
			JOIN requirements r ON r.id = cm.requirement_id
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			WHERE fv.framework_id = $%d
		)`, argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("tags"); v != "" {
		tagList := strings.Split(v, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				where = append(where, fmt.Sprintf("$%d = ANY(ea.tags)", argN))
				args = append(args, tag)
				argN++
			}
		}
	}

	whereClause := strings.Join(where, " AND ")

	// Sort
	sortField := c.DefaultQuery("sort", "created_at")
	if searchQuery != "" && sortField == "relevance" {
		// Will handle below
	}
	allowedSort := map[string]string{
		"collection_date": "ea.collection_date",
		"expires_at":      "ea.expires_at",
		"created_at":      "ea.created_at",
		"title":           "ea.title",
	}
	sortCol, ok := allowedSort[sortField]
	if !ok {
		sortCol = "ea.created_at"
	}
	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	var total int
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM evidence_artifacts ea WHERE %s", whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ea.id, ea.title, ea.description, ea.evidence_type, ea.status,
			   ea.collection_method, ea.file_name, ea.file_size, ea.mime_type,
			   ea.version, ea.is_current, ea.collection_date, ea.expires_at,
			   ea.freshness_period_days, ea.source_system,
			   ea.uploaded_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   ea.tags, ea.created_at, ea.updated_at,
			   COALESCE((SELECT COUNT(*) FROM evidence_links el WHERE el.artifact_id = ea.id), 0),
			   COALESCE((SELECT COUNT(*) FROM evidence_evaluations ee WHERE ee.artifact_id = ea.id), 0)
		FROM evidence_artifacts ea
		LEFT JOIN users u ON u.id = ea.uploaded_by
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search evidence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, title, eType, status, method, fileName, mimeType string
			desc, sourceSystem, uploadedByID                     *string
			fileSize                                             int64
			version                                              int
			isCurrent                                            bool
			collDate                                             string
			expiresAt                                            *time.Time
			freshDays                                            *int
			uploaderName                                         string
			tags                                                 pq.StringArray
			createdAt, updatedAt                                 time.Time
			linksCount, evalsCount                               int
		)
		if err := rows.Scan(&id, &title, &desc, &eType, &status,
			&method, &fileName, &fileSize, &mimeType,
			&version, &isCurrent, &collDate, &expiresAt,
			&freshDays, &sourceSystem,
			&uploadedByID, &uploaderName,
			&tags, &createdAt, &updatedAt,
			&linksCount, &evalsCount); err != nil {
			continue
		}

		item := gin.H{
			"id": id, "title": title, "description": desc,
			"evidence_type": eType, "status": status,
			"collection_method": method, "file_name": fileName,
			"file_size": fileSize, "mime_type": mimeType,
			"version": version, "is_current": isCurrent,
			"collection_date": collDate, "expires_at": expiresAt,
			"freshness_period_days": freshDays,
			"freshness_status":      computeFreshnessStatus(expiresAt),
			"source_system": sourceSystem, "tags": []string(tags),
			"links_count": linksCount, "evaluations_count": evalsCount,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if uploadedByID != nil {
			item["uploaded_by"] = gin.H{"id": *uploadedByID, "name": uploaderName}
		} else {
			item["uploaded_by"] = nil
		}
		results = append(results, item)
	}

	resp := listResponse(c, results, total, page, perPage)
	if searchQuery != "" {
		resp["search_meta"] = gin.H{
			"query":          searchQuery,
			"matched_fields": []string{"title", "description", "tags", "source_system"},
			"suggestion":     nil,
		}
	}

	c.JSON(http.StatusOK, resp)
}

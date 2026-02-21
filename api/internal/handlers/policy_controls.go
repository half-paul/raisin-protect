package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListPolicyControls lists controls linked to a policy.
func ListPolicyControls(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	policyID := c.Param("id")

	// Verify policy exists
	var exists bool
	database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policies WHERE id = $1 AND org_id = $2)`, policyID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}

	rows, err := database.DB.Query(`
		SELECT pc.id, c.id, c.identifier, c.title, c.description, c.category, c.status,
			pc.coverage, pc.notes, pc.linked_by, pc.created_at,
			u.first_name, u.last_name
		FROM policy_controls pc
		JOIN controls c ON c.id = pc.control_id
		LEFT JOIN users u ON u.id = pc.linked_by
		WHERE pc.policy_id = $1 AND pc.org_id = $2
		ORDER BY c.identifier
	`, policyID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list policy controls")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list controls"))
		return
	}
	defer rows.Close()

	controls := []gin.H{}
	for rows.Next() {
		var (
			pcID, cID, cIdentifier, cTitle, cCategory, cStatus string
			cDescription                                        string
			coverage                                            string
			notes                                               *string
			linkedByID                                          *string
			createdAt                                           time.Time
			lbFirst, lbLast                                     *string
		)
		if err := rows.Scan(
			&pcID, &cID, &cIdentifier, &cTitle, &cDescription, &cCategory, &cStatus,
			&coverage, &notes, &linkedByID, &createdAt,
			&lbFirst, &lbLast,
		); err != nil {
			continue
		}

		ctrl := gin.H{
			"id":                cID,
			"policy_control_id": pcID,
			"identifier":        cIdentifier,
			"title":             cTitle,
			"description":       cDescription,
			"category":          cCategory,
			"status":            cStatus,
			"coverage":          coverage,
			"notes":             notes,
			"linked_at":         createdAt,
		}

		if linkedByID != nil && lbFirst != nil {
			ctrl["linked_by"] = gin.H{"id": *linkedByID, "name": *lbFirst + " " + *lbLast}
		}

		// Get frameworks for this control
		frameworks := []string{}
		fRows, err := database.DB.Query(`
			SELECT DISTINCT f.name
			FROM control_mappings cm
			JOIN requirements r ON r.id = cm.requirement_id
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			JOIN frameworks f ON f.id = fv.framework_id
			WHERE cm.control_id = $1
		`, cID)
		if err == nil {
			for fRows.Next() {
				var fName string
				if fRows.Scan(&fName) == nil {
					frameworks = append(frameworks, fName)
				}
			}
			fRows.Close()
		}
		ctrl["frameworks"] = frameworks

		controls = append(controls, ctrl)
	}

	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": controls,
		"meta": gin.H{
			"total":      len(controls),
			"request_id": reqID,
		},
	})
}

// LinkPolicyControl links a control to a policy.
func LinkPolicyControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	var req models.LinkControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	// Verify policy exists and check ownership
	var policyOwnerID *string
	err := database.DB.QueryRow(`SELECT owner_id FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&policyOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to link control"))
		return
	}

	// Auth: owner, compliance_manager, ciso, security_engineer
	isOwner := policyOwnerID != nil && *policyOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.PolicyCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to link controls"))
		return
	}

	// Verify control exists in same org
	var controlExists bool
	database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM controls WHERE id = $1 AND org_id = $2)`, req.ControlID, orgID).Scan(&controlExists)
	if !controlExists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	// Check for duplicate
	var alreadyLinked bool
	database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policy_controls WHERE org_id = $1 AND policy_id = $2 AND control_id = $3)`, orgID, policyID, req.ControlID).Scan(&alreadyLinked)
	if alreadyLinked {
		c.JSON(http.StatusConflict, errorResponse("ALREADY_LINKED", "Control is already linked to this policy"))
		return
	}

	coverage := "full"
	if req.Coverage != nil && (*req.Coverage == "full" || *req.Coverage == "partial") {
		coverage = *req.Coverage
	}

	pcID := uuid.New().String()
	_, err = database.DB.Exec(`
		INSERT INTO policy_controls (id, org_id, policy_id, control_id, coverage, notes, linked_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, pcID, orgID, policyID, req.ControlID, coverage, req.Notes, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to link control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to link control"))
		return
	}

	middleware.LogAudit(c, "policy_control.linked", "policy_control", &pcID, map[string]interface{}{
		"policy_id": policyID, "control_id": req.ControlID,
	})

	// Get control info for response
	var cIdentifier, cTitle string
	database.DB.QueryRow(`SELECT identifier, title FROM controls WHERE id = $1`, req.ControlID).Scan(&cIdentifier, &cTitle)

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":        pcID,
		"policy_id": policyID,
		"control": gin.H{
			"id":         req.ControlID,
			"identifier": cIdentifier,
			"title":      cTitle,
		},
		"coverage": coverage,
		"notes":    req.Notes,
		"linked_by": gin.H{
			"id": userID,
		},
		"created_at": time.Now(),
	}))
}

// UnlinkPolicyControl removes a control from a policy.
func UnlinkPolicyControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")
	controlID := c.Param("control_id")

	// Verify policy ownership
	var policyOwnerID *string
	err := database.DB.QueryRow(`SELECT owner_id FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&policyOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to unlink control"))
		return
	}

	isOwner := policyOwnerID != nil && *policyOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.PolicyCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to unlink controls"))
		return
	}

	result, err := database.DB.Exec(`
		DELETE FROM policy_controls WHERE policy_id = $1 AND control_id = $2 AND org_id = $3
	`, policyID, controlID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to unlink control"))
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy-control link not found"))
		return
	}

	var pcID string = policyID + ":" + controlID
	middleware.LogAudit(c, "policy_control.unlinked", "policy_control", &pcID, map[string]interface{}{
		"policy_id": policyID, "control_id": controlID,
	})

	c.Status(http.StatusNoContent)
}

// BulkLinkPolicyControls links multiple controls to a policy.
func BulkLinkPolicyControls(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	var req models.BulkLinkControlsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if len(req.Links) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "At least one link is required"))
		return
	}
	if len(req.Links) > 50 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Maximum 50 links per request"))
		return
	}

	// Verify policy and auth
	var policyOwnerID *string
	err := database.DB.QueryRow(`SELECT owner_id FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&policyOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to bulk link controls"))
		return
	}

	isOwner := policyOwnerID != nil && *policyOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.PolicyCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to link controls"))
		return
	}

	created := 0
	skipped := 0
	errors := []gin.H{}

	for _, link := range req.Links {
		// Check if control exists
		var controlExists bool
		database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM controls WHERE id = $1 AND org_id = $2)`, link.ControlID, orgID).Scan(&controlExists)
		if !controlExists {
			errors = append(errors, gin.H{"control_id": link.ControlID, "error": "Control not found"})
			continue
		}

		// Check for duplicate
		var alreadyLinked bool
		database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policy_controls WHERE org_id = $1 AND policy_id = $2 AND control_id = $3)`, orgID, policyID, link.ControlID).Scan(&alreadyLinked)
		if alreadyLinked {
			skipped++
			continue
		}

		coverage := "full"
		if link.Coverage != nil && (*link.Coverage == "full" || *link.Coverage == "partial") {
			coverage = *link.Coverage
		}

		pcID := uuid.New().String()
		_, err := database.DB.Exec(`
			INSERT INTO policy_controls (id, org_id, policy_id, control_id, coverage, notes, linked_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, pcID, orgID, policyID, link.ControlID, coverage, link.Notes, userID)
		if err != nil {
			errors = append(errors, gin.H{"control_id": link.ControlID, "error": "Insert failed"})
			continue
		}
		created++

		middleware.LogAudit(c, "policy_control.linked", "policy_control", &pcID, map[string]interface{}{
			"policy_id": policyID, "control_id": link.ControlID,
		})
	}

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"created": created,
		"skipped": skipped,
		"errors":  errors,
	}))
}

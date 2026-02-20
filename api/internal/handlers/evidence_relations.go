package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// ListControlEvidence lists all evidence artifacts linked to a specific control.
func ListControlEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	// Verify control exists
	var cIdentifier, cTitle string
	err := database.QueryRow("SELECT identifier, title FROM controls WHERE id = $1 AND org_id = $2",
		controlID, orgID).Scan(&cIdentifier, &cTitle)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"el.control_id = $1", "el.org_id = $2", "ea.is_current = TRUE"}
	args := []interface{}{controlID, orgID}
	argN := 3

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("ea.status = $%d", argN))
		args = append(args, v)
		argN++
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

	whereClause := strings.Join(where, " AND ")

	var total int
	database.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM evidence_links el
		JOIN evidence_artifacts ea ON ea.id = el.artifact_id
		WHERE %s
	`, whereClause), args...).Scan(&total)

	// Summary stats
	var approved, pending, fresh, expiringSoon, expired int
	database.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE ea.status = 'approved'),
			COUNT(*) FILTER (WHERE ea.status = 'pending_review'),
			COUNT(*) FILTER (WHERE ea.expires_at IS NULL OR ea.expires_at > NOW() + INTERVAL '30 days'),
			COUNT(*) FILTER (WHERE ea.expires_at IS NOT NULL AND ea.expires_at <= NOW() + INTERVAL '30 days' AND ea.expires_at > NOW()),
			COUNT(*) FILTER (WHERE ea.expires_at IS NOT NULL AND ea.expires_at <= NOW())
		FROM evidence_links el
		JOIN evidence_artifacts ea ON ea.id = el.artifact_id
		WHERE el.control_id = $1 AND el.org_id = $2 AND ea.is_current = TRUE
	`, controlID, orgID).Scan(&approved, &pending, &fresh, &expiringSoon, &expired)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ea.id, ea.title, ea.evidence_type, ea.status,
			   ea.collection_date, ea.expires_at,
			   el.id, el.strength, el.notes
		FROM evidence_links el
		JOIN evidence_artifacts ea ON ea.id = el.artifact_id
		WHERE %s
		ORDER BY el.strength, ea.collection_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list control evidence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	evidence := []gin.H{}
	for rows.Next() {
		var eID, eTitle, eType, eStatus, eCollDate string
		var eExpiresAt *time.Time
		var lID, lStrength string
		var lNotes *string

		if err := rows.Scan(&eID, &eTitle, &eType, &eStatus,
			&eCollDate, &eExpiresAt, &lID, &lStrength, &lNotes); err != nil {
			continue
		}

		item := gin.H{
			"id": eID, "title": eTitle, "evidence_type": eType,
			"status": eStatus, "collection_date": eCollDate,
			"expires_at":       eExpiresAt,
			"freshness_status": computeFreshnessStatus(eExpiresAt),
			"link": gin.H{
				"id":       lID,
				"strength": lStrength,
				"notes":    lNotes,
			},
		}

		// Latest evaluation
		var evalVerdict, evalConfidence *string
		database.QueryRow(`
			SELECT verdict, confidence FROM evidence_evaluations
			WHERE artifact_id = $1 ORDER BY created_at DESC LIMIT 1
		`, eID).Scan(&evalVerdict, &evalConfidence)
		if evalVerdict != nil {
			item["latest_evaluation"] = gin.H{"verdict": *evalVerdict, "confidence": *evalConfidence}
		} else {
			item["latest_evaluation"] = nil
		}

		evidence = append(evidence, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"control": gin.H{
				"id": controlID, "identifier": cIdentifier, "title": cTitle,
			},
			"evidence_summary": gin.H{
				"total": total, "approved": approved, "pending_review": pending,
				"fresh": fresh, "expiring_soon": expiringSoon, "expired": expired,
			},
			"evidence": evidence,
		},
		"meta": listResponse(c, nil, total, page, perPage)["meta"],
	})
}

// ListRequirementEvidence lists all evidence linked to a specific requirement.
func ListRequirementEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	requirementID := c.Param("id")

	// Verify requirement exists
	var rIdentifier, rTitle, fName, fvVersion string
	err := database.QueryRow(`
		SELECT r.identifier, r.title, f.name, fv.version
		FROM requirements r
		JOIN framework_versions fv ON fv.id = r.framework_version_id
		JOIN frameworks f ON f.id = fv.framework_id
		WHERE r.id = $1
	`, requirementID).Scan(&rIdentifier, &rTitle, &fName, &fvVersion)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Requirement not found"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	includeTransitive := c.DefaultQuery("include_transitive", "true") == "true"

	// Direct evidence links
	evidence := []gin.H{}

	directRows, err := database.Query(`
		SELECT ea.id, ea.title, ea.evidence_type, ea.status, ea.expires_at,
			   el.strength
		FROM evidence_links el
		JOIN evidence_artifacts ea ON ea.id = el.artifact_id
		WHERE el.requirement_id = $1 AND el.org_id = $2 AND ea.is_current = TRUE
		ORDER BY ea.collection_date DESC
	`, requirementID, orgID)
	if err == nil {
		defer directRows.Close()
		for directRows.Next() {
			var eID, eTitle, eType, eStatus, eStrength string
			var eExpiresAt *time.Time
			if err := directRows.Scan(&eID, &eTitle, &eType, &eStatus, &eExpiresAt, &eStrength); err != nil {
				continue
			}
			evidence = append(evidence, gin.H{
				"id": eID, "title": eTitle, "evidence_type": eType,
				"status": eStatus, "freshness_status": computeFreshnessStatus(eExpiresAt),
				"link_type": "direct", "via_control": nil, "strength": eStrength,
			})
		}
	}

	// Transitive evidence (via controls mapped to this requirement)
	if includeTransitive {
		transitiveRows, err := database.Query(`
			SELECT DISTINCT ea.id, ea.title, ea.evidence_type, ea.status, ea.expires_at,
				   el.strength, c.id, c.identifier, c.title
			FROM control_mappings cm
			JOIN evidence_links el ON el.control_id = cm.control_id AND el.org_id = cm.org_id
			JOIN evidence_artifacts ea ON ea.id = el.artifact_id
			JOIN controls c ON c.id = cm.control_id
			WHERE cm.requirement_id = $1 AND cm.org_id = $2 AND ea.is_current = TRUE
			ORDER BY ea.collection_date DESC
		`, requirementID, orgID)
		if err == nil {
			defer transitiveRows.Close()
			seen := map[string]bool{}
			for transitiveRows.Next() {
				var eID, eTitle, eType, eStatus, eStrength, cID, cIdentifier, cTitle string
				var eExpiresAt *time.Time
				if err := transitiveRows.Scan(&eID, &eTitle, &eType, &eStatus, &eExpiresAt,
					&eStrength, &cID, &cIdentifier, &cTitle); err != nil {
					continue
				}
				// Deduplicate (same evidence might come through multiple controls)
				key := eID + cID
				if seen[key] {
					continue
				}
				seen[key] = true

				evidence = append(evidence, gin.H{
					"id": eID, "title": eTitle, "evidence_type": eType,
					"status": eStatus, "freshness_status": computeFreshnessStatus(eExpiresAt),
					"link_type": "transitive",
					"via_control": gin.H{
						"id": cID, "identifier": cIdentifier, "title": cTitle,
					},
					"strength": eStrength,
				})
			}
		}
	}

	// Pagination is applied in-memory for the combined set
	total := len(evidence)
	offset := (page - 1) * perPage
	end := offset + perPage
	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}
	paged := evidence[offset:end]

	_ = pq.Array // import used in other handlers

	reqIDVal, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"requirement": gin.H{
				"id": requirementID, "identifier": rIdentifier, "title": rTitle,
				"framework": fName, "framework_version": fvVersion,
			},
			"evidence": paged,
		},
		"meta": gin.H{
			"total": total, "page": page, "per_page": perPage,
			"request_id": reqIDVal,
		},
	})
}

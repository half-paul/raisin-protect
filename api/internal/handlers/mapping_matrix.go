package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// GetMappingMatrix generates the cross-framework mapping matrix.
func GetMappingMatrix(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	// Parse params
	frameworkIDsStr := c.Query("framework_ids")
	controlCategory := c.Query("control_category")
	controlStatus := c.DefaultQuery("control_status", "active")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	// Get activated frameworks for this org
	fwQuery := `
		SELECT f.id, f.identifier, f.name, fv.version, fv.id
		FROM org_frameworks of
		JOIN frameworks f ON f.id = of.framework_id
		JOIN framework_versions fv ON fv.id = of.active_version_id
		WHERE of.org_id = $1 AND of.status = 'active'
	`
	fwArgs := []interface{}{orgID}
	fwArgN := 2

	if frameworkIDsStr != "" {
		ids := strings.Split(frameworkIDsStr, ",")
		placeholders := []string{}
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id != "" {
				placeholders = append(placeholders, fmt.Sprintf("$%d", fwArgN))
				fwArgs = append(fwArgs, id)
				fwArgN++
			}
		}
		if len(placeholders) > 0 {
			fwQuery += fmt.Sprintf(" AND f.id IN (%s)", strings.Join(placeholders, ","))
		}
	}
	fwQuery += " ORDER BY f.name"

	fwRows, err := database.Query(fwQuery, fwArgs...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get frameworks for matrix")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer fwRows.Close()

	frameworks := []gin.H{}
	fwIdentifiers := map[string]string{} // fwID -> identifier
	fwVersionIDs := []string{}
	for fwRows.Next() {
		var fID, fIdentifier, fName, fvVersion, fvID string
		fwRows.Scan(&fID, &fIdentifier, &fName, &fvVersion, &fvID)
		frameworks = append(frameworks, gin.H{
			"id":         fID,
			"identifier": fIdentifier,
			"name":       fName,
			"version":    fvVersion,
		})
		fwIdentifiers[fID] = fIdentifier
		fwVersionIDs = append(fwVersionIDs, fvID)
	}

	// Get controls with filters
	where := []string{"c.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if controlStatus != "" {
		where = append(where, fmt.Sprintf("c.status = $%d", argN))
		args = append(args, controlStatus)
		argN++
	}
	if controlCategory != "" {
		where = append(where, fmt.Sprintf("c.category = $%d", argN))
		args = append(args, controlCategory)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(c.identifier ILIKE $%d OR c.title ILIKE $%d)", argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var totalControls int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM controls c WHERE %s", whereClause), countArgs...).Scan(&totalControls)

	offset := (page - 1) * perPage
	ctrlQuery := fmt.Sprintf(`
		SELECT c.id, c.identifier, c.title, c.category, c.status
		FROM controls c
		WHERE %s
		ORDER BY c.identifier
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	ctrlRows, err := database.Query(ctrlQuery, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get controls for matrix")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer ctrlRows.Close()

	controls := []gin.H{}
	for ctrlRows.Next() {
		var cID, cIdentifier, cTitle, cCategory, cStatus string
		ctrlRows.Scan(&cID, &cIdentifier, &cTitle, &cCategory, &cStatus)

		// Get mappings grouped by framework
		mappingsByFw := map[string][]gin.H{}
		mRows, err := database.Query(`
			SELECT f.identifier, r.id, r.identifier, cm.strength
			FROM control_mappings cm
			JOIN requirements r ON r.id = cm.requirement_id
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			JOIN frameworks f ON f.id = fv.framework_id
			WHERE cm.control_id = $1 AND cm.org_id = $2
			ORDER BY f.identifier, r.identifier
		`, cID, orgID)
		if err == nil {
			for mRows.Next() {
				var fIdent, rID, rIdent, strength string
				mRows.Scan(&fIdent, &rID, &rIdent, &strength)
				mappingsByFw[fIdent] = append(mappingsByFw[fIdent], gin.H{
					"requirement_id": rID,
					"identifier":     rIdent,
					"strength":       strength,
				})
			}
			mRows.Close()
		}

		controls = append(controls, gin.H{
			"id":                     cID,
			"identifier":             cIdentifier,
			"title":                  cTitle,
			"category":               cCategory,
			"status":                 cStatus,
			"mappings_by_framework":  mappingsByFw,
		})
	}

	reqID, _ := c.Get("request_id")
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"frameworks": frameworks,
			"controls":   controls,
		},
		"meta": gin.H{
			"total_controls": totalControls,
			"page":           page,
			"per_page":       perPage,
			"request_id":     reqID,
		},
	})
}

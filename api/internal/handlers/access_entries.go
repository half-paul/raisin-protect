package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
)

// ListAccessEntries returns a paginated list of access entries with filters.
func ListAccessEntries(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	where := []string{"e.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("resource_id"); v != "" {
		where = append(where, fmt.Sprintf("e.resource_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("provider_id"); v != "" {
		where = append(where, fmt.Sprintf("e.identity_provider_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("e.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("department"); v != "" {
		where = append(where, fmt.Sprintf("e.user_department = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("is_privileged"); v != "" {
		where = append(where, fmt.Sprintf("e.is_privileged = $%d", argN))
		args = append(args, v == "true")
		argN++
	}
	if v := c.Query("has_role_drift"); v != "" {
		where = append(where, fmt.Sprintf("e.has_role_drift = $%d", argN))
		args = append(args, v == "true")
		argN++
	}
	if v := c.Query("has_anomalies"); v != "" {
		if v == "true" {
			where = append(where, "jsonb_array_length(e.anomalies) > 0")
		} else {
			where = append(where, "(e.anomalies IS NULL OR jsonb_array_length(e.anomalies) = 0)")
		}
	}
	if v := c.Query("is_service_account"); v != "" {
		where = append(where, fmt.Sprintf("e.is_service_account = $%d", argN))
		args = append(args, v == "true")
		argN++
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(e.user_email ILIKE $%d OR e.user_display_name ILIKE $%d OR e.role_name ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	whereClause := "WHERE " + join(where, " AND ")

	var total int
	err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM access_entries e %s", whereClause), args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count entries"))
		return
	}

	queryArgs := append(args, perPage, offset)
	rows, err := database.Query(fmt.Sprintf(`
		SELECT e.id, e.org_id, e.identity_provider_id, e.resource_id, e.external_user_id, e.user_email,
		       e.user_display_name, e.user_department, e.user_title, e.user_manager_email, e.internal_user_id,
		       e.role_name, e.access_level, e.permissions, e.is_privileged, e.expected_role, e.expected_access_level,
		       e.has_role_drift, e.granted_at, e.last_used_at, e.last_login_at, e.status, e.mfa_enabled,
		       e.anomalies, e.last_sync_at, e.is_service_account, e.notes, e.created_at, e.updated_at,
		       r.name as resource_name, r.criticality as resource_criticality
		FROM access_entries e
		JOIN access_resources r ON e.resource_id = r.id
		%s
		ORDER BY e.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1), queryArgs...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query entries"))
		return
	}
	defer rows.Close()

	type EntryWithResource struct {
		models.AccessEntry
		ResourceName        string `json:"resource_name"`
		ResourceCriticality string `json:"resource_criticality"`
	}

	entries := []EntryWithResource{}
	for rows.Next() {
		var e EntryWithResource
		var permJSON, anomJSON []byte
		err := rows.Scan(
			&e.ID, &e.OrgID, &e.IdentityProviderID, &e.ResourceID, &e.ExternalUserID, &e.UserEmail,
			&e.UserDisplayName, &e.UserDepartment, &e.UserTitle, &e.UserManagerEmail, &e.InternalUserID,
			&e.RoleName, &e.AccessLevel, &permJSON, &e.IsPrivileged, &e.ExpectedRole, &e.ExpectedLevel,
			&e.HasRoleDrift, &e.GrantedAt, &e.LastUsedAt, &e.LastLoginAt, &e.Status, &e.MFAEnabled,
			&anomJSON, &e.LastSyncAt, &e.IsServiceAccount, &e.Notes, &e.CreatedAt, &e.UpdatedAt,
			&e.ResourceName, &e.ResourceCriticality,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan entry"))
			return
		}
		json.Unmarshal(permJSON, &e.Permissions)
		json.Unmarshal(anomJSON, &e.Anomalies)
		entries = append(entries, e)
	}

	c.JSON(http.StatusOK, listResponse(c, entries, total, page, perPage))
}

// GetAccessEntry returns a single access entry with review history.
func GetAccessEntry(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var e models.AccessEntry
	var permJSON, anomJSON []byte
	var resourceName, resourceCriticality string
	var providerName *string

	err := database.QueryRow(`
		SELECT e.id, e.org_id, e.identity_provider_id, e.resource_id, e.external_user_id, e.user_email,
		       e.user_display_name, e.user_department, e.user_title, e.user_manager_email, e.internal_user_id,
		       e.role_name, e.access_level, e.permissions, e.is_privileged, e.expected_role, e.expected_access_level,
		       e.has_role_drift, e.granted_at, e.last_used_at, e.last_login_at, e.status, e.mfa_enabled,
		       e.anomalies, e.last_sync_at, e.is_service_account, e.notes, e.created_at, e.updated_at,
		       r.name, r.criticality,
		       ip.name
		FROM access_entries e
		JOIN access_resources r ON e.resource_id = r.id
		LEFT JOIN identity_providers ip ON e.identity_provider_id = ip.id
		WHERE e.id = $1 AND e.org_id = $2
	`, id, orgID).Scan(
		&e.ID, &e.OrgID, &e.IdentityProviderID, &e.ResourceID, &e.ExternalUserID, &e.UserEmail,
		&e.UserDisplayName, &e.UserDepartment, &e.UserTitle, &e.UserManagerEmail, &e.InternalUserID,
		&e.RoleName, &e.AccessLevel, &permJSON, &e.IsPrivileged, &e.ExpectedRole, &e.ExpectedLevel,
		&e.HasRoleDrift, &e.GrantedAt, &e.LastUsedAt, &e.LastLoginAt, &e.Status, &e.MFAEnabled,
		&anomJSON, &e.LastSyncAt, &e.IsServiceAccount, &e.Notes, &e.CreatedAt, &e.UpdatedAt,
		&resourceName, &resourceCriticality,
		&providerName,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Access entry not found"))
		return
	}

	json.Unmarshal(permJSON, &e.Permissions)
	json.Unmarshal(anomJSON, &e.Anomalies)

	// Fetch review history
	type ReviewHistoryItem struct {
		CampaignID    string     `json:"campaign_id"`
		CampaignName  string     `json:"campaign_name"`
		Decision      string     `json:"decision"`
		DecidedByName *string    `json:"decided_by_name"`
		DecidedAt     *time.Time `json:"decided_at"`
		Justification *string    `json:"justification"`
	}

	histRows, err := database.Query(`
		SELECT ar.campaign_id, c.name, ar.decision,
		       u.full_name, ar.decided_at, ar.justification
		FROM access_reviews ar
		JOIN access_review_campaigns c ON ar.campaign_id = c.id
		LEFT JOIN users u ON ar.decided_by = u.id
		WHERE ar.entry_id = $1 AND ar.org_id = $2 AND ar.decision != 'pending'
		ORDER BY ar.decided_at DESC
	`, id, orgID)

	reviewHistory := []ReviewHistoryItem{}
	if err == nil {
		defer histRows.Close()
		for histRows.Next() {
			var h ReviewHistoryItem
			histRows.Scan(&h.CampaignID, &h.CampaignName, &h.Decision, &h.DecidedByName, &h.DecidedAt, &h.Justification)
			reviewHistory = append(reviewHistory, h)
		}
	}

	result := gin.H{
		"id":                    e.ID,
		"org_id":                e.OrgID,
		"resource_id":           e.ResourceID,
		"resource_name":         resourceName,
		"resource_criticality":  resourceCriticality,
		"identity_provider_id":  e.IdentityProviderID,
		"identity_provider_name": providerName,
		"external_user_id":      e.ExternalUserID,
		"user_email":            e.UserEmail,
		"user_display_name":     e.UserDisplayName,
		"user_department":       e.UserDepartment,
		"user_title":            e.UserTitle,
		"user_manager_email":    e.UserManagerEmail,
		"internal_user_id":      e.InternalUserID,
		"role_name":             e.RoleName,
		"access_level":          e.AccessLevel,
		"permissions":           e.Permissions,
		"is_privileged":         e.IsPrivileged,
		"expected_role":         e.ExpectedRole,
		"expected_access_level": e.ExpectedLevel,
		"has_role_drift":        e.HasRoleDrift,
		"granted_at":            e.GrantedAt,
		"last_used_at":          e.LastUsedAt,
		"last_login_at":         e.LastLoginAt,
		"status":                e.Status,
		"mfa_enabled":           e.MFAEnabled,
		"is_service_account":    e.IsServiceAccount,
		"anomalies":             e.Anomalies,
		"review_history":        reviewHistory,
		"last_sync_at":          e.LastSyncAt,
		"notes":                 e.Notes,
		"created_at":            e.CreatedAt,
		"updated_at":            e.UpdatedAt,
	}

	c.JSON(http.StatusOK, successResponse(c, result))
}

// GetAccessEntryAnomalies returns an aggregated anomaly summary across all access entries.
func GetAccessEntryAnomalies(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	where := []string{"e.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("resource_id"); v != "" {
		where = append(where, fmt.Sprintf("e.resource_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("department"); v != "" {
		where = append(where, fmt.Sprintf("e.user_department = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("criticality"); v != "" {
		where = append(where, fmt.Sprintf("r.criticality = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := "WHERE " + join(where, " AND ")

	// Count total entries and entries with anomalies
	var totalEntries, entriesWithAnomalies int
	err := database.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE jsonb_array_length(e.anomalies) > 0)
		FROM access_entries e
		JOIN access_resources r ON e.resource_id = r.id
		%s
	`, whereClause), args...).Scan(&totalEntries, &entriesWithAnomalies)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query anomaly summary"))
		return
	}

	// Get anomaly breakdown by type
	anomRows, err := database.Query(fmt.Sprintf(`
		SELECT a.value->>'type' as anomaly_type, COUNT(*) as cnt
		FROM access_entries e
		JOIN access_resources r ON e.resource_id = r.id,
		jsonb_array_elements(e.anomalies) a
		%s
		GROUP BY a.value->>'type'
		ORDER BY cnt DESC
	`, whereClause), args...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query anomaly breakdown"))
		return
	}
	defer anomRows.Close()

	anomalyDescriptions := map[string]struct {
		Description string
		Severity    string
	}{
		"stale_access":         {"Access not used in 90+ days", "medium"},
		"orphaned_account":     {"No matching employee record found", "high"},
		"excessive_privileges": {"More access than role requires", "high"},
		"role_drift":           {"Actual access diverges from expected", "medium"},
		"no_mfa":               {"Account lacks MFA on critical resource", "critical"},
		"departed_user":        {"User marked as departed in HRIS", "critical"},
	}

	type AnomalyBreakdown struct {
		Type        string `json:"type"`
		Count       int    `json:"count"`
		Description string `json:"description"`
		Severity    string `json:"severity"`
	}

	breakdowns := []AnomalyBreakdown{}
	for anomRows.Next() {
		var ab AnomalyBreakdown
		anomRows.Scan(&ab.Type, &ab.Count)
		if info, ok := anomalyDescriptions[ab.Type]; ok {
			ab.Description = info.Description
			ab.Severity = info.Severity
		}
		breakdowns = append(breakdowns, ab)
	}

	// By resource criticality
	critRows, err := database.Query(fmt.Sprintf(`
		SELECT r.criticality, COUNT(DISTINCT e.id)
		FROM access_entries e
		JOIN access_resources r ON e.resource_id = r.id
		%s AND jsonb_array_length(e.anomalies) > 0
		GROUP BY r.criticality
	`, whereClause), args...)

	byCriticality := map[string]int{}
	if err == nil {
		defer critRows.Close()
		for critRows.Next() {
			var crit string
			var cnt int
			critRows.Scan(&crit, &cnt)
			byCriticality[crit] = cnt
		}
	}

	// By department
	deptRows, err := database.Query(fmt.Sprintf(`
		SELECT COALESCE(e.user_department, 'Unknown'), COUNT(DISTINCT e.id)
		FROM access_entries e
		JOIN access_resources r ON e.resource_id = r.id
		%s AND jsonb_array_length(e.anomalies) > 0
		GROUP BY e.user_department
	`, whereClause), args...)

	byDepartment := map[string]int{}
	if err == nil {
		defer deptRows.Close()
		for deptRows.Next() {
			var dept string
			var cnt int
			deptRows.Scan(&dept, &cnt)
			byDepartment[dept] = cnt
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"total_entries":          totalEntries,
		"entries_with_anomalies": entriesWithAnomalies,
		"anomaly_breakdown":      breakdowns,
		"by_resource_criticality": byCriticality,
		"by_department":          byDepartment,
	}))
}

// DetectAnomalies triggers anomaly detection across access entries.
func DetectAnomalies(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req struct {
		ResourceIDs            []string `json:"resource_ids"`
		StaleDaysThreshold     int      `json:"stale_days_threshold"`
		IncludeServiceAccounts bool     `json:"include_service_accounts"`
	}

	// Body is optional
	c.ShouldBindJSON(&req)

	if req.StaleDaysThreshold <= 0 {
		req.StaleDaysThreshold = 90
	}

	where := []string{"e.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if len(req.ResourceIDs) > 0 {
		placeholders := ""
		for i, rid := range req.ResourceIDs {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += fmt.Sprintf("$%d", argN)
			args = append(args, rid)
			argN++
		}
		where = append(where, fmt.Sprintf("e.resource_id IN (%s)", placeholders))
	}
	if !req.IncludeServiceAccounts {
		where = append(where, "e.is_service_account = FALSE")
	}

	whereClause := "WHERE " + join(where, " AND ")

	// Count entries to scan
	var entriesScanned int
	err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM access_entries e %s", whereClause), args...).Scan(&entriesScanned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count entries"))
		return
	}

	// Detect stale access
	staleArgs := append(args, req.StaleDaysThreshold)
	staleQuery := fmt.Sprintf(`
		UPDATE access_entries e
		SET anomalies = COALESCE(anomalies, '[]'::jsonb) || jsonb_build_array(jsonb_build_object(
			'type', 'stale_access',
			'detected_at', NOW(),
			'details', 'Access not used in ' || $%d || '+ days'
		)),
		updated_at = NOW()
		%s
		AND e.status = 'active'
		AND e.last_used_at < NOW() - ($%d || ' days')::interval
		AND NOT EXISTS (
			SELECT 1 FROM jsonb_array_elements(COALESCE(e.anomalies, '[]'::jsonb)) a
			WHERE a->>'type' = 'stale_access'
		)
	`, argN, whereClause, argN)

	staleResult, _ := database.Exec(staleQuery, staleArgs...)
	staleCount, _ := staleResult.RowsAffected()

	// Detect role drift
	driftQuery := fmt.Sprintf(`
		UPDATE access_entries e
		SET anomalies = COALESCE(anomalies, '[]'::jsonb) || jsonb_build_array(jsonb_build_object(
			'type', 'role_drift',
			'detected_at', NOW(),
			'details', 'Actual role ' || e.role_name || ' differs from expected ' || COALESCE(e.expected_role, 'N/A')
		)),
		has_role_drift = TRUE,
		updated_at = NOW()
		%s
		AND e.expected_role IS NOT NULL
		AND e.role_name != e.expected_role
		AND NOT EXISTS (
			SELECT 1 FROM jsonb_array_elements(COALESCE(e.anomalies, '[]'::jsonb)) a
			WHERE a->>'type' = 'role_drift'
		)
	`, whereClause)

	driftResult, _ := database.Exec(driftQuery, args...)
	driftCount, _ := driftResult.RowsAffected()

	// Detect no MFA on critical resources
	mfaQuery := fmt.Sprintf(`
		UPDATE access_entries e
		SET anomalies = COALESCE(anomalies, '[]'::jsonb) || jsonb_build_array(jsonb_build_object(
			'type', 'no_mfa',
			'detected_at', NOW(),
			'details', 'Account lacks MFA on critical resource'
		)),
		updated_at = NOW()
		FROM access_resources r
		%s
		AND e.resource_id = r.id
		AND r.criticality = 'critical'
		AND (e.mfa_enabled = FALSE OR e.mfa_enabled IS NULL)
		AND NOT EXISTS (
			SELECT 1 FROM jsonb_array_elements(COALESCE(e.anomalies, '[]'::jsonb)) a
			WHERE a->>'type' = 'no_mfa'
		)
	`, whereClause)

	mfaResult, _ := database.Exec(mfaQuery, args...)
	mfaCount, _ := mfaResult.RowsAffected()

	totalNew := int(staleCount + driftCount + mfaCount)

	// Count total anomalies after detection
	var totalAnomalies int
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM access_entries e %s AND jsonb_array_length(COALESCE(e.anomalies, '[]'::jsonb)) > 0", whereClause), args...).Scan(&totalAnomalies)

	middleware.LogAudit(c, "anomaly_detection.triggered", "access_entry", nil, map[string]interface{}{
		"entries_scanned":    entriesScanned,
		"anomalies_detected": totalAnomalies,
		"new_anomalies":      totalNew,
	})

	c.JSON(http.StatusAccepted, successResponse(c, gin.H{
		"status":              "started",
		"message":             fmt.Sprintf("Anomaly detection started for %d access entries", entriesScanned),
		"entries_scanned":     entriesScanned,
		"anomalies_detected":  totalAnomalies,
		"new_anomalies":       totalNew,
		"resolved_anomalies":  0,
	}))
}

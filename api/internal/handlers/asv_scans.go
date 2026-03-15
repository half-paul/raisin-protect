package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// asvRoleIn checks if role is in the allowed set.
func asvRoleIn(role string, allowed []string) bool {
	for _, r := range allowed {
		if r == role {
			return true
		}
	}
	return false
}

// asvSelectCols is the ordered column list for SELECT queries.
const asvSelectCols = `id, org_id, asv_vendor, scan_type, quarter, year, scan_date,
       status, findings_count, critical_count, high_count, medium_count,
       low_count, informational_count, remediation_deadline, report_path,
       import_format, raw_findings, import_notes, imported_by, reviewed_by,
       created_by, created_at, updated_at`

// scanASVRow scans a full asv_scans row into an ASVScan struct.
func scanASVRow(row interface {
	Scan(...interface{}) error
}) (*models.ASVScan, error) {
	var s models.ASVScan
	err := row.Scan(
		&s.ID, &s.OrgID, &s.ASVVendor, &s.ScanType,
		&s.Quarter, &s.Year, &s.ScanDate,
		&s.Status,
		&s.FindingsCount, &s.CriticalCount, &s.HighCount, &s.MediumCount,
		&s.LowCount, &s.InformationalCount,
		&s.RemediationDeadline, &s.ReportPath,
		&s.ImportFormat, &s.RawFindings, &s.ImportNotes,
		&s.ImportedBy, &s.ReviewedBy, &s.CreatedBy,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return &s, err
}

// insertASVScan executes the INSERT and returns the created ASVScan.
// Columns: id, org_id, asv_vendor, scan_type, quarter, year, scan_date, status,
//
//	remediation_deadline, import_notes, created_by
func insertASVScan(id, orgID, asvVendor, scanType string, quarter, year int,
	scanDate time.Time, status string, remediationDeadline *time.Time,
	importNotes *string, createdBy string,
) (*models.ASVScan, error) {
	row := database.DB.QueryRow(
		fmt.Sprintf(`INSERT INTO asv_scans
			(id, org_id, asv_vendor, scan_type, quarter, year, scan_date, status,
			 remediation_deadline, import_notes, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING %s`, asvSelectCols),
		id, orgID, asvVendor, scanType, quarter, year, scanDate, status,
		remediationDeadline, importNotes, createdBy,
	)
	return scanASVRow(row)
}

// =============================================================================
// ListASVScans — GET /api/v1/asv-scans
// =============================================================================

func ListASVScans(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	role := middleware.GetUserRole(c)

	if !asvRoleIn(role, models.ASVViewRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
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

	where := []string{"org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("year"); v != "" {
		if yr, err := strconv.Atoi(v); err == nil {
			where = append(where, fmt.Sprintf("year = $%d", argN))
			args = append(args, yr)
			argN++
		}
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := database.DB.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM asv_scans WHERE %s", whereClause),
		countArgs...,
	).Scan(&total)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count asv_scans")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to count scans"))
		return
	}

	offset := (page - 1) * perPage
	dataArgs := append(args, perPage, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(
			"SELECT %s FROM asv_scans WHERE %s ORDER BY year DESC, quarter DESC, created_at DESC LIMIT $%d OFFSET $%d",
			asvSelectCols, whereClause, argN, argN+1,
		),
		dataArgs...,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list asv_scans")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list scans"))
		return
	}
	defer rows.Close()

	scans := make([]*models.ASVScan, 0)
	for rows.Next() {
		s, err := scanASVRow(rows)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan asv_scan row")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read scan data"))
			return
		}
		scans = append(scans, s)
	}

	c.JSON(http.StatusOK, listResponse(c, scans, total, page, perPage))
}

// =============================================================================
// CreateASVScan — POST /api/v1/asv-scans
// =============================================================================

func CreateASVScan(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if !asvRoleIn(role, models.ASVManageRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
		return
	}

	var req models.CreateASVScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Missing required fields"))
		return
	}

	scanDate, err := time.Parse(time.RFC3339, req.ScanDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid scan_date format; expected RFC3339"))
		return
	}

	if req.Quarter < 1 || req.Quarter > 4 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "quarter must be between 1 and 4"))
		return
	}

	scanType := models.ASVScanTypeExternal
	if req.ScanType != nil && *req.ScanType != "" {
		scanType = *req.ScanType
	}

	status := models.ASVStatusInProgress
	if req.Status != nil && *req.Status != "" {
		if !models.IsValidASVStatus(*req.Status) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid status value"))
			return
		}
		status = *req.Status
	}

	var exists bool
	err = database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM asv_scans WHERE org_id = $1 AND scan_type = $2 AND quarter = $3 AND year = $4)",
		orgID, scanType, req.Quarter, req.Year,
	).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check asv_scan uniqueness")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to validate scan uniqueness"))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse("DUPLICATE_SCAN", "A scan already exists for this org/type/quarter/year"))
		return
	}

	var remediationDeadline *time.Time
	if req.RemediationDeadline != nil && *req.RemediationDeadline != "" {
		rd, err := time.Parse(time.RFC3339, *req.RemediationDeadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid remediation_deadline format"))
			return
		}
		remediationDeadline = &rd
	}

	id := uuid.New().String()
	scan, err := insertASVScan(id, orgID, req.ASVVendor, scanType, req.Quarter, req.Year,
		scanDate, status, remediationDeadline, req.ImportNotes, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert asv_scan")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create scan"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(c, scan))
}

// =============================================================================
// GetASVScan — GET /api/v1/asv-scans/:id
// =============================================================================

func GetASVScan(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	role := middleware.GetUserRole(c)

	if !asvRoleIn(role, models.ASVViewRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
		return
	}

	scanID := c.Param("id")
	row := database.DB.QueryRow(
		fmt.Sprintf("SELECT %s FROM asv_scans WHERE id = $1 AND org_id = $2", asvSelectCols),
		scanID, orgID,
	)
	scan, err := scanASVRow(row)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "ASV scan not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get asv_scan")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to retrieve scan"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, scan))
}

// =============================================================================
// UpdateASVScan — PUT /api/v1/asv-scans/:id
// =============================================================================

func UpdateASVScan(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	role := middleware.GetUserRole(c)

	if !asvRoleIn(role, models.ASVManageRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
		return
	}

	scanID := c.Param("id")

	var req models.UpdateASVScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if req.Status != nil && !models.IsValidASVStatus(*req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid status value"))
		return
	}

	var exists bool
	err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM asv_scans WHERE id = $1 AND org_id = $2)",
		scanID, orgID,
	).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check asv_scan existence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to verify scan"))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "ASV scan not found"))
		return
	}

	// Build dynamic SET. Order: status, finding counts, then other fields.
	setClauses := []string{}
	args := []interface{}{}
	argN := 1

	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argN))
		args = append(args, *req.Status)
		argN++
	}
	if req.FindingsCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("findings_count = $%d", argN))
		args = append(args, *req.FindingsCount)
		argN++
	}
	if req.CriticalCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("critical_count = $%d", argN))
		args = append(args, *req.CriticalCount)
		argN++
	}
	if req.HighCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("high_count = $%d", argN))
		args = append(args, *req.HighCount)
		argN++
	}
	if req.MediumCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("medium_count = $%d", argN))
		args = append(args, *req.MediumCount)
		argN++
	}
	if req.LowCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("low_count = $%d", argN))
		args = append(args, *req.LowCount)
		argN++
	}
	if req.InformationalCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("informational_count = $%d", argN))
		args = append(args, *req.InformationalCount)
		argN++
	}
	if req.RemediationDeadline != nil {
		rd, err := time.Parse(time.RFC3339, *req.RemediationDeadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid remediation_deadline format"))
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("remediation_deadline = $%d", argN))
		args = append(args, rd)
		argN++
	}
	if req.ReportPath != nil {
		setClauses = append(setClauses, fmt.Sprintf("report_path = $%d", argN))
		args = append(args, *req.ReportPath)
		argN++
	}
	if req.ImportFormat != nil {
		setClauses = append(setClauses, fmt.Sprintf("import_format = $%d", argN))
		args = append(args, *req.ImportFormat)
		argN++
	}
	if req.ImportNotes != nil {
		setClauses = append(setClauses, fmt.Sprintf("import_notes = $%d", argN))
		args = append(args, *req.ImportNotes)
		argN++
	}
	if req.ASVVendor != nil {
		setClauses = append(setClauses, fmt.Sprintf("asv_vendor = $%d", argN))
		args = append(args, *req.ASVVendor)
		argN++
	}
	if req.ScanDate != nil {
		sd, err := time.Parse(time.RFC3339, *req.ScanDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid scan_date format"))
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("scan_date = $%d", argN))
		args = append(args, sd)
		argN++
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "No fields to update"))
		return
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, scanID, orgID)

	query := fmt.Sprintf(
		"UPDATE asv_scans SET %s WHERE id = $%d AND org_id = $%d RETURNING %s",
		strings.Join(setClauses, ", "), argN, argN+1, asvSelectCols,
	)

	row := database.DB.QueryRow(query, args...)
	scan, err := scanASVRow(row)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "ASV scan not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to update asv_scan")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update scan"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, scan))
}

// =============================================================================
// DeleteASVScan — DELETE /api/v1/asv-scans/:id
// =============================================================================

func DeleteASVScan(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	role := middleware.GetUserRole(c)

	if !asvRoleIn(role, models.ASVManageRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
		return
	}

	scanID := c.Param("id")
	result, err := database.DB.Exec(
		"DELETE FROM asv_scans WHERE id = $1 AND org_id = $2",
		scanID, orgID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete asv_scan")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete scan"))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "ASV scan not found"))
		return
	}

	c.Status(http.StatusNoContent)
}

// =============================================================================
// GetASVQuarterlyStatus — GET /api/v1/asv-scans/quarterly-status
// =============================================================================

func GetASVQuarterlyStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	role := middleware.GetUserRole(c)

	if !asvRoleIn(role, models.ASVViewRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
		return
	}

	yearFilter := 0
	if v := c.Query("year"); v != "" {
		if yr, err := strconv.Atoi(v); err == nil {
			yearFilter = yr
		}
	}

	// Build expected periods
	type periodKey struct{ Year, Quarter int }
	var expectedPeriods []periodKey

	if yearFilter > 0 {
		for q := 1; q <= 4; q++ {
			expectedPeriods = append(expectedPeriods, periodKey{yearFilter, q})
		}
	} else {
		// Last 5 quarters from today
		now := time.Now()
		yr := now.Year()
		q := (int(now.Month())-1)/3 + 1
		for i := 0; i < 5; i++ {
			expectedPeriods = append(expectedPeriods, periodKey{yr, q})
			q--
			if q < 1 {
				q = 4
				yr--
			}
		}
	}

	// Query existing scans for these periods
	query := "SELECT year, quarter, status, id FROM asv_scans WHERE org_id = $1"
	args := []interface{}{orgID}
	if yearFilter > 0 {
		query += " AND year = $2"
		args = append(args, yearFilter)
	}
	query += " ORDER BY year DESC, quarter DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query asv quarterly status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to retrieve quarterly status"))
		return
	}
	defer rows.Close()

	scanMap := make(map[periodKey]*models.ASVQuarterlyStatus)
	for rows.Next() {
		var year, quarter int
		var status, scanID string
		if err := rows.Scan(&year, &quarter, &status, &scanID); err != nil {
			log.Error().Err(err).Msg("Failed to scan quarterly status row")
			continue
		}
		key := periodKey{year, quarter}
		s := status
		sid := scanID
		scanMap[key] = &models.ASVQuarterlyStatus{
			Year:    year,
			Quarter: quarter,
			Status:  &s,
			ScanID:  &sid,
		}
	}

	periods := make([]models.ASVQuarterlyStatus, 0, len(expectedPeriods))
	for _, p := range expectedPeriods {
		if s, ok := scanMap[p]; ok {
			periods = append(periods, *s)
		} else {
			periods = append(periods, models.ASVQuarterlyStatus{
				Year:    p.Year,
				Quarter: p.Quarter,
				Status:  nil,
				ScanID:  nil,
			})
		}
	}

	c.JSON(http.StatusOK, successResponse(c, models.ASVQuarterlyReport{
		OrgID:   orgID,
		Periods: periods,
	}))
}

// =============================================================================
// ImportASVScan — POST /api/v1/asv-scans/import
// =============================================================================

type importASVScanRequest struct {
	ASVVendor           string  `json:"asv_vendor"   binding:"required"`
	ScanType            *string `json:"scan_type"`
	Quarter             int     `json:"quarter"      binding:"required"`
	Year                int     `json:"year"         binding:"required"`
	ScanDate            string  `json:"scan_date"    binding:"required"`
	ImportFormat        *string `json:"import_format"`
	RawFindings         *string `json:"raw_findings"`
	RemediationDeadline *string `json:"remediation_deadline"`
	ImportNotes         *string `json:"import_notes"`
}

var validASVImportFormats = map[string]bool{
	models.ASVImportCSV:    true,
	models.ASVImportXML:    true,
	models.ASVImportPDF:    true,
	models.ASVImportJSON:   true,
	models.ASVImportManual: true,
}

func ImportASVScan(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if !asvRoleIn(role, models.ASVManageRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
		return
	}

	var req importASVScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Missing required fields"))
		return
	}

	scanDate, err := time.Parse(time.RFC3339, req.ScanDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid scan_date format; expected RFC3339"))
		return
	}

	if req.Quarter < 1 || req.Quarter > 4 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "quarter must be between 1 and 4"))
		return
	}

	if req.ImportFormat != nil && *req.ImportFormat != "" && !validASVImportFormats[*req.ImportFormat] {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid import_format; must be csv, xml, pdf, json, or manual"))
		return
	}

	scanType := models.ASVScanTypeExternal
	if req.ScanType != nil && *req.ScanType != "" {
		scanType = *req.ScanType
	}

	// Imported scans default to "pass" (scan result already completed by ASV)
	status := models.ASVStatusPass

	var exists bool
	err = database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM asv_scans WHERE org_id = $1 AND scan_type = $2 AND quarter = $3 AND year = $4)",
		orgID, scanType, req.Quarter, req.Year,
	).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check asv_scan uniqueness for import")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to validate scan uniqueness"))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse("DUPLICATE_SCAN", "A scan already exists for this org/type/quarter/year"))
		return
	}

	var remediationDeadline *time.Time
	if req.RemediationDeadline != nil && *req.RemediationDeadline != "" {
		rd, err := time.Parse(time.RFC3339, *req.RemediationDeadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid remediation_deadline format"))
			return
		}
		remediationDeadline = &rd
	}

	id := uuid.New().String()
	scan, err := insertASVScan(id, orgID, req.ASVVendor, scanType, req.Quarter, req.Year,
		scanDate, status, remediationDeadline, req.ImportNotes, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert imported asv_scan")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to import scan"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(c, scan))
}

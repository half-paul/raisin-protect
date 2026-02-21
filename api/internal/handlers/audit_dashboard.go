package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
)

// GetAuditDashboard returns audit hub dashboard statistics.
func GetAuditDashboard(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	// Summary stats
	var activeAudits, completedAudits, totalOpenReqs, totalOverdueReqs, totalOpenFindings, criticalFindings, highFindings int

	database.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status IN ('planning','fieldwork','review','draft_report','management_response','final_report')),
			COUNT(*) FILTER (WHERE status = 'completed')
		FROM audits WHERE org_id = $1
	`, orgID).Scan(&activeAudits, &completedAudits)

	database.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE ar.status NOT IN ('accepted','closed')),
			COUNT(*) FILTER (WHERE ar.due_date < CURRENT_DATE AND ar.status NOT IN ('accepted','closed'))
		FROM audit_requests ar
		JOIN audits a ON ar.audit_id = a.id
		WHERE ar.org_id = $1 AND a.status NOT IN ('completed','cancelled')
	`, orgID).Scan(&totalOpenReqs, &totalOverdueReqs)

	database.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE af.status NOT IN ('verified','closed','risk_accepted')),
			COUNT(*) FILTER (WHERE af.severity = 'critical' AND af.status NOT IN ('verified','closed','risk_accepted')),
			COUNT(*) FILTER (WHERE af.severity = 'high' AND af.status NOT IN ('verified','closed','risk_accepted'))
		FROM audit_findings af
		JOIN audits a ON af.audit_id = a.id
		WHERE af.org_id = $1 AND a.status NOT IN ('completed','cancelled')
	`, orgID).Scan(&totalOpenFindings, &criticalFindings, &highFindings)

	summary := gin.H{
		"active_audits":          activeAudits,
		"completed_audits":       completedAudits,
		"total_open_requests":    totalOpenReqs,
		"total_overdue_requests": totalOverdueReqs,
		"total_open_findings":    totalOpenFindings,
		"critical_findings":      criticalFindings,
		"high_findings":          highFindings,
	}

	// Active audits detail
	activeAuditsList := []gin.H{}
	rows, err := database.DB.Query(`
		SELECT a.id, a.title, a.audit_type, a.status, a.planned_end,
		       a.total_requests, a.open_requests, a.total_findings, a.open_findings,
		       a.milestones::text
		FROM audits a
		WHERE a.org_id = $1 AND a.status NOT IN ('completed','cancelled')
		ORDER BY a.planned_end ASC NULLS LAST
	`, orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var (
				id, title, auditType, status string
				plannedEnd                   *time.Time
				totalReqs, openReqs          int
				totalFnds, openFnds          int
				milestonesJSON               string
			)
			rows.Scan(&id, &title, &auditType, &status, &plannedEnd,
				&totalReqs, &openReqs, &totalFnds, &openFnds, &milestonesJSON)

			var daysRemaining int
			if plannedEnd != nil {
				daysRemaining = int(time.Until(*plannedEnd).Hours() / 24)
			}

			readinessPct := 0
			if totalReqs > 0 {
				accepted := totalReqs - openReqs
				readinessPct = (accepted * 100) / totalReqs
			}

			activeAuditsList = append(activeAuditsList, gin.H{
				"id": id, "title": title, "audit_type": auditType, "status": status,
				"planned_end": plannedEnd, "days_remaining": daysRemaining,
				"readiness_pct": readinessPct,
				"total_requests": totalReqs, "open_requests": openReqs,
				"total_findings": totalFnds, "open_findings": openFnds,
			})
		}
	}

	// Overdue requests
	overdueRequests := []gin.H{}
	oRows, err := database.DB.Query(`
		SELECT ar.id, ar.title, a.title, ar.due_date,
		       COALESCE(u.first_name || ' ' || u.last_name, ''), ar.priority
		FROM audit_requests ar
		JOIN audits a ON ar.audit_id = a.id
		LEFT JOIN users u ON ar.assigned_to = u.id
		WHERE ar.org_id = $1
		  AND ar.due_date < CURRENT_DATE
		  AND ar.status NOT IN ('accepted','closed')
		  AND a.status NOT IN ('completed','cancelled')
		ORDER BY ar.due_date ASC
		LIMIT 10
	`, orgID)
	if err == nil {
		defer oRows.Close()
		for oRows.Next() {
			var id, title, auditTitle, priority string
			var dueDate time.Time
			var assignedToName string
			oRows.Scan(&id, &title, &auditTitle, &dueDate, &assignedToName, &priority)
			daysOverdue := int(time.Since(dueDate).Hours() / 24)
			overdueRequests = append(overdueRequests, gin.H{
				"id": id, "title": title, "audit_title": auditTitle,
				"due_date": dueDate, "days_overdue": daysOverdue,
				"assigned_to_name": assignedToName, "priority": priority,
			})
		}
	}

	// Critical findings
	criticalFindingsList := []gin.H{}
	cRows, err := database.DB.Query(`
		SELECT af.id, af.title, a.title, af.severity, af.status,
		       af.remediation_due_date, COALESCE(ro.first_name || ' ' || ro.last_name, '')
		FROM audit_findings af
		JOIN audits a ON af.audit_id = a.id
		LEFT JOIN users ro ON af.remediation_owner_id = ro.id
		WHERE af.org_id = $1
		  AND af.severity IN ('critical','high')
		  AND af.status NOT IN ('verified','closed','risk_accepted')
		  AND a.status NOT IN ('completed','cancelled')
		ORDER BY af.severity ASC, af.created_at DESC
		LIMIT 10
	`, orgID)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var id, title, auditTitle, severity, status string
			var remDueDate *time.Time
			var remOwnerName string
			cRows.Scan(&id, &title, &auditTitle, &severity, &status, &remDueDate, &remOwnerName)
			criticalFindingsList = append(criticalFindingsList, gin.H{
				"id": id, "title": title, "audit_title": auditTitle,
				"severity": severity, "status": status,
				"remediation_due_date": remDueDate, "remediation_owner_name": remOwnerName,
			})
		}
	}

	// Recent activity (from audit_log)
	recentActivity := []gin.H{}
	rRows, err := database.DB.Query(`
		SELECT al.action, al.resource_type, al.resource_id,
		       COALESCE(u.first_name || ' ' || u.last_name, ''),
		       al.metadata::text, al.created_at
		FROM audit_log al
		LEFT JOIN users u ON al.actor_id = u.id
		WHERE al.org_id = $1
		  AND al.action LIKE 'audit%'
		ORDER BY al.created_at DESC
		LIMIT 10
	`, orgID)
	if err == nil {
		defer rRows.Close()
		for rRows.Next() {
			var action, resType, actorName, metadataStr string
			var resID *string
			var ts time.Time
			rRows.Scan(&action, &resType, &resID, &actorName, &metadataStr, &ts)
			recentActivity = append(recentActivity, gin.H{
				"type": action, "resource_type": resType, "resource_id": resID,
				"actor_name": actorName, "timestamp": ts,
			})
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"summary":           summary,
		"active_audits":     activeAuditsList,
		"overdue_requests":  overdueRequests,
		"critical_findings": criticalFindingsList,
		"recent_activity":   recentActivity,
	}))
}

// GetAuditStats returns statistics for a specific audit engagement.
func GetAuditStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var title, status string
	var plannedStart, plannedEnd, actualStart *time.Time
	database.DB.QueryRow("SELECT title, status, planned_start, planned_end, actual_start FROM audits WHERE id = $1 AND org_id = $2",
		auditID, orgID).Scan(&title, &status, &plannedStart, &plannedEnd, &actualStart)

	// Request readiness
	var totalReqs, acceptedReqs, submittedReqs, inProgressReqs, openReqs, rejectedReqs, overdueReqs int
	database.DB.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'accepted'),
			COUNT(*) FILTER (WHERE status = 'submitted'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status = 'open'),
			COUNT(*) FILTER (WHERE status = 'rejected'),
			COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND status NOT IN ('accepted','closed'))
		FROM audit_requests WHERE audit_id = $1 AND org_id = $2
	`, auditID, orgID).Scan(&totalReqs, &acceptedReqs, &submittedReqs, &inProgressReqs, &openReqs, &rejectedReqs, &overdueReqs)

	readinessPct := 0
	if totalReqs > 0 {
		readinessPct = (acceptedReqs * 100) / totalReqs
	}

	// Findings breakdown
	var totalFindings int
	bySeverity := gin.H{"critical": 0, "high": 0, "medium": 0, "low": 0, "informational": 0}
	byStatus := gin.H{}

	fSevRows, err := database.DB.Query(
		"SELECT severity, COUNT(*) FROM audit_findings WHERE audit_id = $1 AND org_id = $2 GROUP BY severity",
		auditID, orgID,
	)
	if err == nil {
		defer fSevRows.Close()
		for fSevRows.Next() {
			var sev string
			var cnt int
			fSevRows.Scan(&sev, &cnt)
			bySeverity[sev] = cnt
			totalFindings += cnt
		}
	}

	fStatRows, err := database.DB.Query(
		"SELECT status, COUNT(*) FROM audit_findings WHERE audit_id = $1 AND org_id = $2 GROUP BY status",
		auditID, orgID,
	)
	if err == nil {
		defer fStatRows.Close()
		for fStatRows.Next() {
			var stat string
			var cnt int
			fStatRows.Scan(&stat, &cnt)
			byStatus[stat] = cnt
		}
	}

	var overdueRemediation int
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM audit_findings
		WHERE audit_id = $1 AND org_id = $2
		  AND remediation_due_date < CURRENT_DATE
		  AND status NOT IN ('verified','closed','risk_accepted')
	`, auditID, orgID).Scan(&overdueRemediation)

	// Evidence stats
	var totalEvSubmitted, evAccepted, evPendingReview, evRejected int
	database.DB.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'accepted'),
			COUNT(*) FILTER (WHERE status = 'pending_review'),
			COUNT(*) FILTER (WHERE status = 'rejected')
		FROM audit_evidence_links WHERE audit_id = $1 AND org_id = $2
	`, auditID, orgID).Scan(&totalEvSubmitted, &evAccepted, &evPendingReview, &evRejected)

	// Timeline
	var daysElapsed, daysRemaining int
	if actualStart != nil {
		daysElapsed = int(time.Since(*actualStart).Hours() / 24)
	}
	if plannedEnd != nil {
		daysRemaining = int(time.Until(*plannedEnd).Hours() / 24)
	}

	// Comments count
	var commentsCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM audit_comments WHERE audit_id = $1 AND org_id = $2", auditID, orgID).Scan(&commentsCount)

	var lastActivityAt *time.Time
	database.DB.QueryRow("SELECT MAX(created_at) FROM audit_comments WHERE audit_id = $1 AND org_id = $2", auditID, orgID).Scan(&lastActivityAt)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"audit_id": auditID,
		"title":    title,
		"status":   status,
		"readiness": gin.H{
			"total_requests": totalReqs,
			"accepted":       acceptedReqs,
			"submitted":      submittedReqs,
			"in_progress":    inProgressReqs,
			"open":           openReqs,
			"rejected":       rejectedReqs,
			"overdue":        overdueReqs,
			"readiness_pct":  readinessPct,
		},
		"findings": gin.H{
			"total":                totalFindings,
			"by_severity":         bySeverity,
			"by_status":           byStatus,
			"overdue_remediation": overdueRemediation,
		},
		"evidence": gin.H{
			"total_submitted":  totalEvSubmitted,
			"accepted":         evAccepted,
			"pending_review":   evPendingReview,
			"rejected":         evRejected,
		},
		"timeline": gin.H{
			"planned_start":  plannedStart,
			"planned_end":    plannedEnd,
			"actual_start":   actualStart,
			"days_elapsed":   daysElapsed,
			"days_remaining": daysRemaining,
		},
		"activity": gin.H{
			"comments_count":   commentsCount,
			"last_activity_at": lastActivityAt,
		},
	}))
}

// GetAuditReadiness returns audit readiness breakdown.
func GetAuditReadiness(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	// Overall readiness
	var totalReqs, acceptedReqs int
	database.DB.QueryRow(`
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'accepted')
		FROM audit_requests WHERE audit_id = $1 AND org_id = $2
	`, auditID, orgID).Scan(&totalReqs, &acceptedReqs)

	overallPct := 0
	if totalReqs > 0 {
		overallPct = (acceptedReqs * 100) / totalReqs
	}

	// By requirement
	byRequirement := []gin.H{}
	rRows, err := database.DB.Query(`
		SELECT r.id, r.title,
		       COUNT(ar.id) AS total,
		       COUNT(ar.id) FILTER (WHERE ar.status = 'accepted') AS accepted
		FROM audit_requests ar
		JOIN requirements r ON ar.requirement_id = r.id
		WHERE ar.audit_id = $1 AND ar.org_id = $2 AND ar.requirement_id IS NOT NULL
		GROUP BY r.id, r.title
		ORDER BY r.title
	`, auditID, orgID)
	if err == nil {
		defer rRows.Close()
		for rRows.Next() {
			var rID, rTitle string
			var total, accepted int
			rRows.Scan(&rID, &rTitle, &total, &accepted)
			pct := 0
			if total > 0 {
				pct = (accepted * 100) / total
			}
			byRequirement = append(byRequirement, gin.H{
				"requirement_id":    rID,
				"requirement_title": rTitle,
				"total_requests":    total,
				"accepted_requests": accepted,
				"readiness_pct":     pct,
			})
		}
	}

	// By control
	byControl := []gin.H{}
	ctrRows, err := database.DB.Query(`
		SELECT ctrl.id, ctrl.title,
		       COUNT(ar.id) AS total,
		       COUNT(ar.id) FILTER (WHERE ar.status = 'accepted') AS accepted
		FROM audit_requests ar
		JOIN controls ctrl ON ar.control_id = ctrl.id
		WHERE ar.audit_id = $1 AND ar.org_id = $2 AND ar.control_id IS NOT NULL
		GROUP BY ctrl.id, ctrl.title
		ORDER BY ctrl.title
	`, auditID, orgID)
	if err == nil {
		defer ctrRows.Close()
		for ctrRows.Next() {
			var cID, cTitle string
			var total, accepted int
			ctrRows.Scan(&cID, &cTitle, &total, &accepted)
			pct := 0
			if total > 0 {
				pct = (accepted * 100) / total
			}
			byControl = append(byControl, gin.H{
				"control_id":        cID,
				"control_title":     cTitle,
				"total_requests":    total,
				"accepted_requests": accepted,
				"readiness_pct":     pct,
			})
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"audit_id":             auditID,
		"overall_readiness_pct": overallPct,
		"by_requirement":       byRequirement,
		"by_control":           byControl,
	}))
}

// Ensure imports are used
var _ = models.AuditStatusPlanning

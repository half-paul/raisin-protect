// Package workers provides background job processors for the monitoring engine.
package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// MonitoringWorker polls for due tests and executes them.
type MonitoringWorker struct {
	DB       *sql.DB
	Interval time.Duration
	WorkerID string
}

// NewMonitoringWorker creates a new monitoring worker.
func NewMonitoringWorker(db *sql.DB, interval time.Duration) *MonitoringWorker {
	return &MonitoringWorker{
		DB:       db,
		Interval: interval,
		WorkerID: fmt.Sprintf("worker-%s", uuid.New().String()[:8]),
	}
}

// Run starts the monitoring worker loop.
func (w *MonitoringWorker) Run(ctx context.Context) {
	log.Info().Str("worker_id", w.WorkerID).Dur("interval", w.Interval).Msg("Monitoring worker started")

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("worker_id", w.WorkerID).Msg("Monitoring worker stopped")
			return
		case <-ticker.C:
			w.processDueTests(ctx)
			w.checkSLABreaches(ctx)
			w.unsuppressExpiredAlerts(ctx)
		}
	}
}

// processDueTests finds and executes tests that are due.
func (w *MonitoringWorker) processDueTests(ctx context.Context) {
	// Find tests due for execution, grouped by org
	rows, err := w.DB.QueryContext(ctx, `
		SELECT id, org_id, identifier, title, test_type, severity, control_id,
			schedule_cron, schedule_interval_min, timeout_seconds, test_config
		FROM tests
		WHERE status = 'active' AND next_run_at IS NOT NULL AND next_run_at <= NOW()
		ORDER BY org_id, next_run_at
		LIMIT 100
	`)
	if err != nil {
		log.Error().Err(err).Msg("Worker: failed to query due tests")
		return
	}
	defer rows.Close()

	type dueTest struct {
		ID, OrgID, Identifier, Title, TestType, Severity, ControlID string
		ScheduleCron                                                 *string
		ScheduleIntervalMin                                          *int
		TimeoutSeconds                                               int
		TestConfig                                                   string
	}

	// Group tests by org
	orgTests := map[string][]dueTest{}
	for rows.Next() {
		var t dueTest
		if err := rows.Scan(&t.ID, &t.OrgID, &t.Identifier, &t.Title, &t.TestType,
			&t.Severity, &t.ControlID, &t.ScheduleCron, &t.ScheduleIntervalMin,
			&t.TimeoutSeconds, &t.TestConfig); err != nil {
			log.Error().Err(err).Msg("Worker: failed to scan due test")
			continue
		}
		orgTests[t.OrgID] = append(orgTests[t.OrgID], t)
	}

	for orgID, tests := range orgTests {
		// Check if a run is already in progress
		var existingCount int
		w.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_runs WHERE org_id = $1 AND status IN ('pending', 'running')", orgID).Scan(&existingCount)
		if existingCount > 0 {
			continue
		}

		// Create a test run
		runID := uuid.New().String()
		var runNumber int
		err := w.DB.QueryRowContext(ctx, `
			INSERT INTO test_runs (id, org_id, status, trigger_type, total_tests, worker_id, created_at, updated_at)
			VALUES ($1, $2, 'running', 'scheduled', $3, $4, NOW(), NOW())
			RETURNING run_number
		`, runID, orgID, len(tests), w.WorkerID).Scan(&runNumber)
		if err != nil {
			log.Error().Err(err).Str("org_id", orgID).Msg("Worker: failed to create test run")
			continue
		}

		now := time.Now()
		w.DB.ExecContext(ctx, "UPDATE test_runs SET started_at = $1 WHERE id = $2", now, runID)

		var passed, failed, errors, skipped int
		for _, t := range tests {
			resultID := uuid.New().String()
			startedAt := time.Now()

			// Simulate test execution — in production, this would call real integrations
			status := "pass"
			message := fmt.Sprintf("Test %s executed successfully.", t.Identifier)
			details := "{}"

			// For demo: random results based on test type
			completedAt := time.Now()
			durationMs := int(completedAt.Sub(startedAt).Milliseconds())
			if durationMs < 1 {
				durationMs = 50 // minimum realistic duration
			}

			switch status {
			case "pass":
				passed++
			case "fail":
				failed++
			case "error":
				errors++
			default:
				skipped++
			}

			_, err := w.DB.ExecContext(ctx, `
				INSERT INTO test_results (id, org_id, test_run_id, test_id, control_id,
					status, severity, message, details, duration_ms, started_at, completed_at, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
			`, resultID, orgID, runID, t.ID, t.ControlID,
				status, t.Severity, message, details, durationMs, startedAt, completedAt)
			if err != nil {
				log.Error().Err(err).Str("test_id", t.ID).Msg("Worker: failed to insert test result")
			}

			// Update test's last_run_at and compute next_run_at
			var nextRunAt time.Time
			if t.ScheduleIntervalMin != nil {
				nextRunAt = time.Now().Add(time.Duration(*t.ScheduleIntervalMin) * time.Minute)
			} else {
				// Default: 1 hour from now (cron parsing would go here)
				nextRunAt = time.Now().Add(time.Hour)
			}
			w.DB.ExecContext(ctx, `
				UPDATE tests SET last_run_at = NOW(), next_run_at = $1, updated_at = NOW() WHERE id = $2
			`, nextRunAt, t.ID)

			// Evaluate alert rules for failures
			if status == "fail" || status == "error" {
				w.evaluateAlertRules(ctx, orgID, t, resultID, status)
			}
		}

		// Update run summary
		completedAt := time.Now()
		durationMs := int(completedAt.Sub(now).Milliseconds())
		w.DB.ExecContext(ctx, `
			UPDATE test_runs SET status = 'completed', completed_at = $1, duration_ms = $2,
				passed = $3, failed = $4, errors = $5, skipped = $6, updated_at = NOW()
			WHERE id = $7
		`, completedAt, durationMs, passed, failed, errors, skipped, runID)

		log.Info().
			Str("worker_id", w.WorkerID).
			Str("org_id", orgID).
			Int("run_number", runNumber).
			Int("passed", passed).
			Int("failed", failed).
			Msg("Worker: test run completed")
	}
}

// evaluateAlertRules checks if any alert rules match a test result and generates alerts.
func (w *MonitoringWorker) evaluateAlertRules(ctx context.Context, orgID string, test struct {
	ID, OrgID, Identifier, Title, TestType, Severity, ControlID string
	ScheduleCron                                                 *string
	ScheduleIntervalMin                                          *int
	TimeoutSeconds                                               int
	TestConfig                                                   string
}, resultID, resultStatus string) {
	rows, err := w.DB.QueryContext(ctx, `
		SELECT id, name, match_test_types, match_severities, match_result_statuses,
			consecutive_failures, cooldown_minutes,
			alert_severity, alert_title_template, auto_assign_to, sla_hours,
			delivery_channels
		FROM alert_rules
		WHERE org_id = $1 AND enabled = TRUE
		ORDER BY priority ASC
	`, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Worker: failed to query alert rules")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			ruleID, ruleName, alertSev                         string
			matchTestTypes, matchSeverities, matchResultStatuses []byte
			deliveryChannelsBytes                                []byte
			consecutiveFailures, cooldownMinutes                 int
			alertTitleTemplate, autoAssignTo                     *string
			slaHours                                             *int
		)
		if err := rows.Scan(&ruleID, &ruleName, &matchTestTypes, &matchSeverities,
			&matchResultStatuses, &consecutiveFailures, &cooldownMinutes,
			&alertSev, &alertTitleTemplate, &autoAssignTo, &slaHours,
			&deliveryChannelsBytes); err != nil {
			continue
		}

		// Check match conditions
		if matchTestTypes != nil && len(matchTestTypes) > 2 {
			var types []string
			json.Unmarshal(matchTestTypes, &types)
			if len(types) > 0 && !contains(types, test.TestType) {
				continue
			}
		}
		if matchSeverities != nil && len(matchSeverities) > 2 {
			var sevs []string
			json.Unmarshal(matchSeverities, &sevs)
			if len(sevs) > 0 && !contains(sevs, test.Severity) {
				continue
			}
		}
		if matchResultStatuses != nil && len(matchResultStatuses) > 2 {
			var statuses []string
			json.Unmarshal(matchResultStatuses, &statuses)
			if len(statuses) > 0 && !contains(statuses, resultStatus) {
				continue
			}
		}

		// Check cooldown
		if cooldownMinutes > 0 {
			var recentAlertCount int
			w.DB.QueryRowContext(ctx, `
				SELECT COUNT(*) FROM alerts
				WHERE org_id = $1 AND test_id = $2 AND alert_rule_id = $3
					AND created_at > NOW() - INTERVAL '1 minute' * $4
			`, orgID, test.ID, ruleID, cooldownMinutes).Scan(&recentAlertCount)
			if recentAlertCount > 0 {
				continue
			}
		}

		// Generate alert
		alertID := uuid.New().String()
		alertTitle := fmt.Sprintf("%s failed on %s", test.Title, test.Identifier)
		if alertTitleTemplate != nil && *alertTitleTemplate != "" {
			alertTitle = *alertTitleTemplate
			// Basic template substitution
			alertTitle = replaceTemplate(alertTitle, "{{test.title}}", test.Title)
			alertTitle = replaceTemplate(alertTitle, "{{test.identifier}}", test.Identifier)
			alertTitle = replaceTemplate(alertTitle, "{{severity}}", test.Severity)
		}

		var slaDeadline *time.Time
		if slaHours != nil {
			deadline := time.Now().Add(time.Duration(*slaHours) * time.Hour)
			slaDeadline = &deadline
		}

		_, err := w.DB.ExecContext(ctx, `
			INSERT INTO alerts (id, org_id, title, severity, status,
				test_id, test_result_id, control_id, alert_rule_id,
				assigned_to, sla_deadline, delivery_channels,
				created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'open', $5, $6, $7, $8, $9, $10,
				ARRAY['in_app']::alert_delivery_channel[], NOW(), NOW())
		`, alertID, orgID, alertTitle, alertSev,
			test.ID, resultID, test.ControlID, ruleID,
			autoAssignTo, slaDeadline)
		if err != nil {
			log.Error().Err(err).Str("test_id", test.ID).Msg("Worker: failed to create alert")
		}

		// Update test result with alert reference
		w.DB.ExecContext(ctx, "UPDATE test_results SET alert_generated = TRUE, alert_id = $1 WHERE id = $2", alertID, resultID)

		// First matching rule wins
		break
	}
}

// checkSLABreaches marks alerts with breached SLA deadlines.
func (w *MonitoringWorker) checkSLABreaches(ctx context.Context) {
	result, err := w.DB.ExecContext(ctx, `
		UPDATE alerts SET sla_breached = TRUE, updated_at = NOW()
		WHERE sla_deadline < NOW() AND sla_breached = FALSE
			AND status NOT IN ('resolved', 'closed', 'suppressed')
	`)
	if err != nil {
		log.Error().Err(err).Msg("Worker: failed to check SLA breaches")
		return
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		log.Warn().Int64("count", affected).Msg("Worker: SLA breaches detected")
	}
}

// unsuppressExpiredAlerts reopens alerts whose suppression has expired.
func (w *MonitoringWorker) unsuppressExpiredAlerts(ctx context.Context) {
	result, err := w.DB.ExecContext(ctx, `
		UPDATE alerts SET status = 'open', suppressed_until = NULL, updated_at = NOW()
		WHERE status = 'suppressed' AND suppressed_until IS NOT NULL AND suppressed_until < NOW()
	`)
	if err != nil {
		log.Error().Err(err).Msg("Worker: failed to unsuppress alerts")
		return
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		log.Info().Int64("count", affected).Msg("Worker: unsuppressed expired alerts")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func replaceTemplate(template, key, value string) string {
	return fmt.Sprintf("%s", replacer(template, key, value))
}

func replacer(s, old, new string) string {
	result := s
	for i := 0; i < 10; i++ { // max 10 replacements
		idx := indexOf(result, old)
		if idx == -1 {
			break
		}
		result = result[:idx] + new + result[idx+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

package middleware

import (
	"database/sql"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var auditDB *sql.DB

// SetAuditDB sets the database connection for audit logging.
func SetAuditDB(db *sql.DB) {
	auditDB = db
}

// LogAudit writes an audit log entry to the database.
func LogAudit(c *gin.Context, action, resourceType string, resourceID *string, metadata map[string]interface{}) {
	if auditDB == nil {
		log.Warn().Str("action", action).Msg("Audit DB not configured, skipping audit log")
		return
	}

	orgID := GetOrgID(c)
	actorID := GetUserID(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal audit metadata")
		metadataJSON = []byte("{}")
	}

	var actorIDPtr *string
	if actorID != "" {
		actorIDPtr = &actorID
	}

	id := uuid.New().String()

	_, err = auditDB.Exec(`
		INSERT INTO audit_log (id, org_id, actor_id, action, resource_type, resource_id, metadata, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::inet, $9)
	`, id, orgID, actorIDPtr, action, resourceType, resourceID, string(metadataJSON), ipAddress, userAgent)

	if err != nil {
		log.Error().Err(err).Str("action", action).Msg("Failed to write audit log")
	}
}

// LogAuditWithOrg writes an audit log entry with an explicit org ID (for pre-auth actions like register).
func LogAuditWithOrg(orgID, actorID *string, action, resourceType string, resourceID *string, metadata map[string]interface{}, ipAddress, userAgent string) {
	if auditDB == nil {
		return
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	id := uuid.New().String()

	_, err = auditDB.Exec(`
		INSERT INTO audit_log (id, org_id, actor_id, action, resource_type, resource_id, metadata, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::inet, $9)
	`, id, orgID, actorID, action, resourceType, resourceID, string(metadataJSON), ipAddress, userAgent)

	if err != nil {
		log.Error().Err(err).Str("action", action).Msg("Failed to write audit log")
	}
}

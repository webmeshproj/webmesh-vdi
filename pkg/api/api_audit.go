package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/auth/types"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// auditLogger handles the audit events. This could be overridden with audit
// "providers" that just implement the logr interface.
var auditLogger = logf.Log.WithName("api_audit")

// AuditResult contains information about an audit event from the API router.
type AuditResult struct {
	Allowed     bool
	FromOwner   bool
	Resource    string
	Grant       grants.RoleGrant
	UserSession *types.JWTClaims
	Request     *http.Request
}

// actions maps allowed values to display strings
var actions = map[bool]string{
	true:  "ALLOWED",
	false: "DENIED",
}

// buildAuditMsg builds a user-friendly audit message to pass to the logger.
func buildAuditMsg(result *AuditResult) string {
	msg := fmt.Sprintf(
		"%s %s => %s => %s",
		actions[result.Allowed],
		result.UserSession.User.Name,
		strings.Join(result.Grant.Names(), ","),
		result.Request.URL.Path,
	)
	if result.Resource != "" {
		msg = msg + fmt.Sprintf(" => %s", result.Resource)
	}
	if result.FromOwner {
		msg = msg + " (OWNER)"
	}
	return msg
}

// auditLog logs the event with parseable metadata.
func (d *desktopAPI) auditLog(result *AuditResult) {
	if !d.vdiCluster.AuditLogEnabled() {
		return
	}
	msg := buildAuditMsg(result)
	auditLogger.Info(
		msg,
		"Allowed", result.Allowed,
		"User.Name", result.UserSession.User.Name,
		"Request.Path", result.Request.URL.Path,
		"Grant.Names", result.Grant.Names(),
	)
}

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/util/grants"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var auditLogger = logf.Log.WithName("api_audit")

type AuditResult struct {
	Allowed     bool
	FromOwner   bool
	Grant       grants.RoleGrant
	UserSession *rethinkdb.UserSession
	Request     *http.Request
}

var actions = map[bool]string{
	true:  "ALLOWED",
	false: "DENIED",
}

func buildAuditMsg(result *AuditResult) string {
	msg := fmt.Sprintf(
		"%s %s => %s => %s",
		actions[result.Allowed],
		result.UserSession.User.Name,
		strings.Join(result.Grant.Names(), ","),
		result.Request.URL.Path,
	)
	if result.FromOwner {
		msg = msg + " (OWNER)"
	}
	return msg
}

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

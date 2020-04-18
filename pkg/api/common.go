package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"

	"github.com/gorilla/context"
)

const TokenHeader = "X-Session-Token"
const ContextUserKey = 0

func SetRequestUserSession(r *http.Request, sess *rethinkdb.UserSession) {
	context.Set(r, ContextUserKey, sess)
}

func GetRequestUserSession(r *http.Request) *rethinkdb.UserSession {
	return context.Get(r, ContextUserKey).(*rethinkdb.UserSession)
}

func (d *desktopAPI) WhoAmI(w http.ResponseWriter, r *http.Request) {
	session := GetRequestUserSession(r)
	apiutil.WriteJSON(session, w)
}

package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
	"k8s.io/apimachinery/pkg/types"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
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

// getNamespacedNameFromRequest returns the namespaced name of the Desktop instance
// for the given request.
func getNamespacedNameFromRequest(r *http.Request) types.NamespacedName {
	vars := mux.Vars(r)
	return types.NamespacedName{Name: vars["name"], Namespace: vars["namespace"]}
}

func getUserFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["user"]
}

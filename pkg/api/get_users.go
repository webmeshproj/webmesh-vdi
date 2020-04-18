package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/grants"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

func (d *desktopAPI) GetUsers(w http.ResponseWriter, r *http.Request) {
	if sess := GetRequestUserSession(r); sess == nil || !sess.User.HasGrant(grants.ReadUsers) {
		apiutil.ReturnAPIForbidden(nil, "User does not have ReadUsers grant", w)
		return
	}
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	users, err := sess.GetAllUsers()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(users, w)
}

func (d *desktopAPI) GetUser(w http.ResponseWriter, r *http.Request) {
	if sess := GetRequestUserSession(r); sess == nil || !sess.User.HasGrant(grants.ReadUsers) {
		apiutil.ReturnAPIForbidden(nil, "User does not have ReadUsers grant", w)
		return
	}
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	user, err := sess.GetUser(getUserFromRequest(r))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(user, w)
}

func getUserFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["user"]
}

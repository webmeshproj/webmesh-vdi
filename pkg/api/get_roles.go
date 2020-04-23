package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

func (d *desktopAPI) GetRoles(w http.ResponseWriter, r *http.Request) {
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	roles, err := sess.GetAllRoles()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(roles, w)
}

func (d *desktopAPI) GetRole(w http.ResponseWriter, r *http.Request) {
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	role, err := sess.GetRole(getRoleFromRequest(r))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(role, w)
}

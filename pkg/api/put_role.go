package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

func (d *desktopAPI) UpdateRole(w http.ResponseWriter, r *http.Request) {
	req := GetRequestObject(r).(*PostRoleRequest)
	roleName := getRoleFromRequest(r)
	role := &types.Role{
		Name:             roleName,
		Grants:           req.Grants,
		Namespaces:       req.Namespaces,
		TemplatePatterns: req.TemplatePatterns,
	}
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	if err := sess.UpdateRole(role); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

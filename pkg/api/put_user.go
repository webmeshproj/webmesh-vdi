package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

func (d *desktopAPI) UpdateUser(w http.ResponseWriter, r *http.Request) {
	req := GetRequestObject(r).(*PostUserRequest)
	userName := getUserFromRequest(r)
	user := &types.User{Name: userName}

	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()

	if req.Roles != nil && len(req.Roles) > 0 {
		user.Roles = make([]*types.Role, 0)
		for _, role := range req.Roles {
			user.Roles = append(user.Roles, &types.Role{Name: role})
		}
		if err := sess.UpdateUser(user); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
	}

	if req.Password != "" {
		if err := sess.SetUserPassword(user, req.Password); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
	}

	apiutil.WriteOK(w)
}

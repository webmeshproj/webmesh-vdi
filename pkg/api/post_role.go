package api

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	apierrors "github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

type PostRoleRequest struct {
	Name             string           `json:"name"`
	Grants           grants.RoleGrant `json:"grants"`
	Namespaces       []string         `json:"namespaces"`
	TemplatePatterns []string         `json:"templatePatterns"`
}

func (p *PostRoleRequest) Validate() error {
	if p.Name == "" || p.Grants == 0 {
		return errors.New("'name' and 'grants' must be provided in the request")
	}
	for _, x := range p.TemplatePatterns {
		if _, err := regexp.Compile(x); err != nil {
			return err
		}
	}
	return nil
}

func (d *desktopAPI) CreateRole(w http.ResponseWriter, r *http.Request) {
	req := GetRequestObject(r).(*PostRoleRequest)
	if err := req.Validate(); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	role := &types.Role{
		Name:             req.Name,
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
	if _, err := sess.GetRole(role.Name); err == nil {
		apiutil.ReturnAPIError(fmt.Errorf("A role with the name %s already exists", role.Name), w)
		return
	} else if !apierrors.IsRoleNotFoundError(err) {
		apiutil.ReturnAPIError(err, w)
		return
	}
	if err := sess.CreateRole(role); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

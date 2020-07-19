package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

func allowSameUser(d *desktopAPI, reqUser *v1.VDIUser, r *http.Request) (allowed, owner bool, err error) {
	pathUser := apiutil.GetUserFromRequest(r)
	if reqUser.Name != pathUser {
		return false, false, nil
	}
	// make sure the user isn't trying to change their permission level
	allowed, _, err = denyUserElevatePerms(d, reqUser, r)
	return allowed, true, err
}

func allowSessionOwner(d *desktopAPI, reqUser *v1.VDIUser, r *http.Request) (allowed, owner bool, err error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &v1alpha1.Desktop{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		return false, false, err
	}
	userDesktopLabels := d.vdiCluster.GetUserDesktopLabels(reqUser.Name)
	// extra safety check - cant accurately determine ownership without labels
	if found.GetLabels() == nil {
		return false, false, nil
	}
	for key, val := range userDesktopLabels {
		if expected, ok := found.GetLabels()[key]; !ok {
			return false, false, nil
		} else if expected != val {
			return false, false, nil
		}
	}
	return true, true, nil
}

func allowAll(d *desktopAPI, reqUser *v1.VDIUser, r *http.Request) (allowed, owner bool, err error) {
	return true, false, nil
}

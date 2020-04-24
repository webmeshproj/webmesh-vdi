package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

type AllowFunc func(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed, owner bool, err error)
type OverrideFunc func(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed bool, resource string, err error)

type MethodPermissions struct {
	RoleGrant    grants.RoleGrant
	AllowFunc    AllowFunc
	ResourceFunc OverrideFunc
}

func allowSameUser(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed, owner bool, err error) {
	pathUser := getUserFromRequest(r)
	if reqUser.Name != pathUser {
		return false, false, nil
	}
	reqUserRoles := reqUser.RoleNames()
	// make sure the user isn't trying to change their permission level
	if reqObj, ok := GetRequestObject(r).(*PostUserRequest); ok {
		if !reqUser.HasGrant(grants.WriteUsers) || !reqUser.HasGrant(grants.WriteRoles) {
			for _, role := range reqObj.Roles {
				if !util.StringSliceContains(reqUserRoles, role) {
					return false, false, nil
				}
			}
		}
	}
	return true, true, nil
}

func allowSessionOwner(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed, owner bool, err error) {
	nn := getNamespacedNameFromRequest(r)
	found := &v1alpha1.Desktop{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		return false, false, err
	}
	if !reflect.DeepEqual(found.GetLabels(), d.vdiCluster.GetUserDesktopLabels(reqUser.Name)) {
		return false, false, nil
	}
	return true, true, nil
}

func allowAll(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed, owner bool, err error) {
	return true, false, nil
}

func checkUserLaunchRestraints(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed bool, resource string, err error) {
	reqObj, ok := GetRequestObject(r).(*PostSessionsRequest)
	if !ok {
		return false, "Invalid", errors.New("PostSessionsRequest object is nil")
	}
	resourceName := fmt.Sprintf("%s/%s", reqObj.GetNamespace(), reqObj.GetTemplate())
	return reqUser.CanLaunch(reqObj.GetNamespace(), reqObj.GetTemplate()), resourceName, nil
}

var RouterGrantRequirements = map[string]map[string]MethodPermissions{
	"/api/whoami": {
		"GET": {
			AllowFunc: allowAll,
		},
	},
	"/api/logout": {
		"POST": {
			AllowFunc: allowAll,
		},
	},
	"/api/config": {
		"GET": {
			AllowFunc: allowAll,
		},
	},
	"/api/grants": {
		"GET": {
			AllowFunc: allowAll,
		},
	},
	"/api/namespaces": {
		"GET": {
			AllowFunc: allowAll,
		},
	},
	"/api/users": {
		"GET":  {RoleGrant: grants.ReadUsers},
		"POST": {RoleGrant: grants.WriteUsers},
	},
	"/api/users/{user}": {
		"GET": {
			RoleGrant: grants.ReadUsers,
			AllowFunc: allowSameUser,
		},
		"PUT": {
			RoleGrant: grants.WriteUsers,
			AllowFunc: allowSameUser,
		},
		"DELETE": {RoleGrant: grants.WriteUsers},
	},
	"/api/roles": {
		"GET":  {RoleGrant: grants.ReadRoles},
		"POST": {RoleGrant: grants.WriteRoles},
	},
	"/api/roles/{role}": {
		"GET":    {RoleGrant: grants.ReadRoles},
		"PUT":    {RoleGrant: grants.WriteRoles},
		"DELETE": {RoleGrant: grants.WriteRoles},
	},
	"/api/templates": {
		"GET": {RoleGrant: grants.ReadTemplates},
	},
	"/api/sessions": {
		"POST": {
			RoleGrant:    grants.LaunchTemplates,
			ResourceFunc: checkUserLaunchRestraints,
		},
	},
	"/api/sessions/{namespace}/{name}": {
		"GET": {
			RoleGrant: grants.ReadDesktopSessions,
			AllowFunc: allowSessionOwner,
		},
		"DELETE": {
			RoleGrant: grants.WriteDesktopSessions,
			AllowFunc: allowSessionOwner,
		},
	},
	"/api/websockify/{namespace}/{name}": {
		"GET": {
			RoleGrant: grants.UseDesktopSessions,
			AllowFunc: allowSessionOwner,
		},
	},
}

func (d *desktopAPI) ValidateUserGrants(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := &AuditResult{Request: r}

		// todo - can't use defer because websocket connections block and we won't
		// see the log until the session is over
		// // defer d.auditLog(result)

		// rertrieve the user and the path to match required grants
		userSession := GetRequestUserSession(r)
		result.UserSession = userSession

		path := getGorillaPath(r)

		// Safety checks should not fire when deployed, more to catch errors in testing
		grants, ok := RouterGrantRequirements[path]
		if !ok {
			apiutil.ReturnAPIForbidden(errors.New(path), "Could not determine required grants for route", w)
			result.Allowed = false
			d.auditLog(result)
			return
		}
		methodGrant, ok := grants[r.Method]
		if !ok {
			apiutil.ReturnAPIForbidden(errors.New(r.Method), "Could not determine required grants for method", w)
			result.Allowed = false
			d.auditLog(result)
			return
		}
		result.Grant = methodGrant.RoleGrant

		// Check if the route supports validating resource ownership
		if methodGrant.AllowFunc != nil {
			if allowed, owner, err := methodGrant.AllowFunc(d, userSession.User, r); err != nil {
				apiutil.ReturnAPIForbidden(err, "An error ocurred validating permission to the requested resource", w)
				result.Allowed = false
				d.auditLog(result)
				return
			} else if allowed {
				result.Allowed = true
				result.FromOwner = owner
				d.auditLog(result)
				next.ServeHTTP(w, r)
				return
			}
			// We were not allowed, but we may have a grant that lets us anyway
		}

		if !userSession.User.HasGrant(methodGrant.RoleGrant) {
			names := methodGrant.RoleGrant.Names()
			msg := fmt.Sprintf("%s does not have the %s grant", userSession.User.Name, strings.Join(names, ","))
			if len(names) > 1 {
				msg = msg + "s"
			}
			apiutil.ReturnAPIForbidden(nil, msg, w)
			result.Allowed = false
			d.auditLog(result)
			return
		}

		if methodGrant.ResourceFunc != nil {
			if allowed, resource, err := methodGrant.ResourceFunc(d, userSession.User, r); err != nil {
				apiutil.ReturnAPIForbidden(err, "An error ocurred validating permission to the requested resource", w)
				result.Allowed = false
				d.auditLog(result)
				return
			} else {
				result.Resource = resource
				if !allowed {
					result.Allowed = false
					names := methodGrant.RoleGrant.Names()
					msg := fmt.Sprintf("%s cannot %s on resource %s", userSession.User.Name, strings.Join(names, ","), resource)
					apiutil.ReturnAPIForbidden(err, msg, w)
					d.auditLog(result)
					return
				}
			}
		}

		result.Allowed = true
		d.auditLog(result)
		next.ServeHTTP(w, r)
	})
}

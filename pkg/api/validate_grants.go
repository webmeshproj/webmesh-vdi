package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/grants"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

type MethodPermissions struct {
	RoleGrant      grants.RoleGrant
	AllowOwnerFunc func(d *desktopAPI, reqUser *rethinkdb.User, r *http.Request) (allowed, owner bool, err error)
}

func allowSameUser(d *desktopAPI, reqUser *rethinkdb.User, r *http.Request) (allowed, owner bool, err error) {
	user := getUserFromRequest(r)
	if reqUser.Name == user {
		return true, true, nil
	}
	return false, false, nil
}

func allowSessionOwner(d *desktopAPI, reqUser *rethinkdb.User, r *http.Request) (allowed, owner bool, err error) {
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

func allowAll(d *desktopAPI, reqUser *rethinkdb.User, r *http.Request) (allowed, owner bool, err error) {
	return true, false, nil
}

var RouterGrantRequirements = map[string]map[string]MethodPermissions{
	"/api/whoami": {
		"GET": {
			AllowOwnerFunc: allowAll,
		},
	},
	"/api/logout": {
		"POST": {
			AllowOwnerFunc: allowAll,
		},
	},
	"/api/config": {
		"GET": {
			AllowOwnerFunc: allowAll,
		},
	},
	"/api/grants": {
		"GET": {
			AllowOwnerFunc: allowAll,
		},
	},
	"/api/users": {
		"GET":  {RoleGrant: grants.ReadUsers},
		"POST": {RoleGrant: grants.WriteUsers},
	},
	"/api/users/{user}": {
		"GET": {
			RoleGrant:      grants.ReadUsers,
			AllowOwnerFunc: allowSameUser,
		},
		"PUT": {
			RoleGrant:      grants.WriteUsers,
			AllowOwnerFunc: allowSameUser,
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
		"POST": {RoleGrant: grants.LaunchTemplates},
	},
	"/api/sessions/{namespace}/{name}": {
		"GET": {
			RoleGrant:      grants.ReadDesktopSessions,
			AllowOwnerFunc: allowSessionOwner,
		},
		"DELETE": {
			RoleGrant:      grants.WriteDesktopSessions,
			AllowOwnerFunc: allowSessionOwner,
		},
	},
	"/api/websockify/{namespace}/{name}": {
		"GET": {
			RoleGrant:      grants.UseDesktopSessions,
			AllowOwnerFunc: allowSessionOwner,
		},
	},
}

func getGorillaPath(r *http.Request) string {
	vars := mux.Vars(r)
	path := strings.TrimSuffix(r.URL.Path, "/")
	for k, v := range vars {
		path = strings.Replace(path, v, fmt.Sprintf("{%s}", k), 1)
	}
	return path
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
		if methodGrant.AllowOwnerFunc != nil {
			if allowed, owner, err := methodGrant.AllowOwnerFunc(d, userSession.User, r); err != nil {
				apiutil.ReturnAPIForbidden(err, "An error ocurred validating ownership of the requested resource", w)
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

		result.Allowed = true
		d.auditLog(result)
		next.ServeHTTP(w, r)
	})
}

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

type AllowFunc func(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed, owner bool, err error)
type ResourceFunc func(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed bool, resource, reason string, err error)

type MethodPermissions struct {
	RoleGrant    grants.RoleGrant
	AllowFunc    AllowFunc
	ResourceFunc ResourceFunc
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
		"GET": {RoleGrant: grants.ReadUsers},
		"POST": {
			RoleGrant:    grants.WriteUsers,
			ResourceFunc: denyUserElevatePerms,
		},
	},
	"/api/users/{user}": {
		"GET": {
			RoleGrant: grants.ReadUsers,
			AllowFunc: allowSameUser,
		},
		"PUT": {
			RoleGrant:    grants.WriteUsers,
			AllowFunc:    allowSameUser,
			ResourceFunc: denyUserElevatePerms,
		},
		"DELETE": {RoleGrant: grants.WriteUsers},
	},
	"/api/roles": {
		"GET": {RoleGrant: grants.ReadRoles},
		"POST": {
			RoleGrant:    grants.WriteRoles,
			ResourceFunc: denyUserElevatePerms,
		},
	},
	"/api/roles/{role}": {
		"GET": {RoleGrant: grants.ReadRoles},
		"PUT": {
			RoleGrant:    grants.WriteRoles,
			ResourceFunc: denyUserElevatePerms,
		},
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
			if allowed, resource, reason, err := methodGrant.ResourceFunc(d, userSession.User, r); err != nil {
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
					if reason != "" {
						msg = msg + fmt.Sprintf(". %s", reason)
					}
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

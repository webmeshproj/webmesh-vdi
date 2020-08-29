package api

import (
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// OverrideFunc is a function that takes precedence over any other action evaluations.
// If it returns false for allowed, the next rules in the chain will be considered.
// Errors are considered forbidden.
type OverrideFunc func(d *desktopAPI, reqUser *v1.VDIUser, r *http.Request) (allowed, owner bool, err error)

// ExtraCheckFunc is a function that fires after the action itself has been evaluated.
// Allowed being false or any errors are considered forbidden.
type ExtraCheckFunc func(d *desktopAPI, reqUser *v1.VDIUser, r *http.Request) (allowed bool, reason string, err error)

// ResourceValueFunc returns the name of a requested resource based off the contents
// of a request.
type ResourceValueFunc func(r *http.Request) (name string)

// MethodPermissions represents a set of checks to run for an API method.
type MethodPermissions struct {
	OverrideFunc          OverrideFunc
	Actions               []v1.APIAction
	ResourceNameFunc      ResourceValueFunc
	ResourceNamespaceFunc ResourceValueFunc
	ExtraCheckFunc        ExtraCheckFunc
}

// RouterGrantRequirements defines all the methods that are protected, and what
// rules should be evaluated for them.
var RouterGrantRequirements = map[string]map[string]MethodPermissions{
	"/api/whoami": {
		"GET": {
			OverrideFunc: allowAll,
		},
	},
	"/api/authorize": {
		"POST": {
			OverrideFunc: allowAll,
		},
	},
	"/api/logout": {
		"POST": {
			OverrideFunc: allowAll,
		},
	},
	"/api/config": {
		"GET": {
			OverrideFunc: allowAll,
		},
	},
	"/api/config/reload": {
		"POST": {
			OverrideFunc: allowAll,
		},
	},
	"/api/namespaces": {
		"GET": {
			OverrideFunc: allowAll,
		},
	},
	"/api/users": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceUsers,
				},
			},
		},
		"POST": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbCreate,
					ResourceType: v1.ResourceUsers,
				},
			},
			ExtraCheckFunc: denyUserElevatePerms,
		},
	},
	"/api/users/{user}": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceUsers,
				},
			},
			ResourceNameFunc: apiutil.GetUserFromRequest,
			OverrideFunc:     allowSameUser,
		},
		"PUT": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUpdate,
					ResourceType: v1.ResourceUsers,
				},
			},
			ResourceNameFunc: apiutil.GetUserFromRequest,
			OverrideFunc:     allowSameUser,
			ExtraCheckFunc:   denyUserElevatePerms,
		},
		"DELETE": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbDelete,
					ResourceType: v1.ResourceUsers,
				},
			},
			ResourceNameFunc: apiutil.GetUserFromRequest,
		},
	},
	"/api/users/{user}/mfa": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceUsers,
				},
			},
			ResourceNameFunc: apiutil.GetUserFromRequest,
			OverrideFunc:     allowSameUser,
		},
		"PUT": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUpdate,
					ResourceType: v1.ResourceUsers,
				},
			},
			ResourceNameFunc: apiutil.GetUserFromRequest,
			OverrideFunc:     allowSameUser,
		},
	},
	"/api/users/{user}/mfa/verify": {
		"PUT": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUpdate,
					ResourceType: v1.ResourceUsers,
				},
			},
			ResourceNameFunc: apiutil.GetUserFromRequest,
			OverrideFunc:     allowSameUser,
		},
	},
	"/api/roles": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceRoles,
				},
			},
		},
		"POST": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbCreate,
					ResourceType: v1.ResourceRoles,
				},
			},
			ExtraCheckFunc: denyUserElevatePerms,
		},
	},
	"/api/roles/{role}": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceRoles,
				},
			},
			ResourceNameFunc: apiutil.GetRoleFromRequest,
		},
		"PUT": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUpdate,
					ResourceType: v1.ResourceRoles,
				},
			},
			ResourceNameFunc: apiutil.GetRoleFromRequest,
			ExtraCheckFunc:   denyUserElevatePerms,
		},
		"DELETE": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbDelete,
					ResourceType: v1.ResourceRoles,
				},
			},
			ResourceNameFunc: apiutil.GetRoleFromRequest,
		},
	},
	"/api/templates": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceTemplates,
				},
			},
		},
		"POST": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbCreate,
					ResourceType: v1.ResourceTemplates,
				},
			},
		},
	},
	"/api/templates/{template}": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc: apiutil.GetTemplateFromRequest,
		},
		"PUT": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUpdate,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc: apiutil.GetTemplateFromRequest,
		},
		"DELETE": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbDelete,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc: apiutil.GetTemplateFromRequest,
		},
	},
	"/api/sessions": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceTemplates,
				},
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceUsers,
				},
			},
		},
		"POST": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbLaunch,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc: func(r *http.Request) string {
				req := apiutil.GetRequestObject(r).(*v1.CreateSessionRequest)
				return req.GetTemplate()
			},
			ResourceNamespaceFunc: func(r *http.Request) string {
				req := apiutil.GetRequestObject(r).(*v1.CreateSessionRequest)
				return req.GetNamespace()
			},
		},
	},
	"/api/sessions/{namespace}/{name}": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbRead,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc:      apiutil.GetNameFromRequest,
			ResourceNamespaceFunc: apiutil.GetNamespaceFromRequest,
			OverrideFunc:          allowSessionOwner,
		},
		"DELETE": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbDelete,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc:      apiutil.GetNameFromRequest,
			ResourceNamespaceFunc: apiutil.GetNamespaceFromRequest,
			OverrideFunc:          allowSessionOwner,
		},
	},
	"/api/desktops/websockify/{namespace}/{name}": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUse,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc:      apiutil.GetNameFromRequest,
			ResourceNamespaceFunc: apiutil.GetNamespaceFromRequest,
			OverrideFunc:          allowSessionOwner,
		},
	},
	"/api/desktops/wsaudio/{namespace}/{name}": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUse,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc:      apiutil.GetNameFromRequest,
			ResourceNamespaceFunc: apiutil.GetNamespaceFromRequest,
			OverrideFunc:          allowSessionOwner,
		},
	},
	"/api/desktops/fs/{namespace}/{name}/stat/": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUse,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc:      apiutil.GetNameFromRequest,
			ResourceNamespaceFunc: apiutil.GetNamespaceFromRequest,
			OverrideFunc:          allowSessionOwner,
		},
	},
	"/api/desktops/fs/{namespace}/{name}/get/": {
		"GET": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUse,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc:      apiutil.GetNameFromRequest,
			ResourceNamespaceFunc: apiutil.GetNamespaceFromRequest,
			OverrideFunc:          allowSessionOwner,
		},
	},
	"/api/desktops/fs/{namespace}/{name}/put": {
		"PUT": {
			Actions: []v1.APIAction{
				{
					Verb:         v1.VerbUse,
					ResourceType: v1.ResourceTemplates,
				},
			},
			ResourceNameFunc:      apiutil.GetNameFromRequest,
			ResourceNamespaceFunc: apiutil.GetNamespaceFromRequest,
			OverrideFunc:          allowSessionOwner,
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
		userSession := apiutil.GetRequestUserSession(r)
		result.UserSession = userSession

		path := apiutil.GetGorillaPath(r)

		// Safety checks should not fire when deployed, more to catch errors in testing
		grants, ok := RouterGrantRequirements[path]
		if !ok {
			apiutil.ReturnAPIForbidden(errors.New(path), "Could not determine required grants for route", w)
			return
		}
		methodGrant, ok := grants[r.Method]
		if !ok {
			apiutil.ReturnAPIForbidden(errors.New(r.Method), "Could not determine required grants for method", w)
			return
		}

		// Check if the route supports validating resource ownership
		if methodGrant.OverrideFunc != nil {
			if allowed, owner, err := methodGrant.OverrideFunc(d, userSession.User, r); err != nil {
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

		for _, action := range methodGrant.Actions {
			apiAction := buildActionFromTemplate(methodGrant, action, r)
			result.Actions = append(result.Actions, apiAction)
			if !userSession.User.Evaluate(apiAction) {
				msg := fmt.Sprintf("%s does not have the ability to %s", userSession.User.Name, apiAction.String())
				apiutil.ReturnAPIForbidden(nil, msg, w)
				result.Allowed = false
				d.auditLog(result)
				return
			}
		}

		if methodGrant.ExtraCheckFunc != nil {
			allowed, reason, err := methodGrant.ExtraCheckFunc(d, userSession.User, r)
			if err != nil {
				apiutil.ReturnAPIForbidden(err, "An error ocurred checking extra restraints on the resource", w)
				result.Allowed = false
				d.auditLog(result)
				return
			}
			if !allowed {
				msg := fmt.Sprintf("%s denied access to %s: %s", userSession.User.Name, r.URL.Path, reason)
				apiutil.ReturnAPIForbidden(nil, msg, w)
				result.Allowed = false
				d.auditLog(result)
				return
			}
		}

		result.Allowed = true
		d.auditLog(result)
		next.ServeHTTP(w, r)
	})
}

// buildActionFromTemplate will create an APIAction to evaluate based off the
// parameters in the MethodPermissions.
func buildActionFromTemplate(perms MethodPermissions, action v1.APIAction, r *http.Request) *v1.APIAction {
	// build a new action object
	tmplAction := &v1.APIAction{
		Verb:         action.Verb,
		ResourceType: action.ResourceType,
	}

	// populate the name if possible
	if perms.ResourceNameFunc != nil {
		tmplAction.ResourceName = perms.ResourceNameFunc(r)
	}

	// populate the namespace if possible
	if perms.ResourceNamespaceFunc != nil {
		tmplAction.ResourceNamespace = perms.ResourceNamespaceFunc(r)
	}

	return tmplAction
}

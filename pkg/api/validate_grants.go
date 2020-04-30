package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// OverrideFunc is a function that takes precedence over any other action evaluations.
// If it returns false for allowed, the next rules in the chain will be considered.
// Errors are considered forbidden.
type OverrideFunc func(d *desktopAPI, reqUser *v1alpha1.VDIUser, r *http.Request) (allowed, owner bool, err error)

// ExtraCheckFunc is a function that fires after the action itself has been evaluated.
// Allowed being false or any errors are considered forbidden.
type ExtraCheckFunc func(d *desktopAPI, reqUser *v1alpha1.VDIUser, r *http.Request) (allowed bool, reason string, err error)

// ResourceNameFunc returns the name of a requested resource based off the contents
// of a request.
type ResourceValueFunc func(r *http.Request) (name string)

// MethodPermissions represents a set of checks to run for an API method.
type MethodPermissions struct {
	OverrideFunc          OverrideFunc
	Action                v1alpha1.APIAction
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
	"/api/namespaces": {
		"GET": {
			OverrideFunc: allowAll,
		},
	},
	"/api/users": {
		"GET": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbRead,
				ResourceType: v1alpha1.ResourceUsers,
			},
		},
		"POST": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbCreate,
				ResourceType: v1alpha1.ResourceUsers,
			},
			ExtraCheckFunc: denyUserElevatePerms,
		},
	},
	"/api/users/{user}": {
		"GET": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbRead,
				ResourceType: v1alpha1.ResourceUsers,
			},
			ResourceNameFunc: func(r *http.Request) string { return mux.Vars(r)["user"] },
			OverrideFunc:     allowSameUser,
		},
		"PUT": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbUpdate,
				ResourceType: v1alpha1.ResourceUsers,
			},
			ResourceNameFunc: func(r *http.Request) string { return mux.Vars(r)["user"] },
			OverrideFunc:     allowSameUser,
			ExtraCheckFunc:   denyUserElevatePerms,
		},
		"DELETE": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbDelete,
				ResourceType: v1alpha1.ResourceUsers,
			},
			ResourceNameFunc: func(r *http.Request) string { return mux.Vars(r)["user"] },
		},
	},
	"/api/roles": {
		"GET": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbRead,
				ResourceType: v1alpha1.ResourceRoles,
			},
		},
		"POST": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbCreate,
				ResourceType: v1alpha1.ResourceRoles,
			},
			ExtraCheckFunc: denyUserElevatePerms,
		},
	},
	"/api/roles/{role}": {
		"GET": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbRead,
				ResourceType: v1alpha1.ResourceRoles,
			},
			ResourceNameFunc: func(r *http.Request) string { return mux.Vars(r)["role"] },
		},
		"PUT": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbUpdate,
				ResourceType: v1alpha1.ResourceRoles,
			},
			ResourceNameFunc: func(r *http.Request) string { return mux.Vars(r)["role"] },
			ExtraCheckFunc:   denyUserElevatePerms,
		},
		"DELETE": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbDelete,
				ResourceType: v1alpha1.ResourceRoles,
			},
			ResourceNameFunc: func(r *http.Request) string { return mux.Vars(r)["role"] },
		},
	},
	"/api/templates": {
		"GET": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbRead,
				ResourceType: v1alpha1.ResourceTemplates,
			},
		},
	},
	"/api/sessions": {
		"POST": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbLaunch,
				ResourceType: v1alpha1.ResourceTemplates,
			},
			ResourceNameFunc: func(r *http.Request) string {
				req := apiutil.GetRequestObject(r).(*v1alpha1.CreateSessionRequest)
				return req.GetTemplate()
			},
			ResourceNamespaceFunc: func(r *http.Request) string {
				req := apiutil.GetRequestObject(r).(*v1alpha1.CreateSessionRequest)
				return req.GetNamespace()
			},
		},
	},
	"/api/sessions/{namespace}/{name}": {
		"GET": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbRead,
				ResourceType: v1alpha1.ResourceTemplates,
			},
			ResourceNameFunc: func(r *http.Request) string {
				return mux.Vars(r)["name"]
			},
			ResourceNamespaceFunc: func(r *http.Request) string {
				return mux.Vars(r)["namespace"]
			},
			OverrideFunc: allowSessionOwner,
		},
		"DELETE": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbDelete,
				ResourceType: v1alpha1.ResourceTemplates,
			},
			ResourceNameFunc: func(r *http.Request) string {
				return mux.Vars(r)["name"]
			},
			ResourceNamespaceFunc: func(r *http.Request) string {
				return mux.Vars(r)["namespace"]
			},
			OverrideFunc: allowSessionOwner,
		},
	},
	"/api/websockify/{namespace}/{name}": {
		"GET": {
			Action: v1alpha1.APIAction{
				Verb:         v1alpha1.VerbUse,
				ResourceType: v1alpha1.ResourceTemplates,
			},
			ResourceNameFunc: func(r *http.Request) string {
				return fmt.Sprintf("%s/%s", mux.Vars(r)["namespace"], mux.Vars(r)["name"])
			},
			OverrideFunc: allowSessionOwner,
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

		apiAction := buildActionFromTemplate(methodGrant, r)
		result.Action = apiAction

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

		if !userSession.User.Evaluate(apiAction) {
			msg := fmt.Sprintf("%s does not have the ability to %s", userSession.User.Name, apiAction.String())
			apiutil.ReturnAPIForbidden(nil, msg, w)
			result.Allowed = false
			d.auditLog(result)
			return
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
				msg := fmt.Sprintf("%s does not have the ability to %s: %s", userSession.User.Name, apiAction.String(), reason)
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
func buildActionFromTemplate(perms MethodPermissions, r *http.Request) *v1alpha1.APIAction {
	// build a new action object
	action := &v1alpha1.APIAction{
		Verb:         perms.Action.Verb,
		ResourceType: perms.Action.ResourceType,
	}

	// populate the name if possible
	if perms.ResourceNameFunc != nil {
		action.ResourceName = perms.ResourceNameFunc(r)
	}

	// populate the namespace if possible
	if perms.ResourceNamespaceFunc != nil {
		action.ResourceNamespace = perms.ResourceNamespaceFunc(r)
	}

	return action
}

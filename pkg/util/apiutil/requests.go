package apiutil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/types"
)

// ContextUserKey is the key where user sessions are stored in the request context
const ContextUserKey = 0

// ContextRequestObjectKey is the key where decoded request objects are stored
// in the request context
const ContextRequestObjectKey = 1

// SetRequestUserSession writes the user session to the request context
func SetRequestUserSession(r *http.Request, sess *v1.JWTClaims) {
	context.Set(r, ContextUserKey, sess)
}

// GetRequestUserSession retrieves the user session from the request context.
func GetRequestUserSession(r *http.Request) *v1.JWTClaims {
	return context.Get(r, ContextUserKey).(*v1.JWTClaims)
}

// SetRequestObject sets the given interface to the decoded request object in the context.
func SetRequestObject(r *http.Request, obj interface{}) {
	context.Set(r, ContextRequestObjectKey, obj)
}

// GetRequestObject retrieves the decoded request from the request context.
func GetRequestObject(r *http.Request) interface{} {
	return context.Get(r, ContextRequestObjectKey)
}

// GetNamespacedNameFromRequest returns the namespaced name of the Desktop instance
// for the given request.
func GetNamespacedNameFromRequest(r *http.Request) types.NamespacedName {
	vars := mux.Vars(r)
	return types.NamespacedName{Name: vars["name"], Namespace: vars["namespace"]}
}

// GetUserFromRequest will retrieve the user variable from a request path.
func GetUserFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["user"]
}

// GetRoleFromRequest will retrieve the role variable from a request path.
func GetRoleFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["role"]
}

// GetTemplateFromRequest will retrieve the template variable from a request path.
func GetTemplateFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["template"]
}

// GetGorillaPath will retrieve the URL path as it was configured in mux.
func GetGorillaPath(r *http.Request) string {
	vars := mux.Vars(r)
	path := strings.TrimSuffix(r.URL.Path, "/")
	for k, v := range vars {
		path = rev(strings.Replace(rev(path), rev(v), rev(fmt.Sprintf("{%s}", k)), 1))
	}
	return path
}

// rev will reverse a string so we can call strings.Replace from the end
func rev(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

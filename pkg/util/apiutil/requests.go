/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package apiutil

import (
	"net/http"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

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

func getRequestVar(r *http.Request, name string) string {
	vars := mux.Vars(r)
	return vars[name]
}

// GetNameFromRequest returns the name of the Desktop instance for the given request.
func GetNameFromRequest(r *http.Request) string {
	return getRequestVar(r, "name")
}

// GetNamespaceFromRequest returns the namespace of the Desktop instance for the given
// request.
func GetNamespaceFromRequest(r *http.Request) string {
	return getRequestVar(r, "namespace")
}

// GetContainerFromRequest returns the container inside the Desktop instance for the given
// request.
func GetContainerFromRequest(r *http.Request) string {
	return getRequestVar(r, "container")
}

// GetNamespacedNameFromRequest returns the namespaced name of the Desktop instance
// for the given request.
func GetNamespacedNameFromRequest(r *http.Request) types.NamespacedName {
	return types.NamespacedName{
		Name:      GetNameFromRequest(r),
		Namespace: GetNamespaceFromRequest(r),
	}
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
	rt := mux.CurrentRoute(r)
	path, _ := rt.GetPathTemplate()
	return path
}

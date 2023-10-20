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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/errors"
	rbacutil "github.com/kvdi/kvdi/pkg/util/rbac"
)

// WriteOrLogError will write the provided content to the response writer, or
// log any error. It assumes the content is valid JSON.
func WriteOrLogError(out []byte, w http.ResponseWriter, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(append(out, []byte("\n")...)); err != nil {
		log.Println("Failed to write API response:", string(out), "error", err)
	}
}

// ReturnAPIErrors returns a BadRequest status code with a json encoded list
// of errors.
func ReturnAPIErrors(errs []error, w http.ResponseWriter) {
	out := make([]string, 0)
	for _, err := range errs {
		out = append(out, err.Error())
	}
	jout, err := json.Marshal(map[string][]string{
		"errors": out,
	})
	if err != nil {
		log.Println("Failed to marshal errors to json:", err)
		jout = []byte(`{"error": "Multiple errors happened while processing the request"}`)
	}
	WriteOrLogError(jout, w, http.StatusBadRequest)
}

// ReturnAPIError returns a BadRequest status code with a json encoded error
// message.
func ReturnAPIError(err error, w http.ResponseWriter) {
	WriteOrLogError(errors.ToAPIError(err, errors.ServerError).JSON(), w, http.StatusBadRequest)
}

// ReturnAPINotFound returns a NotFound status code with a json encoded error
// message.
func ReturnAPINotFound(err error, w http.ResponseWriter) {
	WriteOrLogError(errors.ToAPIError(err, errors.NotFound).JSON(), w, http.StatusNotFound)
}

// ReturnAPIForbidden returns a Forbidden status code with a json encoded error
// message. If the denial happened due to an error, it logs the error server side.
func ReturnAPIForbidden(err error, msg string, w http.ResponseWriter) {
	if err != nil {
		log.Println("Forbidden request due to:", err.Error())
	}
	WriteOrLogError(errors.ToAPIError(fmt.Errorf("forbidden: %s", msg), errors.Forbidden).JSON(), w, http.StatusForbidden)
}

// ReturnAPIUnauthorized returns an Unauthorized status with a json encoded error message.
// If the denial happened due to an error, it is logged server side.
func ReturnAPIUnauthorized(err error, msg string, w http.ResponseWriter) {
	if err != nil {
		log.Println("Unauthorized request due to:", err.Error())
	}
	WriteOrLogError(errors.ToAPIError(fmt.Errorf("unauthorized: %s", msg), errors.Unauthorized).JSON(), w, http.StatusUnauthorized)
}

// WriteJSON encodes the provided interface to JSON and writes it to the response
// stream.
func WriteJSON(i interface{}, w http.ResponseWriter) {
	out, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		ReturnAPIError(err, w)
		return
	}
	WriteOrLogError(out, w, http.StatusOK)
}

// UnmarshalRequest will read the body of the given request and decode it into
// the given interface.
func UnmarshalRequest(r *http.Request, in interface{}) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, in)
}

// WriteOK write a simple boolean okay response.
func WriteOK(w http.ResponseWriter) {
	WriteJSON(map[string]bool{
		"ok": true,
	}, w)
}

// FilterUserRolesByNames returns a list of UserRoles matching the provided names
// and cluster roles.
func FilterUserRolesByNames(roles []*rbacv1.VDIRole, names []string) []*types.VDIUserRole {
	userRoles := make([]*types.VDIUserRole, 0)
	for _, name := range names {
		for _, role := range roles {
			if role.GetName() == name {
				userRoles = append(userRoles, rbacutil.VDIRoleToUserRole(role))
			}
		}
	}
	return userRoles
}

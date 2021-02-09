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

package api

import (
	"net/http"
	"reflect"

	desktopsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	"github.com/tinyzimmer/kvdi/pkg/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// Decoders is a map of request paths/methods to the request object that
// should be used for deserialization.
var Decoders = map[string]map[string]interface{}{
	"/api/authorize": {
		"POST": types.AuthorizeRequest{},
	},
	"/api/sessions": {
		"POST": types.CreateSessionRequest{},
	},
	"/api/users": {
		"POST": types.CreateUserRequest{},
	},
	"/api/users/{user}": {
		"PUT": types.UpdateUserRequest{},
	},
	"/api/users/{user}/mfa": {
		"PUT": types.UpdateMFARequest{},
	},
	"/api/users/{user}/mfa/verify": {
		"PUT": types.AuthorizeRequest{},
	},
	"/api/roles": {
		"POST": types.CreateRoleRequest{},
	},
	"/api/templates": {
		"POST": desktopsv1.Template{},
	},
	"/api/roles/{role}": {
		"PUT": types.UpdateRoleRequest{},
	},
	"/api/login": {
		"POST": types.LoginRequest{},
	},
}

// DecodeRequest will inspect the request object for the type of object
// to deserialize the request to, and then apply the object to the request context.
func DecodeRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := apiutil.GetGorillaPath(r)
		if decoder, ok := Decoders[path]; ok {
			if decoderType, ok := decoder[r.Method]; ok {
				req, err := reflectAndDecodeRequest(r, decoderType)
				if err != nil {
					apiutil.ReturnAPIError(err, w)
					return
				}
				apiutil.SetRequestObject(r, req)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func reflectAndDecodeRequest(r *http.Request, t interface{}) (interface{}, error) {
	rType := reflect.TypeOf(t)
	req := reflect.New(rType).Interface()
	if err := apiutil.UnmarshalRequest(r, req); err != nil {
		return nil, err
	}
	if validator, ok := req.(interface{ Validate() error }); ok {
		return req, validator.Validate()
	}
	return req, nil
}

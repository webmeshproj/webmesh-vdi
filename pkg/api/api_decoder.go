package api

import (
	"net/http"
	"reflect"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// Decoders is a map of request paths/methods to the request object that
// should be used for deserialization.
var Decoders = map[string]map[string]interface{}{
	"/api/authorize": {
		"POST": v1.AuthorizeRequest{},
	},
	"/api/sessions": {
		"POST": v1.CreateSessionRequest{},
	},
	"/api/users": {
		"POST": v1.CreateUserRequest{},
	},
	"/api/users/{user}": {
		"PUT": v1.UpdateUserRequest{},
	},
	"/api/users/{user}/mfa": {
		"PUT": v1.UpdateMFARequest{},
	},
	"/api/users/{user}/mfa/verify": {
		"PUT": v1.AuthorizeRequest{},
	},
	"/api/roles": {
		"POST": v1.CreateRoleRequest{},
	},
	"/api/templates": {
		"POST": v1alpha1.DesktopTemplate{},
	},
	"/api/roles/{role}": {
		"PUT": v1.UpdateRoleRequest{},
	},
	"/api/login": {
		"POST": v1.LoginRequest{},
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

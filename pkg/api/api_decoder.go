package api

import (
	"net/http"
	"reflect"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

var Decoders = map[string]map[string]interface{}{
	"/api/authorize": {
		"POST": v1alpha1.AuthorizeRequest{},
	},
	"/api/sessions": {
		"POST": v1alpha1.CreateSessionRequest{},
	},
	"/api/users": {
		"POST": v1alpha1.CreateUserRequest{},
	},
	"/api/users/{user}": {
		"PUT": v1alpha1.UpdateUserRequest{},
	},
	"/api/users/{user}/mfa": {
		"PUT": v1alpha1.UpdateMFARequest{},
	},
	"/api/roles": {
		"POST": v1alpha1.CreateRoleRequest{},
	},
	"/api/roles/{role}": {
		"PUT": v1alpha1.UpdateRoleRequest{},
	},
	"/api/login": {
		"POST": v1alpha1.LoginRequest{},
	},
}

func DecodeRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := apiutil.GetGorillaPath(r)
		if decoder, ok := Decoders[path]; ok {
			if decoderType, ok := decoder[r.Method]; ok {
				req, err := decodeRequest(r, decoderType)
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

func decodeRequest(r *http.Request, t interface{}) (interface{}, error) {
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

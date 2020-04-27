package api

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
)

func (d *desktopAPI) ValidateUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get(TokenHeader)
		if authToken == "" {
			if keys, ok := r.URL.Query()["token"]; ok {
				authToken = keys[0]
			}
		}
		if authToken == "" {
			apiutil.ReturnAPIForbidden(nil, "No token provided in request", w)
			return
		}
		parser := &jwt.Parser{UseJSONNumber: true}
		token, err := parser.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Incorrect signing algorithm on token")
			}
			_, key := tlsutil.ServerKeypair()
			return ioutil.ReadFile(key)
		})
		if !token.Valid {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					apiutil.ReturnAPIForbidden(nil, "Malformed token provided in request", w)
					return
				} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
					apiutil.ReturnAPIForbidden(nil, "User session has expired", w)
					return
				} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
					apiutil.ReturnAPIForbidden(nil, "User session is not valid yet", w)
					return
				} else {
					apiutil.ReturnAPIForbidden(nil, "Could not parse provided token", w)
					return
				}
			}
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			session := &types.JWTClaims{}
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  session,
			})
			if err != nil {
				apiutil.ReturnAPIError(err, w)
				return
			}
			if err := decoder.Decode(claims); err != nil {
				apiutil.ReturnAPIError(err, w)
				return
			}
			SetRequestUserSession(r, session)
			next.ServeHTTP(w, r)
			return
		}
		apiutil.ReturnAPIError(errors.New("Could not parse provided token"), w)
		return
	})
}

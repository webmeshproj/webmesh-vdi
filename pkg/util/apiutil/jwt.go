package apiutil

import (
	"fmt"
	"time"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
)

// GenerateJWT will create a new JWT with the given user object's fields
// embedded in the claims.
func GenerateJWT(secret []byte, authResult *v1.AuthResult, authorized bool, sessionLength time.Duration) (v1.JWTClaims, string, error) {
	claims := v1.JWTClaims{
		User:       authResult.User,
		Data:       authResult.Data,
		Authorized: authorized,
		Renewable:  !authResult.RefreshNotSupported,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(sessionLength).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	return claims, tokenString, err
}

// Token verification errors
var errTokenMalformedError = errors.New("Malformed token provided in the request")
var errTokenNotValidYetError = errors.New("Provided token is not valid yet")
var errTokenExpiredError = errors.New("Token provided in the request has expired")
var errTokenSigInvalidError = errors.New("Token provided in the request has an invalid signature")

// DecodeAndVerifyJWT will decode the provided JWT and verify the validity of its claims.
// If the claims are valid, they are returned, otherwise an error with the reason why
// they are invalid.
func DecodeAndVerifyJWT(secret []byte, authToken string) (*v1.JWTClaims, error) {
	// parse the token
	parser := &jwt.Parser{UseJSONNumber: true}
	token, err := parser.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Incorrect signing algorithm on token")
		}
		// use cache for the JWT secret, since we use it for every request
		return secret, nil
	})
	// Check if token is nil and return error. The error will also be populated
	// if the token was parsed successfully but is invalid.
	if token == nil {
		return nil, fmt.Errorf("Could not parse provided token: %s", err.Error())
	}

	// check token validity
	if !token.Valid {

		// Just the error conditions we have specific messages for
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errTokenMalformedError
			} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				return nil, errTokenExpiredError
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				// would need to stub out time in generation to test this
				return nil, errTokenNotValidYetError
			} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return nil, errTokenSigInvalidError
			}
		}

		// Unhandled token error - generic message
		return nil, fmt.Errorf("Token is invalid: %s", err.Error())
	}

	// Retrieve the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// The claims in the token weren't as expected
		return nil, errors.New("Could not coerce token claims to MapClaims")
	}

	// decode the claims into a session object
	session := &v1.JWTClaims{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  session,
	})
	if err != nil {
		return nil, err
	}
	return session, decoder.Decode(claims)
}

package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// Authenticate is called for API authentication requests. It should generate
// a new JWTClaims object and serve an AuthResult back to the API.
func (a *AuthProvider) Authenticate(req *v1alpha1.LoginRequest) (*v1alpha1.AuthResult, error) {
	r := req.GetRequest()

	// POST methods are the start and end of an oidc flow. If we recorded claims
	// for the provided state we return them back to the API. Otherwise, we start a new flow
	// with the provided state.
	if r.Method == http.MethodPost {
		if req.State == "" {
			return nil, errors.New("No 'state' provided in the request")
		}
		// get the key where we would have stored already authorized claims.
		// This flow should be thought through more. On one hand we are providing an
		// extra verification of the state for the client. On the other hand, if an
		// attacker gets the user's state token mid-flow, they could impersonate the
		// user and steal their token.
		// The client should be generating new state tokens each time, and as long
		// as the full auth flow is encrypted I _think_ the risk is pretty low.
		stateKey := getStateSecretKey(req.State)
		existingClaim, err := a.secrets.ReadSecret(stateKey, true)
		if err != nil {
			if errors.IsSecretNotFoundError(err) {
				return &v1alpha1.AuthResult{
					RedirectURL: a.oauthCfg.AuthCodeURL(req.State),
				}, nil
			}
			return nil, err
		}
		// clear the state secret for this auth session
		if err := a.secrets.Lock(); err != nil {
			return nil, err
		}
		defer a.secrets.Release()
		if err := a.secrets.WriteSecret(stateKey, nil); err != nil {
			return nil, err
		}
		authResult := &v1alpha1.AuthResult{}
		return authResult, json.Unmarshal(existingClaim, authResult)
	}

	// GET is the middle part of the oauth flow. This is to trick the client into
	// sending another post to retrieve its token.

	// fetch the state key from the request
	stateKey := getStateSecretKey(r.URL.Query().Get("state"))
	// get the oauth token from the provider
	oauth2Token, err := a.oauthCfg.Exchange(a.ctx, r.URL.Query().Get("code"))
	if err != nil {
		return nil, err
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, err
	}

	// Parse and verify ID Token payload.
	idToken, err := a.verifier.Verify(a.ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	// parse the claims from the token
	claims := make(map[string]interface{})
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	// start building a user from the claims object
	username, err := getUsernameFromClaims(claims)
	if err != nil {
		return nil, err
	}
	user := &v1alpha1.VDIUser{
		Name:  username,
		Roles: make([]*v1alpha1.VDIUserRole, 0),
	}

	// check if we can handle group membership
	groups, ok := claims[a.cluster.GetOIDCGroupScope()]
	if !ok {
		if a.cluster.AllowNonGroupedReadOnly() {
			user.Roles = []*v1alpha1.VDIUserRole{a.cluster.GetLaunchTemplatesRole().ToUserRole()}
			return nil, a.marshalClaimsToSecret(stateKey, &v1alpha1.AuthResult{User: user})
		}
		return nil, errors.New("No groups provided in claims and allow non-grouped users is set to false")
	}

	userGroupSlc, err := groupClaimToStringSlice(groups)
	if err != nil {
		return nil, err
	}

	// At this point we are ready to authorize the user
	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}

	boundRoles := make([]string, 0)
RoleLoop:
	for _, role := range roles {
		if annotations := role.GetAnnotations(); annotations != nil {
			if oidcGroupStr, ok := annotations[v1alpha1.OIDCGroupRoleAnnotation]; ok {
				oidcGroups := strings.Split(oidcGroupStr, v1alpha1.AuthGroupSeparator)
				for _, group := range oidcGroups {
					if common.StringSliceContains(userGroupSlc, group) {
						boundRoles = common.AppendStringIfMissing(boundRoles, role.GetName())
						continue RoleLoop
					}
				}
			}
		}
	}

	user.Roles = apiutil.FilterUserRolesByNames(roles, boundRoles)
	fmt.Println("Saving claims to state key", stateKey)
	return nil, a.marshalClaimsToSecret(stateKey, &v1alpha1.AuthResult{User: user})
}

func (a *AuthProvider) marshalClaimsToSecret(stateKey string, result *v1alpha1.AuthResult) error {
	if err := a.secrets.Lock(); err != nil {
		return err
	}
	defer a.secrets.Release()
	out, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return a.secrets.WriteSecret(stateKey, out)
}

func getStateSecretKey(state string) string {
	return fmt.Sprintf("oidc_%s", state)
}

func groupClaimToStringSlice(ifc interface{}) ([]string, error) {
	userGroupSlc, ok := ifc.([]interface{})
	if !ok {
		return nil, errors.New("Could not coerce groups claims to string slice")
	}
	out := make([]string, 0)
	for _, item := range userGroupSlc {
		i, ok := item.(string)
		if !ok {
			return nil, errors.New("Could not coerce slice item to string")
		}
		out = append(out, i)
	}
	return out, nil
}

func getUsernameFromClaims(claims map[string]interface{}) (string, error) {
	if preferred, ok := claims["preferred_username"]; ok {
		if prfStr, ok := preferred.(string); ok {
			return prfStr, nil
		}
	}
	if email, ok := claims["email"]; ok {
		if emailStr, ok := email.(string); ok {
			return strings.Split(emailStr, "@")[0], nil
		}
	}
	return "", fmt.Errorf("Could not parse username from claims: %+v", claims)
}

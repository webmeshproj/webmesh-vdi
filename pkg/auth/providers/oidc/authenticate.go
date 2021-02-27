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

package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	rbacv1 "github.com/tinyzimmer/kvdi/apis/rbac/v1"

	"github.com/tinyzimmer/kvdi/pkg/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/rbac"

	"golang.org/x/oauth2"
)

// Authenticate is called for API authentication requests. It should generate
// a new JWTClaims object and serve an AuthResult back to the API.
func (a *AuthProvider) Authenticate(req *types.LoginRequest) (*types.AuthResult, error) {
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
		stateKey := getStateSecretKey(req.GetState())
		existingClaim, err := a.secrets.ReadSecret(stateKey, true)
		if err != nil {
			// If the secret is not found it means we have not generated claims yet
			// for this user. Return the oauth redirect.
			if errors.IsSecretNotFoundError(err) {
				return &types.AuthResult{
					// Use offline access to get a refresh token that we can use to generate new
					// internal access tokens for the user.
					RedirectURL: a.oauthCfg.AuthCodeURL(req.GetState(), oauth2.AccessTypeOffline),
				}, nil
			}
			return nil, err
		}
		// clear the state secret for this auth session
		if err := a.secrets.Lock(15); err != nil {
			return nil, err
		}
		defer a.secrets.Release()
		if err := a.secrets.WriteSecret(stateKey, nil); err != nil {
			return nil, err
		}
		authResult := &types.AuthResult{}
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

	result := &types.AuthResult{
		User: &types.VDIUser{
			Name:  username,
			Roles: make([]*types.VDIUserRole, 0),
		},
		RefreshNotSupported: true,
	}

	// BADDDDD
	if a.cluster.PreserveOIDCTokens() {
		result.Data = map[string]string{
			"access_token":  oauth2Token.AccessToken,
			"token_type":    oauth2Token.TokenType,
			"refresh_token": oauth2Token.RefreshToken,
			"expiry":        oauth2Token.Expiry.Format(time.RFC3339),
		}
	}

	// check if we can handle group membership
	groups, ok := claims[a.cluster.GetOIDCGroupScope()]
	if !ok {
		// if we can't determine group membership, check if cluster configuration
		// allows the user in anyway.
		if a.cluster.AllowNonGroupedReadOnly() {
			result.User.Roles = []*types.VDIUserRole{rbac.VDIRoleToUserRole(a.cluster.GetLaunchTemplatesRole())}
			return nil, a.marshalClaimsToSecret(stateKey, result)
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
	for _, role := range roles {
		boundRoles = appendRoleIfBound(boundRoles, userGroupSlc, role)
	}

	result.User.Roles = apiutil.FilterUserRolesByNames(roles, boundRoles)
	fmt.Println("Saving claims to state key", stateKey)

	// save the claims to the secret backend, they will be retrieved on the next POST
	// for this state.
	return nil, a.marshalClaimsToSecret(stateKey, result)
}

func (a *AuthProvider) marshalClaimsToSecret(stateKey string, result *types.AuthResult) error {
	if err := a.secrets.Lock(15); err != nil {
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

func appendRoleIfBound(boundRoles, userGroups []string, role *rbacv1.VDIRole) []string {
	if annotations := role.GetAnnotations(); annotations != nil {
		if oidcGroupStr, ok := annotations[v1.OIDCGroupRoleAnnotation]; ok {
			oidcGroups := strings.Split(oidcGroupStr, v1.AuthGroupSeparator)
			for _, group := range oidcGroups {
				if group == "" {
					continue
				}
				if common.StringSliceContains(userGroups, group) {
					boundRoles = common.AppendStringIfMissing(boundRoles, role.GetName())
				}
			}
		}
	}
	return boundRoles
}

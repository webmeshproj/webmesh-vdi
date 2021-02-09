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

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

var elevateDenyReason = "The requested operation grants more privileges than the user has."

func denyUserElevatePerms(d *desktopAPI, reqUser *v1.VDIUser, r *http.Request) (allowed bool, reason string, err error) {

	// This is an ugly hack at the moment. This will be triggered if called from
	// allowSameUser while configuring MFA options. No need to check.
	if apiutil.GetGorillaPath(r) == "/api/users/{user}/mfa" {
		return true, "", nil
	}

	// Check that a POST /users will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.CreateUserRequest); ok {
		vdiRoles, err := d.vdiCluster.GetRoles(d.client)
		if err != nil {
			return false, "", err
		}
		for _, role := range reqObj.Roles {
			roleObj := getRoleByName(vdiRoles, role)
			if roleObj == nil {
				continue
			}
			for _, rule := range roleObj.GetRules() {
				if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
					return false, elevateDenyReason, nil
				}
			}
		}
		return true, "", nil
	}

	// Check that a PUT /users/{user} will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.UpdateUserRequest); ok {
		vdiRoles, err := d.vdiCluster.GetRoles(d.client)
		if err != nil {
			return false, "", err
		}
		for _, role := range reqObj.Roles {
			roleObj := getRoleByName(vdiRoles, role)
			if roleObj == nil {
				continue
			}
			for _, rule := range roleObj.GetRules() {
				if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
					return false, elevateDenyReason, nil
				}
			}
		}
		return true, "", nil
	}

	// Check that a POST /roles will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.CreateRoleRequest); ok {
		for _, rule := range reqObj.GetRules() {
			if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
				return false, elevateDenyReason, nil
			}
		}
		return true, "", nil
	}

	// Check that a PUT /roles/{role} will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.UpdateRoleRequest); ok {
		for _, rule := range reqObj.GetRules() {
			if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
				return false, elevateDenyReason, nil
			}
		}
		return true, "", nil
	}

	apiLogger.Info("Method used privilege validator without adding request logic")
	return false, elevateDenyReason, nil
}

func getRoleByName(roles []v1alpha1.VDIRole, name string) *v1alpha1.VDIRole {
	for _, role := range roles {
		if role.GetName() == name {
			return &role
		}
	}
	return nil
}

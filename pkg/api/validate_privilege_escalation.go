package api

import (
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
)

var elevateDenyReason = "The requested operation grants more privileges than the user has."

func denyUserElevatePerms(d *desktopAPI, reqUser *types.User, r *http.Request) (allowed bool, resource, reason string, err error) {
	db, err := d.getDB()
	if err != nil {
		return false, "Invalid", "Failed to connect to database", err
	}
	defer db.Close()

	// Check that a POST /users will not grant permissions the user does not have.
	if reqObj, ok := GetRequestObject(r).(*PostUserRequest); ok {
		resource := fmt.Sprintf("user/%s", reqObj.Username)
		for _, role := range reqObj.Roles {
			roleObj, err := db.GetRole(role)
			if err != nil {
				return false, resource, elevateDenyReason, err
			}
			if reqUser.ElevatedBy(roleObj) {
				return false, resource, elevateDenyReason, nil
			}
		}
		return true, resource, elevateDenyReason, nil
	}

	// Check that a PUT /users/{user} will not grant permissions the user does not have.
	if reqObj, ok := GetRequestObject(r).(*PutUserRequest); ok {
		resource := fmt.Sprintf("user/%s", getUserFromRequest(r))
		for _, role := range reqObj.Roles {
			roleObj, err := db.GetRole(role)
			if err != nil {
				return false, resource, "Error fetching role resource", err
			}
			if reqUser.ElevatedBy(roleObj) {
				return false, resource, elevateDenyReason, nil
			}
		}
		return true, resource, "", nil
	}

	// Check that a POST /roles will not grant permissions the user does not have.
	if reqObj, ok := GetRequestObject(r).(*PostRoleRequest); ok {
		resource := fmt.Sprintf("role/%s", reqObj.Name)
		role := newRoleFromRequest(reqObj)
		if reqUser.ElevatedBy(role) {
			return false, resource, elevateDenyReason, nil
		}
		return true, resource, "", nil
	}

	// Check that a PUT /roles/{role} will not grant permissions the user does not have.
	if reqObj, ok := GetRequestObject(r).(*PutRoleRequest); ok {
		role := newRoleFromPutRequest(getRoleFromRequest(r), reqObj)
		resource := fmt.Sprintf("role/%s", role.Name)
		if reqUser.ElevatedBy(role) {
			return false, resource, elevateDenyReason, nil
		}
		return true, resource, "", nil
	}

	apiLogger.Info("Method used privilege validator without adding request logic")
	return false, "", elevateDenyReason, nil
}

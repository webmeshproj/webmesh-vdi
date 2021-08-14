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

package rbac

import (
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
	"github.com/kvdi/kvdi/pkg/types"
)

// VDIRoleToUserRole converts the given VDIRole to the VDIUserRole format. The VDIUserRole is
// a condensed representation meant to be stored in JWTs.
func VDIRoleToUserRole(v *rbacv1.VDIRole) *types.VDIUserRole {
	return &types.VDIUserRole{
		Name:  v.GetName(),
		Rules: v.GetRules(),
	}
}

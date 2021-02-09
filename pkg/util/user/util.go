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

// Package user is a new packager where user related utilities will be migrated to.
package user

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
)

// FilterTemplates will take a list of DesktopTemplates and filter them based
// off which ones the user is allowed to use.
func FilterTemplates(u *v1.VDIUser, tmpls []v1alpha1.DesktopTemplate) []v1alpha1.DesktopTemplate {
	filtered := make([]v1alpha1.DesktopTemplate, 0)
	for _, tmpl := range tmpls {
		action := &v1.APIAction{
			Verb:         v1.VerbLaunch,
			ResourceType: v1.ResourceTemplates,
			ResourceName: tmpl.GetName(),
		}
		if u.Evaluate(action) {
			filtered = append(filtered, tmpl)
		}
	}
	return filtered
}

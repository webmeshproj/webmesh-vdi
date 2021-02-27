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

package v1

// NamespaceAll represents all namespaces
const NamespaceAll = "*"

// Resource represents the target of an API action
// +kubebuilder:validation:Enum=users;roles;templates;serviceaccounts;*
type Resource string

// Resource options
const (
	// ResourceUsers represents users of kVDI. This action would only apply
	// when using local auth.
	ResourceUsers Resource = "users"
	// ResourceRoles represents the auth roles in kVDI. This would allow a user
	// to manipulate policies via the app API.
	ResourceRoles Resource = "roles"
	// ResourceTeemplates represents desktop templates in kVDI. Mainly the ability
	// to launch seessions from them and connect to them. The "launch" verb can used
	// in this context when referring to launching templates, and the "use" verb for
	// connecting to them via the UI.
	ResourceTemplates Resource = "templates"
	// ResourceServiceAccounts represents kubernetes service accounts. Specifically,
	// the ability to launch desktops that assume them. The API does not expose any
	// CRUD operations on these, but the "use" verb can be used to signal that a user
	// is allowed to assume the given service accounts.
	ResourceServiceAccounts Resource = "serviceaccounts"
	// ResourceAll matches all resources
	ResourceAll Resource = "*"
)

func resourcesToStrings(r []Resource) []string {
	out := make([]string, len(r))
	for x, y := range r {
		out[x] = string(y)
	}
	return out
}

// Verb represents an API action
// +kubebuilder:validation:Enum=create;read;update;delete;use;launch;*
type Verb string

// Verb options
const (
	// Create operations
	VerbCreate Verb = "create"
	// Read operations
	VerbRead Verb = "read"
	// Update operations
	VerbUpdate Verb = "update"
	// Delete operations
	VerbDelete Verb = "delete"
	// Use operations
	VerbUse Verb = "use"
	// Launch operations
	VerbLaunch Verb = "launch"
	// VerbAll matches all actions
	VerbAll Verb = "*"
)

func verbsToStrings(r []Verb) []string {
	out := make([]string, len(r))
	for x, y := range r {
		out[x] = string(y)
	}
	return out
}

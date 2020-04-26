package grants

// RoleGrant resembles a user permission against a requested resource
type RoleGrant int

// Grants that can be assigned to user roles
const (
	// ReadUsers allows a user to retrieve details about all the users
	// in kVDI.
	ReadUsers RoleGrant = 1 << iota
	// WriteUsers allows a user to create/update users in kVDI. Without this
	// permission, a user can still update to their own password.
	WriteUsers
	// ReadGroups: WIP - Not used yet, but will be for grouping users.
	ReadGroups
	// WriteGroups: WIP - Not used yet, but will be for grouping users.
	WriteGroups
	// ReadRoles allow a user to retrieve details about all the roles
	// in kVDI.
	ReadRoles
	// WriteRoles allows a user to create/update roles in kVDI
	WriteRoles
	// ReadTemplates allows a user to list templates available to boot.
	ReadTemplates
	// WriteTemplates: WIP - Will allow a user to create templates from the UI
	// (or API).
	WriteTemplates
	// LaunchTemplates allows a user to launch desktops.
	LaunchTemplates
	// ReadDesktopSessions allows a user to query the status of desktop sessions.
	// Note that without this permission, the user can still query the status of
	// sessions belonging to them.
	ReadDesktopSessions
	// WriteDesktopSessions is intended for making modifications to running desktop
	// sessions. Currently this just provides th ability to `DELETE` a desktop
	// session. Note that without this permission, a user can still end their own
	// desktop sessions.
	WriteDesktopSessions
	// UseDesktopSessions allows a user to connect to the display of desktop sessions.
	// Note that without this permission, a user can still connect to the display
	// of sessions they create.
	UseDesktopSessions
)

const (
	// All contains all the possible grants. This value is intended for assigning
	// to admin roles.
	All RoleGrant = ReadUsers | WriteUsers |
		ReadGroups | WriteGroups |
		ReadRoles | WriteRoles |
		ReadTemplates | WriteTemplates |
		ReadDesktopSessions | WriteDesktopSessions |
		LaunchTemplates | UseDesktopSessions

	// LaunchTemplatesGrant is the permissions applied to the default launch-templates
	// role.
	LaunchTemplatesGrant RoleGrant = ReadTemplates | LaunchTemplates
)

// Grants is an ordered list of all the available grants.
var Grants = []RoleGrant{
	ReadUsers,
	WriteUsers,
	ReadGroups,
	WriteGroups,
	ReadRoles,
	WriteRoles,
	ReadTemplates,
	WriteTemplates,
	LaunchTemplates,
	ReadDesktopSessions,
	WriteDesktopSessions,
	UseDesktopSessions,
}

// GrantNames is an ordered list of the string representation of all the available
// grants.
var GrantNames = []string{
	"ReadUsers",
	"WriteUsers",
	"ReadGroups",
	"WriteGroups",
	"ReadRoles",
	"WriteRoles",
	"ReadTemplates",
	"WriteTemplates",
	"LaunchTemplates",
	"ReadDesktopSessions",
	"WriteDesktopSessions",
	"UseDesktopSessions",
}

// Has returns true if this set of grants includes the one provided.
func (r RoleGrant) Has(grant RoleGrant) bool { return r&grant != 0 }

// Grants returns a slice of the numeical representations for each grant in this
// set.
func (r RoleGrant) Grants() []RoleGrant {
	var result []RoleGrant
	for i := 0; i < len(Grants); i++ {
		if r&(1<<uint(i)) != 0 {
			result = append(result, Grants[i])
		}
	}
	return result
}

// Names returns a list of the string representations for this grant. If
// the grant is the All grant, a slice is returned with just "ADMIN".
func (r RoleGrant) Names() []string {
	if r == All {
		return []string{"ADMIN"}
	}
	var result []string
	for i := 0; i < len(GrantNames); i++ {
		if r&(1<<uint(i)) != 0 {
			result = append(result, GrantNames[i])
		}
	}
	return result
}

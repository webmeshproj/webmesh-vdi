package types

import (
	"reflect"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
)

func newUserWithGrant(grant grants.RoleGrant) *User {
	return &User{
		Name: "test-user",
		Roles: []*Role{
			{
				Grants: grant,
			},
		},
	}
}

func TestUserHasGrant(t *testing.T) {
	allUser := newUserWithGrant(grants.All)
	for _, grant := range grants.Grants {
		if !allUser.HasGrant(grant) {
			t.Error("Expected user with all grant to have grant", grant.Names())
		}
	}

	for idx, grant := range grants.Grants {
		user := newUserWithGrant(grant)
		if !user.HasGrant(grants.Grants[idx]) {
			t.Error("Expected user to have grant", grants.Grants[idx].Names())
		}
		for chkIdx, chkGrant := range grants.Grants {
			if chkIdx != idx {
				if user.HasGrant(chkGrant) {
					t.Error("Expected user not to have grant", chkGrant.Names())
				}
			}
		}
	}
}

func TestUserNamespaces(t *testing.T) {
	nss := []string{"default", "kube-public"}
	user := &User{
		Name: "test-user",
		Roles: []*Role{
			{
				Namespaces: nss,
			},
		},
	}
	if !reflect.DeepEqual(user.Namespaces(), nss) {
		t.Error("Namespace list not as expected, got:", user.Namespaces())
	}

	user.Roles = append(user.Roles, &Role{Namespaces: []string{"default", "kube-system"}})
	if !reflect.DeepEqual(user.Namespaces(), append(nss, "kube-system")) {
		t.Error("Namespace list not as expected, got:", user.Namespaces())
	}
}

func TestRoleNames(t *testing.T) {
	user := &User{
		Name: "test-user",
		Roles: []*Role{
			{
				Name: "test-role",
			},
		},
	}
	if !reflect.DeepEqual(user.RoleNames(), []string{"test-role"}) {
		t.Error("Got unexpected role names:", user.RoleNames())
	}
}

func TestCanLaunch(t *testing.T) {
	testCases := []struct {
		Roles     []*Role
		Template  string
		Namespace string
		Result    bool
	}{
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "lxde-minimal",
			Namespace: "default",
			Result:    true,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.ReadUsers,
				},
			},
			Template:  "lxde-minimal",
			Namespace: "default",
			Result:    false,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "kde-full",
			Namespace: "default",
			Result:    false,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "lxde-minimal",
			Namespace: "kube-system",
			Result:    false,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
				{
					Namespaces:       []string{"kube-system"},
					TemplatePatterns: []string{"kde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "lxde-minimal",
			Namespace: "kube-system",
			Result:    false,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
				{
					Namespaces:       []string{"kube-system"},
					TemplatePatterns: []string{"kde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "kde-full",
			Namespace: "default",
			Result:    false,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
				{
					Namespaces:       []string{"kube-system"},
					TemplatePatterns: []string{"kde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "kde-full",
			Namespace: "kube-system",
			Result:    true,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{},
					Grants:           grants.LaunchTemplates,
				},
				{
					Namespaces:       []string{},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "lxde-minimal",
			Namespace: "kube-system",
			Result:    true,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{},
					Grants:           grants.LaunchTemplates,
				},
				{
					Namespaces:       []string{},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Template:  "kde-full",
			Namespace: "default",
			Result:    true,
		},
	}

	for _, testCase := range testCases {
		user := &User{Roles: testCase.Roles}
		if user.CanLaunch(testCase.Namespace, testCase.Template) != testCase.Result {
			t.Error(testCase)
		}
	}
}

func TestElevatedBy(t *testing.T) {
	testCases := []struct {
		Roles  []*Role
		Input  *Role
		Result bool
	}{
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Input:  &Role{Grants: grants.All},
			Result: true,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Input: &Role{
				Grants:     grants.LaunchTemplates,
				Namespaces: []string{"kube-system"},
			},
			Result: true,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates | grants.ReadUsers,
				},
			},
			Input: &Role{
				Grants: grants.ReadUsers,
			},
			Result: false,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Input: &Role{
				Grants:           grants.LaunchTemplates,
				Namespaces:       []string{"default"},
				TemplatePatterns: []string{"kde-.*"},
			},
			Result: true,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.LaunchTemplates,
				},
			},
			Input: &Role{
				Grants:           grants.LaunchTemplates,
				Namespaces:       []string{"default"},
				TemplatePatterns: []string{"lxde-.*"},
			},
			Result: false,
		},
	}

	for _, testCase := range testCases {
		user := &User{Roles: testCase.Roles}
		if user.ElevatedBy(testCase.Input) != testCase.Result {
			t.Error(testCase)
		}
	}
}

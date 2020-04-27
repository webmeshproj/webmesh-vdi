package types

import (
	"reflect"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/grants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestUserFilterNamespaces(t *testing.T) {
	allNss := []string{"default", "kube-public", "kube-system", "kube-node-lease"}
	nss := []string{"default", "kube-public"}
	user := &User{
		Name: "test-user",
		Roles: []*Role{
			{
				Namespaces: nss,
			},
		},
	}
	if !reflect.DeepEqual(user.FilterNamespaces(allNss), nss) {
		t.Error("Namespace list not as expected, got:", user.FilterNamespaces(allNss))
	}
	user.Roles[0].Namespaces = []string{}
	if !reflect.DeepEqual(user.FilterNamespaces(allNss), allNss) {
		t.Error("Namespace list not as expected, got:", user.FilterNamespaces(allNss))
	}
}

func TestFilterTemplates(t *testing.T) {
	tmpls := []v1alpha1.DesktopTemplate{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lxde-minimal",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kde-full",
			},
		},
	}
	user := &User{
		Name: "test-user",
		Roles: []*Role{
			{
				TemplatePatterns: []string{"lxde-.*"},
			},
		},
	}

	filtered := user.FilterTemplates(tmpls)
	if len(filtered) != 1 {
		t.Error("Expected only one template after filtering, got", len(filtered))
	} else if filtered[0].Name != "lxde-minimal" {
		t.Error("Got wrong template back after filtering, got:", filtered[0].Name)
	}

	user.Roles[0].TemplatePatterns = []string{}
	filtered = user.FilterTemplates(tmpls)
	if len(filtered) != 2 {
		t.Error("Expected both templates after filtering, got", len(filtered))
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
					Grants:           grants.LaunchTemplates | grants.ReadUsers,
				},
			},
			Input:  &Role{Grants: grants.WriteUsers},
			Result: true,
		},
		{
			Roles: []*Role{
				{
					Namespaces:       []string{"default"},
					TemplatePatterns: []string{"lxde-.*"},
					Grants:           grants.WriteUsers,
				},
			},
			Input:  &Role{Grants: grants.WriteRoles},
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

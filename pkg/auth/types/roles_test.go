package types

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
)

func TestRoleGrants(t *testing.T) {
	role := &Role{Grants: grants.All}
	if role.RoleGrants() != grants.All {
		t.Error("Role grants are malformed")
	}
}

func TestRoleMatchesNamespaces(t *testing.T) {
	testCases := []struct {
		RoleNamespaces  []string
		InputNamespaces []string
		ExpectedResult  bool
	}{
		{
			RoleNamespaces:  []string{"default"},
			InputNamespaces: []string{"kube-system"},
			ExpectedResult:  false,
		},
		{
			RoleNamespaces:  []string{},
			InputNamespaces: []string{"kube-system"},
			ExpectedResult:  true,
		},
		{
			RoleNamespaces:  []string{"default"},
			InputNamespaces: []string{},
			ExpectedResult:  false,
		},
		{
			RoleNamespaces:  []string{},
			InputNamespaces: []string{},
			ExpectedResult:  true,
		},
	}
	for _, test := range testCases {
		role := &Role{Namespaces: test.RoleNamespaces}
		if role.MatchesNamespaces(test.InputNamespaces) != test.ExpectedResult {
			t.Error(test)
		}
	}
}

func TestRoleHasNamespace(t *testing.T) {
	testCases := []struct {
		RoleNamespaces []string
		InputNamespace string
		ExpectedResult bool
	}{
		{
			RoleNamespaces: []string{"default"},
			InputNamespace: "kube-system",
			ExpectedResult: false,
		},
		{
			RoleNamespaces: []string{},
			InputNamespace: "kube-system",
			ExpectedResult: true,
		},
		{
			RoleNamespaces: []string{"default", "kube-public"},
			InputNamespace: "app-namespace",
			ExpectedResult: false,
		},
	}
	for _, test := range testCases {
		role := &Role{Namespaces: test.RoleNamespaces}
		if role.HasNamespace(test.InputNamespace) != test.ExpectedResult {
			t.Error(test)
		}
	}
}

func TestRoleHasTemplatePattern(t *testing.T) {
	testCases := []struct {
		RolePatterns   []string
		InputPattern   string
		ExpectedResult bool
	}{
		{
			RolePatterns:   []string{"lxde-.*"},
			InputPattern:   "root-template",
			ExpectedResult: false,
		},
		{
			RolePatterns:   []string{"lxde-.*"},
			InputPattern:   "lxde-.*",
			ExpectedResult: true,
		},
		{
			RolePatterns:   []string{},
			InputPattern:   "root-template",
			ExpectedResult: true,
		},
	}
	for _, test := range testCases {
		role := &Role{TemplatePatterns: test.RolePatterns}
		if role.HasTemplatePattern(test.InputPattern) != test.ExpectedResult {
			t.Error(test)
		}
	}
}

func TestRoleMatchesTemplatePattern(t *testing.T) {
	testCases := []struct {
		RolePatterns   []string
		InputTemplate  string
		ExpectedResult bool
	}{
		{
			RolePatterns:   []string{"lxde-.*"},
			InputTemplate:  "root-template",
			ExpectedResult: false,
		},
		{
			RolePatterns:   []string{"lxde-.*"},
			InputTemplate:  "lxde-minimal",
			ExpectedResult: true,
		},
		{
			RolePatterns:   []string{},
			InputTemplate:  "root-template",
			ExpectedResult: true,
		},
		{
			RolePatterns:   []string{"lxde-.*"},
			InputTemplate:  "kde-full",
			ExpectedResult: false,
		},
	}
	for _, test := range testCases {
		role := &Role{TemplatePatterns: test.RolePatterns}
		if role.MatchesTemplatePattern(test.InputTemplate) != test.ExpectedResult {
			t.Error(test)
		}
	}
}

func TestRoleCanLaunch(t *testing.T) {
	testCases := []struct {
		RoleNamespaces []string
		RolePatterns   []string
		InputTemplate  string
		InputNamespace string
		ExpectedResult bool
	}{
		{
			RoleNamespaces: []string{"default"},
			RolePatterns:   []string{"lxde-.*"},
			InputTemplate:  "lxde-minimal",
			InputNamespace: "default",
			ExpectedResult: true,
		},
		{
			RoleNamespaces: []string{"default"},
			RolePatterns:   []string{"lxde-.*"},
			InputTemplate:  "lxde-minimal",
			InputNamespace: "kube-system",
			ExpectedResult: false,
		},
		{
			RoleNamespaces: []string{"default"},
			RolePatterns:   []string{},
			InputTemplate:  "root-template",
			InputNamespace: "kube-system",
			ExpectedResult: false,
		},
		{
			RoleNamespaces: []string{"default"},
			RolePatterns:   []string{},
			InputTemplate:  "root-template",
			InputNamespace: "default",
			ExpectedResult: true,
		},
		{
			RoleNamespaces: []string{"kube-system"},
			RolePatterns:   []string{},
			InputTemplate:  "root-template",
			InputNamespace: "kube-system",
			ExpectedResult: true,
		},
		{
			RoleNamespaces: []string{"default"},
			RolePatterns:   []string{"lxde-.*"},
			InputTemplate:  "kde-full",
			InputNamespace: "default",
			ExpectedResult: false,
		},
	}
	for _, test := range testCases {
		role := &Role{TemplatePatterns: test.RolePatterns, Namespaces: test.RoleNamespaces}
		if role.CanLaunch(test.InputNamespace, test.InputTemplate) {
			t.Error("Role should require LaunchTemplates grant first")
		}
		role.Grants = grants.LaunchTemplates
		if role.CanLaunch(test.InputNamespace, test.InputTemplate) != test.ExpectedResult {
			t.Error(test)
		}
	}
}

package user

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var testTemplates = []v1alpha1.DesktopTemplate{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-template",
		},
		Spec: v1alpha1.DesktopTemplateSpec{},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "restricted-template",
		},
		Spec: v1alpha1.DesktopTemplateSpec{},
	},
}

var testUser = &v1.VDIUser{
	Roles: []*v1.VDIUserRole{
		{
			Rules: []v1.Rule{
				{
					Verbs:            []v1.Verb{v1.VerbLaunch},
					Resources:        []v1.Resource{v1.ResourceTemplates},
					ResourcePatterns: []string{"test-.*"},
				},
			},
		},
	},
}

func TestFilterTemplates(t *testing.T) {
	allowedTemplates := FilterTemplates(testUser, testTemplates)
	if len(allowedTemplates) != 1 {
		t.Fatalf("Expected only one allowed template, got: %d", len(allowedTemplates))
	}
	if allowedTemplates[0].GetName() != "test-template" {
		t.Errorf("Expected 'test-template' allowed, got: %s", allowedTemplates[0].GetName())
	}
}

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

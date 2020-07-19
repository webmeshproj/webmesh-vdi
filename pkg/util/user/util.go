package user

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
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

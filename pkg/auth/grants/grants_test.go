package grants

import (
	"reflect"
	"testing"
)

func TestGrantHas(t *testing.T) {
	for _, grant := range Grants {
		if !All.Has(grant) {
			t.Error("Expected all grant to have grant", grant.Names())
		}
	}

	for idx, grant := range Grants {
		if !grant.Has(Grants[idx]) {
			t.Error("Expected grant to have itself")
		}
		for chkIdx, chkGrant := range Grants {
			if chkIdx != idx {
				if grant.Has(chkGrant) {
					t.Error("Expected grant not to have others")
				}
			}
		}
	}
}

func TestGrantNames(t *testing.T) {
	if !reflect.DeepEqual(All.Names(), []string{"ADMIN"}) {
		t.Error("Expected ADMIN for all grant, got:", All.Names())
	}
	if !reflect.DeepEqual(LaunchTemplatesGrant.Names(), []string{"ReadTemplates", "LaunchTemplates"}) {
		t.Error("Expected ReadTemplates,LaunchTemplates for LaunchTemplatesGrant, got:", LaunchTemplatesGrant.Names())
	}
}

func TestGrants(t *testing.T) {
	if !reflect.DeepEqual(All.Grants(), Grants) {
		t.Error("Expected all grants for all grants, got:", All.Grants())
	}
}

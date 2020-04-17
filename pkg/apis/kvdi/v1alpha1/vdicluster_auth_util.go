package v1alpha1

import "fmt"

func (c *VDICluster) GetAdminSecret() string {
	if c.Spec.Auth != nil && c.Spec.Auth.AdminSecret != "" {
		return c.Spec.Auth.AdminSecret
	}
	return fmt.Sprintf("%s-admin-secret", c.GetName())
}

func (c *VDICluster) AnonymousAllowed() bool {
	if c.Spec.Auth != nil {
		return c.Spec.Auth.AllowAnonymous
	}
	return false
}

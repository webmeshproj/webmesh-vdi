package v1alpha1

import (
	"fmt"
)

// GetCertManagerNamespace returns the namespace where cert-manager is running.
func (c *VDICluster) GetCertManagerNamespace() string {
	if c.Spec.CertManagerNamespace != "" {
		return c.Spec.CertManagerNamespace
	}
	return "cert-manager"
}

// GetSignerName returns the name of the signing certificate for the VDICluster.
func (c *VDICluster) GetSignerName() string {
	return fmt.Sprintf("%s-mtls-signer.%s.svc", c.GetName(), c.GetCoreNamespace())
}

// GetCAName returns the name of the CA certificate for the VDICluster.
func (c *VDICluster) GetCAName() string {
	return fmt.Sprintf("%s-mtls-root-ca.%s.svc", c.GetName(), c.GetCoreNamespace())
}

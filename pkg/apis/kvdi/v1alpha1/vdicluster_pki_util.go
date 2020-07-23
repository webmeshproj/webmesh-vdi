package v1alpha1

import (
	"fmt"
)

// GetSignerName returns the name of the signing certificate for the VDICluster.
func (c *VDICluster) GetSignerName() string {
	return fmt.Sprintf("%s-mtls-signer.%s.svc", c.GetName(), c.GetCoreNamespace())
}

// GetCAName returns the name of the CA certificate for the VDICluster.
func (c *VDICluster) GetCAName() string {
	return fmt.Sprintf("%s-mtls-root-ca.%s.svc", c.GetName(), c.GetCoreNamespace())
}

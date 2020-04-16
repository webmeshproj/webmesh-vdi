package v1alpha1

import (
	"fmt"
)

func (c *VDICluster) GetCertManagerNamespace() string {
	if c.Spec.CertManagerNamespace != "" {
		return c.Spec.CertManagerNamespace
	}
	return "cert-manager"
}

func (c *VDICluster) GetSignerName() string {
	return fmt.Sprintf("%s-mtls-signer.%s.svc", c.GetName(), c.GetCoreNamespace())
}

func (c *VDICluster) GetCAName() string {
	return fmt.Sprintf("%s-mtls-root-ca.%s.svc", c.GetName(), c.GetCoreNamespace())
}

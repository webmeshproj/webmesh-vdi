package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/client"

// SecretsProvider provides an interface for an app instance to get and store
// any secrets it needs. Currenetly there is only a k8s secret provider, but
// this intreface could be implemented for things like vault.
type SecretsProvider interface {
	// Setup is called before the interface is used for any operations
	Setup(client.Client, *VDICluster) error
	// GetName should return a distinct name that can be used for grabbing external
	// locks during write operations.
	GetName() string
	// ReadSecret should return the contents of a secret by name.
	ReadSecret(name string, cache bool) (contents []byte, err error)
	// WriteSecret should store a secret, replacing any existing one with the
	// same name.
	WriteSecret(name string, contents []byte) error
}

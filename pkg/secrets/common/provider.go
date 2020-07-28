// Package common defines the core interface for various secrets backends to implement.
package common

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretsProvider provides an interface for an app instance to get and store
// any secrets it needs. Currenetly there is only a k8s secret provider, but
// this intreface could be implemented for things like vault.
type SecretsProvider interface {
	// Setup is called before the interface is used for any operations
	Setup(client.Client, *v1alpha1.VDICluster) error
	// ReadSecret should return the contents of a secret by name.
	ReadSecret(name string) (contents []byte, err error)
	// ReadSecretMap should return the contents of a secret map by name.
	ReadSecretMap(name string) (contents map[string][]byte, err error)
	// WriteSecret should store a secret, replacing any existing one with the
	// same name. When contents is nil, the intent is that the secret is removed.
	WriteSecret(name string, contents []byte) error
	// WriteSecretMap should write a map to the secret backend. It should be written in
	// a way that it can be retrieved back into a map of the same types.
	WriteSecretMap(name string, contents map[string][]byte) error
	// Close should handle any cleanup logic for the backend. This method is invoked
	// after temporary usages of the secret engine. This shouldn't be destructive,
	// but it should ensure any opened sockets are closed cleanly, spawned
	// goroutines are finished, and no other dangling references left behind.
	Close() error
}

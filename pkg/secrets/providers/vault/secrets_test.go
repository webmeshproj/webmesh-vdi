package vault

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

func TestReadAndWriteSecret(t *testing.T) {
	provider, srvr, core := mustSetupProvider(t)

	// Cleanup everything after the test
	defer srvr.Close()
	defer core.Shutdown()
	defer provider.Close()

	// check for secret that doesn't exist
	if _, err := provider.ReadSecret("test-secret"); err == nil {
		t.Fatal("Expecter error reading non-existing secret, got nil")
	} else if !errors.IsSecretNotFoundError(err) {
		t.Error("Expected secret not found error, got:", err)
	}

	// write the secret
	if err := provider.WriteSecret("test-secret", []byte("test-value")); err != nil {
		t.Fatal(err)
	}

	// read the secret
	if value, err := provider.ReadSecret("test-secret"); err != nil {
		t.Fatal(err)
	} else if string(value) != "test-value" {
		t.Error("Secret value malformed on retrieval, got:", string(value))
	}

	// delete the secret
	if err := provider.WriteSecret("test-secret", nil); err != nil {
		t.Fatal(err)
	}

	// should be not found again
	if _, err := provider.ReadSecret("test-secret"); err == nil {
		t.Fatal("Expecter error reading non-existing secret, got nil")
	} else if !errors.IsSecretNotFoundError(err) {
		t.Error("Expected secret not found error, got:", err)
	}
}

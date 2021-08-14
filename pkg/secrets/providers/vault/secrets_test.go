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

package vault

import (
	"testing"

	"github.com/kvdi/kvdi/pkg/util/errors"
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

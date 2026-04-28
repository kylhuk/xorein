package storage_test

import (
	"crypto/sha256"
	"encoding/base64"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aether/code_aether/pkg/v0_1/spectest"
)

func vectorDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("spectest: cannot determine test file path")
	}
	// pkg/v0_1/spectest/storage/ → repo root is 4 dirs up.
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
	return filepath.Join(root, "docs", "spec", "v0.1", "91-test-vectors")
}

func TestMain(m *testing.M) {
	// pin.sha256 verification is done per-test below; no global TestMain needed.
	m.Run()
}

// TestStorageKDFSha256 verifies SHA-256(salt || secret) derivation per spec 70 §3.
func TestStorageKDFSha256(t *testing.T) {
	vdir := vectorDir(t)
	spectest.VerifyPin(t, vdir)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "storage_kdf_sha256.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			salt := spectest.Hex(v.Inputs["salt_hex"])
			secret := spectest.Hex(v.Inputs["secret_hex"])
			expected := spectest.Hex(v.ExpectedOutput["store_key_hex"])

			combined := append(append([]byte(nil), salt...), secret...)
			got := sha256.Sum256(combined)
			if string(got[:]) != string(expected) {
				t.Fatalf("store_key mismatch:\n  want %x\n  got  %x", expected, got[:])
			}
		})
	}
}

// TestStorageKeyCheck verifies key_check = base64url_no_pad(SHA-256(store_key || suffix)) per spec 70 §3.1.
func TestStorageKeyCheck(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "storage_key_check.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			storeKey := spectest.Hex(v.Inputs["store_key_hex"])
			expected := v.ExpectedOutput["key_check_b64url"]

			suffix := []byte("xorein-state-store-key-check")
			input := append(append([]byte(nil), storeKey...), suffix...)
			raw := sha256.Sum256(input)
			got := base64.RawURLEncoding.EncodeToString(raw[:])
			if got != expected {
				t.Fatalf("key_check mismatch:\n  want %s\n  got  %s", expected, got)
			}
		})
	}
}

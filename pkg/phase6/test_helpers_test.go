package phase6

import "testing"

func mustSignManifest(t *testing.T, m *Manifest, identity string) *Manifest {
	t.Helper()
	if _, err := m.Sign(identity); err != nil {
		t.Fatalf("sign manifest: %v", err)
	}
	return m
}

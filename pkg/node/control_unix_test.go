//go:build !windows

package node

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultControlEndpointFallsBackToTempDirWhenPathTooLong(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), strings.Repeat("a", 120), strings.Repeat("b", 40))
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	defaultPath := filepath.Join(dataDir, defaultControlName)
	endpoint := defaultControlEndpoint(dataDir)
	if endpoint == defaultPath {
		t.Fatalf("expected shortened control endpoint, got %q", endpoint)
	}
	if filepath.Dir(endpoint) != os.TempDir() {
		t.Fatalf("endpoint dir = %q want %q", filepath.Dir(endpoint), os.TempDir())
	}
	if len(endpoint) > defaultUnixSocketLimit {
		t.Fatalf("endpoint length = %d want <= %d", len(endpoint), defaultUnixSocketLimit)
	}
	ln, got, err := createControlListener("", dataDir)
	if err != nil {
		t.Fatalf("createControlListener() error = %v", err)
	}
	defer func() { _ = ln.Close(); _ = os.Remove(got) }()
	if got != endpoint {
		t.Fatalf("got endpoint = %q want %q", got, endpoint)
	}
}

//go:build !windows

package node

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultControlName     = "xorein-control.sock"
	defaultUnixSocketLimit = 100
)

func createControlListener(endpoint, dataDir string) (net.Listener, string, error) {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = defaultControlEndpoint(dataDir)
	}
	_ = os.Remove(endpoint)
	ln, err := net.Listen("unix", endpoint)
	if err != nil {
		return nil, "", fmt.Errorf("listen control socket: %w", err)
	}
	if err := os.Chmod(endpoint, 0o600); err != nil {
		_ = ln.Close()
		return nil, "", fmt.Errorf("chmod control socket: %w", err)
	}
	return ln, endpoint, nil
}

func defaultControlEndpoint(dataDir string) string {
	endpoint := filepath.Join(dataDir, defaultControlName)
	if len(endpoint) <= defaultUnixSocketLimit {
		return endpoint
	}
	sum := sha256.Sum256([]byte(dataDir))
	shortName := "xorein-" + hex.EncodeToString(sum[:6]) + ".sock"
	return filepath.Join(os.TempDir(), shortName)
}

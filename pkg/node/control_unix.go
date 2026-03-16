//go:build !windows

package node

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const defaultControlName = "xorein-control.sock"

func createControlListener(endpoint, dataDir string) (net.Listener, string, error) {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = filepath.Join(dataDir, defaultControlName)
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

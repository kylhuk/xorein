//go:build windows

package node

import (
	"fmt"
	"net"
	"strings"
)

func createControlListener(endpoint, dataDir string) (net.Listener, string, error) {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = "127.0.0.1:0"
	}
	ln, err := net.Listen("tcp", endpoint)
	if err != nil {
		return nil, "", fmt.Errorf("listen control endpoint: %w", err)
	}
	return ln, ln.Addr().String(), nil
}

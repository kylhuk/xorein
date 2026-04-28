package control_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aether/code_aether/pkg/v0_1/control"
)

func newTestServer(t *testing.T) (*control.Server, string) {
	t.Helper()
	dir := t.TempDir()
	srv, err := control.New(control.Config{
		DataDir:  dir,
		Handlers: &control.Handlers{PeerID: "test-peer-id"},
	})
	if err != nil {
		t.Fatalf("control.New: %v", err)
	}
	t.Cleanup(func() { srv.Shutdown(context.Background()) }) //nolint:errcheck
	go srv.Serve() //nolint:errcheck

	// Read token for authenticated requests.
	raw, err := os.ReadFile(filepath.Join(dir, "control.token"))
	if err != nil {
		t.Fatalf("read token: %v", err)
	}
	return srv, strings.TrimSpace(string(raw))
}

func doReq(t *testing.T, srv *control.Server, token, method, path, body string) *http.Response {
	t.Helper()
	addr := srv.Addr()

	var transport http.RoundTripper
	if strings.HasPrefix(addr, "/") || strings.Contains(addr, ".sock") {
		transport = &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.Dial("unix", addr)
			},
		}
	} else {
		transport = http.DefaultTransport
	}
	client := &http.Client{Transport: transport}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, "http://x"+path, bodyReader)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	return resp
}

func TestTokenGeneration(t *testing.T) {
	dir := t.TempDir()
	_, err := control.New(control.Config{DataDir: dir, Handlers: &control.Handlers{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "control.token"))
	if err != nil {
		t.Fatalf("read token: %v", err)
	}
	token := strings.TrimSpace(string(raw))
	if len(token) < 20 {
		t.Fatalf("token too short: %q", token)
	}
	// Verify file permissions (0600).
	info, err := os.Stat(filepath.Join(dir, "control.token"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("token file perm = %04o, want 0600", perm)
	}
}

func TestTokenReuse(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "control.token"), []byte("my-existing-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := control.New(control.Config{DataDir: dir, Handlers: &control.Handlers{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "control.token"))
	if strings.TrimSpace(string(raw)) != "my-existing-token" {
		t.Fatalf("expected token to be reused, got %q", raw)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doReq(t, srv, "" /*no token*/, "GET", "/v1/state", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", resp.StatusCode)
	}
	var e map[string]string
	json.NewDecoder(resp.Body).Decode(&e) //nolint:errcheck
	if e["code"] != control.CodeUnauthorized {
		t.Fatalf("want code=%q, got %q", control.CodeUnauthorized, e["code"])
	}
}

func TestAuthMiddleware_WrongToken(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doReq(t, srv, "wrong-token", "GET", "/v1/state", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", resp.StatusCode)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv, token := newTestServer(t)
	// POST to a GET-only endpoint should return 405.
	resp := doReq(t, srv, token, "POST", "/v1/state", "{}")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", resp.StatusCode)
	}
	var e map[string]string
	json.NewDecoder(resp.Body).Decode(&e) //nolint:errcheck
	if e["code"] != control.CodeMethodNotAllowed {
		t.Fatalf("want code=%q, got %q", control.CodeMethodNotAllowed, e["code"])
	}
}

func TestNotFound(t *testing.T) {
	srv, token := newTestServer(t)
	resp := doReq(t, srv, token, "GET", "/v1/unknown-endpoint", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404, got %d", resp.StatusCode)
	}
	var e map[string]string
	json.NewDecoder(resp.Body).Decode(&e) //nolint:errcheck
	if e["code"] != control.CodeNotFound {
		t.Fatalf("want code=%q, got %q", control.CodeNotFound, e["code"])
	}
}

func TestBodyTooLarge(t *testing.T) {
	srv, token := newTestServer(t)
	// Build a body larger than 1 MiB.
	large := fmt.Sprintf(`{"display_name":"%s"}`, strings.Repeat("x", 1<<20+1))
	resp := doReq(t, srv, token, "POST", "/v1/identities", large)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("want 413, got %d", resp.StatusCode)
	}
}

func TestGetState(t *testing.T) {
	srv, token := newTestServer(t)
	resp := doReq(t, srv, token, "GET", "/v1/state", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	var snap map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if snap["peer_id"] != "test-peer-id" {
		t.Fatalf("unexpected peer_id: %v", snap["peer_id"])
	}
}

func TestSSEReadyEvent(t *testing.T) {
	srv, token := newTestServer(t)
	addr := srv.Addr()

	var transport http.RoundTripper
	if strings.HasPrefix(addr, "/") || strings.Contains(addr, ".sock") {
		transport = &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.Dial("unix", addr)
			},
		}
	} else {
		transport = http.DefaultTransport
	}
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", "http://x/v1/events", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("SSE connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("want Content-Type: text/event-stream, got %q", ct)
	}

	// Read the first line of the SSE stream — should be "event: ready"
	buf := make([]byte, 512)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])
	if !strings.Contains(body, "event: ready") {
		t.Fatalf("expected ready event, got: %q", body)
	}
	if !strings.Contains(body, `"version":"1"`) {
		t.Fatalf("expected version:1 in ready event, got: %q", body)
	}
}

func TestStaleSocketCleanup(t *testing.T) {
	if isWindows() {
		t.Skip("Unix-socket test skipped on Windows")
	}
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "xorein-control.sock")

	// Simulate a stale socket file from an unclean shutdown.
	// net.UnixListener.Close() auto-unlinks, so write a plain file instead.
	if err := os.WriteFile(sockPath, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}

	// New() should remove the stale file and bind successfully.
	srv, err := control.New(control.Config{DataDir: dir, Handlers: &control.Handlers{}})
	if err != nil {
		t.Fatalf("New with stale socket: %v", err)
	}
	srv.Shutdown(context.Background()) //nolint:errcheck
}

func TestUnixSocketBinding(t *testing.T) {
	if isWindows() {
		t.Skip("Unix-socket test skipped on Windows")
	}
	srv, token := newTestServer(t)
	resp := doReq(t, srv, token, "GET", "/v1/state", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 over Unix socket, got %d", resp.StatusCode)
	}
}

func TestAllErrorCodes(t *testing.T) {
	codes := []string{
		control.CodeUnauthorized,
		control.CodeForbidden,
		control.CodeMethodNotAllowed,
		control.CodeInvalidRequest,
		control.CodeNotFound,
		control.CodeInvalidSignature,
		control.CodeExpiredInvite,
		control.CodeJoinFailed,
		control.CodePreviewFailed,
		control.CodeBackupFailed,
		control.CodeUnsupported,
	}
	if len(codes) != 11 {
		t.Fatalf("expected 11 error codes, got %d", len(codes))
	}
	for _, c := range codes {
		if c == "" {
			t.Fatalf("empty error code in list")
		}
	}
}

func isWindows() bool {
	return strings.EqualFold(os.Getenv("GOOS"), "windows")
}

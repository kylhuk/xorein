package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBodyLimit = 1 << 20  // 1 MiB (spec 60 §6)
	voiceFrameLimit  = 8 << 20  // 8 MiB (spec 60 §6)
	sockFileName     = "xorein-control.sock"
	addrFileName     = "control.addr"
)

// Config configures the local control API server.
type Config struct {
	// DataDir is the data directory (for token, socket path, addr file).
	DataDir string
	// Addr overrides the default socket path (Linux/macOS) or TCP addr (Windows).
	Addr     string
	Handlers *Handlers
}

// Server is the local control HTTP server.
type Server struct {
	httpSrv *http.Server
	ln      net.Listener
	hs      *Handlers
	sse     *Mux
	st      *controlState
	dataDir string
}

// New creates a Server: generates/reads the bearer token, binds the listener,
// and sets up the full route table. Call Serve() to start accepting connections.
func New(cfg Config) (*Server, error) {
	token, err := loadOrCreateToken(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	addr := cfg.Addr
	if addr == "" {
		if runtime.GOOS == "windows" {
			addr = "127.0.0.1:0"
		} else {
			addr = filepath.Join(cfg.DataDir, sockFileName)
		}
	}

	ln, err := createListener(addr, cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("control: bind listener: %w", err)
	}

	hs := cfg.Handlers
	if hs == nil {
		hs = &Handlers{}
	}

	srv := &Server{
		hs:      hs,
		sse:     &Mux{},
		st:      newControlState(),
		dataDir: cfg.DataDir,
	}

	routes := srv.buildRoutes()
	authed := newAuthMiddleware(token)(routes)
	guarded := localhostGuard(authed)

	srv.httpSrv = &http.Server{
		Handler:           guarded,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}
	srv.ln = ln
	return srv, nil
}

// Serve starts accepting connections (blocks until listener is closed).
func (s *Server) Serve() error {
	return s.httpSrv.Serve(s.ln)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

// Addr returns the listener address (socket path or TCP addr).
func (s *Server) Addr() string {
	return s.ln.Addr().String()
}

// SSE returns the event multiplexer so callers can publish events.
func (s *Server) SSE() *Mux { return s.sse }

// AddNotification appends a notification to the control state and publishes
// a "notification" SSE event. Called by the runtime when a notify.deliver
// arrives over the P2P stream.
func (s *Server) AddNotification(n *NotificationRecord) {
	s.st.mu.Lock()
	s.st.notifications = append(s.st.notifications, n)
	s.st.mu.Unlock()
	s.sse.Publish(Event{Type: "notification", Data: n})
}

// createListener creates the appropriate listener for the platform.
// On Windows: loopback TCP on an OS-assigned port; writes port to <dataDir>/control.addr.
// On Unix: Unix-domain socket; removes stale socket file, sets 0600.
func createListener(addr, dataDir string) (net.Listener, error) {
	if runtime.GOOS == "windows" {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
		addrPath := filepath.Join(dataDir, addrFileName)
		if err := os.WriteFile(addrPath, []byte("127.0.0.1:"+port+"\n"), 0o600); err != nil {
			ln.Close()
			return nil, fmt.Errorf("control: write control.addr: %w", err)
		}
		return ln, nil
	}
	// Unix socket: clean up stale socket, then bind.
	_ = os.Remove(addr)
	ln, err := net.Listen("unix", addr)
	if err != nil {
		return nil, err
	}
	_ = os.Chmod(addr, 0o600)
	return ln, nil
}

// localhostGuard rejects non-loopback connections on Windows TCP listeners.
func localhostGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if runtime.GOOS == "windows" {
			host := r.RemoteAddr
			if idx := strings.LastIndex(host, ":"); idx >= 0 {
				host = host[:idx]
			}
			host = strings.TrimPrefix(host, "[")
			host = strings.TrimSuffix(host, "]")
			if host != "127.0.0.1" && host != "::1" && host != "localhost" {
				writeError(w, http.StatusForbidden, CodeForbidden, "non-local connection rejected")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// pathRoute maps a URL pattern to the handlers for each HTTP method.
type pathRoute struct {
	pattern  string
	methods  map[string]http.HandlerFunc
	limit    int64 // 0 → defaultBodyLimit
}

// buildRoutes sets up the ServeMux using path-only patterns. Each pattern's
// handler checks the HTTP method and returns 405 for unrecognised methods,
// giving us JSON-formatted 405 bodies without ServeMux pattern conflicts.
func (s *Server) buildRoutes() http.Handler {
	mux := http.NewServeMux()

	routes := []pathRoute{
		// State + Events
		{"/v1/state", map[string]http.HandlerFunc{"GET": s.handleState}, 0},
		{"/v1/events", map[string]http.HandlerFunc{"GET": s.sse.ServeHTTP}, 0},

		// Identities
		{"/v1/identities/backup", map[string]http.HandlerFunc{"GET": s.handleBackupIdentity}, 0},
		{"/v1/identities/restore", map[string]http.HandlerFunc{"POST": s.handleRestoreIdentity}, 0},
		{"/v1/identities", map[string]http.HandlerFunc{
			"GET":  s.handleGetIdentities,
			"POST": s.handleCreateIdentity,
		}, 0},

		// Servers (fixed paths must be listed before wildcard paths)
		{"/v1/servers/join", map[string]http.HandlerFunc{"POST": s.handleJoinServer}, 0},
		{"/v1/servers/preview", map[string]http.HandlerFunc{"POST": s.handlePreviewServer}, 0},
		{"/v1/servers/{serverID}/channels", map[string]http.HandlerFunc{"POST": s.handleCreateChannel}, 0},
		{"/v1/servers", map[string]http.HandlerFunc{
			"GET":  s.handleListServers,
			"POST": s.handleCreateServer,
		}, 0},

		// Messages
		{"/v1/messages/search", map[string]http.HandlerFunc{"POST": s.handleSearchMessages}, 0},
		{"/v1/messages/{messageID}", map[string]http.HandlerFunc{
			"PATCH":  s.handleEditMessage,
			"DELETE": s.handleDeleteMessage,
		}, 0},
		{"/v1/channels/{channelID}/messages", map[string]http.HandlerFunc{"POST": s.handleSendChannelMessage}, 0},

		// DMs
		{"/v1/dms/{dmID}/messages", map[string]http.HandlerFunc{"POST": s.handleSendDMMessage}, 0},
		{"/v1/dms", map[string]http.HandlerFunc{
			"GET":  s.handleListDMs,
			"POST": s.handleCreateDM,
		}, 0},

		// Friends
		{"/v1/friends/requests/{requestID}", map[string]http.HandlerFunc{"PUT": s.handleActOnFriendRequest}, 0},
		{"/v1/friends/requests", map[string]http.HandlerFunc{"POST": s.handleSendFriendRequest}, 0},
		{"/v1/friends/{friendID}", map[string]http.HandlerFunc{"DELETE": s.handleDeleteFriend}, 0},
		{"/v1/friends", map[string]http.HandlerFunc{"GET": s.handleListFriends}, 0},

		// Voice (frames get a larger body limit per spec 60 §6)
		{"/v1/voice/{channelID}/frames", map[string]http.HandlerFunc{"POST": s.handleVoiceFrames}, voiceFrameLimit},
		{"/v1/voice/{channelID}/join", map[string]http.HandlerFunc{"POST": s.handleVoiceJoin}, 0},
		{"/v1/voice/{channelID}/leave", map[string]http.HandlerFunc{"POST": s.handleVoiceLeave}, 0},
		{"/v1/voice/{channelID}/mute", map[string]http.HandlerFunc{"POST": s.handleVoiceMute}, 0},

		// Presence
		{"/v1/presence", map[string]http.HandlerFunc{"GET": s.handleGetPresence}, 0},

		// Notifications
		{"/v1/notifications/summary", map[string]http.HandlerFunc{"GET": s.handleNotificationSummary}, 0},
		{"/v1/notifications/search", map[string]http.HandlerFunc{"POST": s.handleSearchNotifications}, 0},
		{"/v1/notifications/read", map[string]http.HandlerFunc{"POST": s.handleMarkRead}, 0},

		// Mentions
		{"/v1/mentions/search", map[string]http.HandlerFunc{"POST": s.handleSearchMentions}, 0},

		// Peers + Relays
		{"/v1/peers/manual", map[string]http.HandlerFunc{
			"POST":   s.handleAddManualPeer,
			"DELETE": s.handleRemoveManualPeer,
		}, 0},
		{"/v1/relays", map[string]http.HandlerFunc{"POST": s.handleAddRelay}, 0},

		// Per-peer connection type (spec 32 §5)
		{"/v1/peers/{peerID}/connection", map[string]http.HandlerFunc{"GET": s.handleGetPeerConnection}, 0},
	}

	for _, r := range routes {
		limit := r.limit
		if limit == 0 {
			limit = defaultBodyLimit
		}
		mux.HandleFunc(r.pattern, withBodyLimit(limit, methodRouter(r.methods)))
	}

	// 404 catch-all for unrecognised paths.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, CodeNotFound, "endpoint not found")
	})

	return mux
}

// methodRouter dispatches to per-method handlers and returns 405 for unknown methods.
func methodRouter(handlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h, ok := handlers[r.Method]
		if !ok {
			writeError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		h(w, r)
	}
}

// withBodyLimit wraps a handler with an http.MaxBytesReader body limit.
// Requests exceeding the limit get 413 and invalid_request.
func withBodyLimit(limit int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next(w, r)
	}
}

// decodeBody decodes JSON from r.Body into v; writes an error response and
// returns false if decoding fails (including body-too-large).
func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "request body too large") || strings.Contains(msg, "http: request body too large") {
			writeError(w, http.StatusRequestEntityTooLarge, CodeInvalidRequest, "request body too large")
		} else {
			writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body: "+err.Error())
		}
		return false
	}
	return true
}

// newID generates a simple collision-resistant ID using time + a short random suffix.
func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}


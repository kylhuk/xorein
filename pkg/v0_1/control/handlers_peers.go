package control

import (
	"net/http"
)

type manualPeerRequest struct {
	Address string `json:"address"`
}

func (s *Server) handleAddManualPeer(w http.ResponseWriter, r *http.Request) {
	var req manualPeerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Address == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "address is required")
		return
	}
	s.st.mu.Lock()
	for _, p := range s.st.manualPeers {
		if p == req.Address {
			s.st.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	s.st.manualPeers = append(s.st.manualPeers, req.Address)
	s.st.mu.Unlock()
	if s.hs != nil && s.hs.AddManualPeerFn != nil {
		s.hs.AddManualPeerFn(req.Address)
	}
	s.sse.Publish(Event{Type: "peer", Data: map[string]any{"action": "add", "address": req.Address}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRemoveManualPeer(w http.ResponseWriter, r *http.Request) {
	var req manualPeerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Address == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "address is required")
		return
	}
	s.st.mu.Lock()
	remaining := s.st.manualPeers[:0]
	for _, p := range s.st.manualPeers {
		if p != req.Address {
			remaining = append(remaining, p)
		}
	}
	s.st.manualPeers = remaining
	s.st.mu.Unlock()
	if s.hs != nil && s.hs.RemoveManualPeerFn != nil {
		s.hs.RemoveManualPeerFn(req.Address)
	}
	s.sse.Publish(Event{Type: "peer", Data: map[string]any{"action": "remove", "address": req.Address}})
	w.WriteHeader(http.StatusNoContent)
}

// peerConnectionResponse is the JSON shape for GET /v1/peers/{peerID}/connection.
type peerConnectionResponse struct {
	PeerID    string `json:"peer_id"`
	Type      string `json:"type"`
	// LatencyMs is the EWMA round-trip latency to the peer in milliseconds.
	// -1 means the latency is unavailable (peer not tracked or peerstore unsupported).
	// Spec 32 §5.
	LatencyMs int64  `json:"latency_ms"`
}

// handleGetPeerConnection returns the connection type and latency indicator for
// a specific peer (spec 32 §5).
func (s *Server) handleGetPeerConnection(w http.ResponseWriter, r *http.Request) {
	peerID := r.PathValue("peerID")
	if peerID == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "peerID is required")
		return
	}
	connType := "unknown"
	if s.hs != nil && s.hs.ConnectionTypeFn != nil {
		connType = s.hs.ConnectionTypeFn(peerID)
	}
	var latencyMs int64 = -1
	if s.hs != nil && s.hs.LatencyFn != nil {
		latencyMs = s.hs.LatencyFn(peerID)
	}
	writeJSON(w, http.StatusOK, peerConnectionResponse{
		PeerID:    peerID,
		Type:      connType,
		LatencyMs: latencyMs,
	})
}

type addRelayRequest struct {
	Multiaddr string `json:"multiaddr"`
}

func (s *Server) handleAddRelay(w http.ResponseWriter, r *http.Request) {
	var req addRelayRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Multiaddr == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "multiaddr is required")
		return
	}
	s.st.mu.Lock()
	for _, ra := range s.st.relayAddrs {
		if ra == req.Multiaddr {
			s.st.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	s.st.relayAddrs = append(s.st.relayAddrs, req.Multiaddr)
	s.st.mu.Unlock()
	if s.hs != nil && s.hs.AddRelayFn != nil {
		s.hs.AddRelayFn(req.Multiaddr)
	}
	s.sse.Publish(Event{Type: "relay", Data: map[string]any{"action": "add", "multiaddr": req.Multiaddr}})
	w.WriteHeader(http.StatusNoContent)
}

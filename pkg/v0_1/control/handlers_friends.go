package control

import (
	"net/http"
	"time"
)

func (s *Server) handleListFriends(w http.ResponseWriter, r *http.Request) {
	s.st.mu.RLock()
	var out []*FriendRequestRecord
	for _, fr := range s.st.friendReqs {
		if fr.Status == "accepted" {
			out = append(out, fr)
		}
	}
	s.st.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"friends": out})
}

type sendFriendRequestRequest struct {
	PeerAddr string `json:"peer_addr"`
}

func (s *Server) handleSendFriendRequest(w http.ResponseWriter, r *http.Request) {
	var req sendFriendRequestRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.PeerAddr == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "peer_addr is required")
		return
	}
	fr := &FriendRequestRecord{
		ID:         newID("freq"),
		FromPeerID: s.hs.PeerID,
		ToPeerAddr: req.PeerAddr,
		Status:     "pending",
		CreatedAt:  time.Now().UTC(),
	}
	s.st.mu.Lock()
	s.st.friendReqs = append(s.st.friendReqs, fr)
	s.st.mu.Unlock()
	s.sse.Publish(Event{Type: "friend", Data: fr})
	writeJSON(w, http.StatusCreated, fr)
}

type actOnFriendRequestBody struct {
	Action string `json:"action"` // accept | decline | cancel | block
}

func (s *Server) handleActOnFriendRequest(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("requestID")
	var req actOnFriendRequestBody
	if !decodeBody(w, r, &req) {
		return
	}
	validActions := map[string]bool{"accept": true, "decline": true, "cancel": true, "block": true}
	if !validActions[req.Action] {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "action must be one of: accept, decline, cancel, block")
		return
	}

	s.st.mu.Lock()
	var found *FriendRequestRecord
	for _, fr := range s.st.friendReqs {
		if fr.ID == requestID {
			found = fr
			break
		}
	}
	if found != nil {
		found.Status = req.Action + "ed"
		if req.Action == "accept" {
			found.Status = "accepted"
		} else if req.Action == "decline" {
			found.Status = "declined"
		} else if req.Action == "cancel" {
			found.Status = "cancelled"
		} else if req.Action == "block" {
			found.Status = "blocked"
		}
	}
	s.st.mu.Unlock()

	if found == nil {
		writeError(w, http.StatusNotFound, CodeNotFound, "friend request not found")
		return
	}
	s.sse.Publish(Event{Type: "friend", Data: found})
	writeJSON(w, http.StatusOK, found)
}

func (s *Server) handleDeleteFriend(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friendID")

	s.st.mu.Lock()
	var found bool
	remaining := s.st.friendReqs[:0]
	for _, fr := range s.st.friendReqs {
		if fr.ID == friendID && fr.Status == "accepted" {
			found = true
			continue
		}
		remaining = append(remaining, fr)
	}
	if found {
		s.st.friendReqs = remaining
	}
	s.st.mu.Unlock()

	if !found {
		writeError(w, http.StatusNotFound, CodeNotFound, "friend not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

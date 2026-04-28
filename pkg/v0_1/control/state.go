package control

import (
	"sync"
	"time"
)

// ServerRecord is a local server entry.
type ServerRecord struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	OwnerPeerID  string    `json:"owner_peer_id"`
	SecurityMode string    `json:"security_mode"`
	CreatedAt    time.Time `json:"created_at"`
}

// ChannelRecord is a channel inside a server.
type ChannelRecord struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"server_id"`
	Name      string    `json:"name"`
	Voice     bool      `json:"voice"`
	CreatedAt time.Time `json:"created_at"`
}

// DMRecord is a DM conversation record.
type DMRecord struct {
	ID        string    `json:"id"`
	PeerID    string    `json:"peer_id"`
	CreatedAt time.Time `json:"created_at"`
}

// MessageRecord is a message (channel or DM).
type MessageRecord struct {
	ID          string    `json:"id"`
	ScopeID     string    `json:"scope_id"`
	ScopeType   string    `json:"scope_type"` // "channel" or "dm"
	SenderPeerID string   `json:"sender_peer_id"`
	Body        string    `json:"body"`
	EditedAt    *time.Time `json:"edited_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// FriendRequestRecord is a friend request or accepted friendship.
type FriendRequestRecord struct {
	ID          string    `json:"id"`
	FromPeerID  string    `json:"from_peer_id"`
	ToPeerAddr  string    `json:"to_peer_addr,omitempty"`
	Status      string    `json:"status"` // "pending", "accepted", "declined", "cancelled", "blocked"
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationRecord is a notification entry.
type NotificationRecord struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	ScopeID   string    `json:"scope_id,omitempty"`
	ScopeType string    `json:"scope_type,omitempty"`
	ServerID  string    `json:"server_id,omitempty"`
	MessageID string    `json:"message_id,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// ReadThrough records the read-through position for a scope.
type ReadThrough struct {
	ScopeID         string    `json:"scope_id"`
	ScopeType       string    `json:"scope_type"`
	ReadThroughMsgID string   `json:"read_through_message_id"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// controlState is the mutable local control-layer state.
type controlState struct {
	mu            sync.RWMutex
	servers       []*ServerRecord
	channels      []*ChannelRecord
	dms           []*DMRecord
	messages      []*MessageRecord
	friendReqs    []*FriendRequestRecord
	notifications []*NotificationRecord
	readThrough   map[string]*ReadThrough // scopeID → position
	manualPeers   []string
	relayAddrs    []string
}

func newControlState() *controlState {
	return &controlState{
		readThrough: make(map[string]*ReadThrough),
	}
}

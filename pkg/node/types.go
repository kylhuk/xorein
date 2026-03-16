package node

import "time"

type Role string

const (
	RoleClient    Role = "client"
	RoleRelay     Role = "relay"
	RoleBootstrap Role = "bootstrap"
)

func (r Role) Valid() bool {
	switch r {
	case RoleClient, RoleRelay, RoleBootstrap:
		return true
	default:
		return false
	}
}

type Config struct {
	Role              Role
	DataDir           string
	ListenAddr        string
	BootstrapAddrs    []string
	ManualPeers       []string
	RelayAddrs        []string
	ControlEndpoint   string
	DiscoveryInterval time.Duration
	HistoryLimit      int
}

type Profile struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio,omitempty"`
}

type Identity struct {
	ID         string    `json:"id"`
	PeerID     string    `json:"peer_id"`
	PublicKey  string    `json:"public_key"`
	PrivateKey string    `json:"private_key"`
	CreatedAt  time.Time `json:"created_at"`
	Profile    Profile   `json:"profile"`
}

type PeerRecord struct {
	PeerID     string    `json:"peer_id"`
	Role       Role      `json:"role"`
	Addresses  []string  `json:"addresses"`
	PublicKey  string    `json:"public_key,omitempty"`
	Source     string    `json:"source,omitempty"`
	LastSeenAt time.Time `json:"last_seen_at,omitempty"`
}

type ChannelRecord struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"server_id"`
	Name      string    `json:"name"`
	Voice     bool      `json:"voice"`
	CreatedAt time.Time `json:"created_at"`
}

type DMRecord struct {
	ID           string    `json:"id"`
	Participants []string  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
}

type MessageRecord struct {
	ID           string    `json:"id"`
	ScopeType    string    `json:"scope_type"`
	ScopeID      string    `json:"scope_id"`
	ServerID     string    `json:"server_id,omitempty"`
	SenderPeerID string    `json:"sender_peer_id"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	Deleted      bool      `json:"deleted,omitempty"`
}

type VoiceParticipant struct {
	PeerID      string    `json:"peer_id"`
	Muted       bool      `json:"muted"`
	JoinedAt    time.Time `json:"joined_at"`
	LastFrameAt time.Time `json:"last_frame_at,omitempty"`
}

type VoiceSession struct {
	ChannelID    string                         `json:"channel_id"`
	Participants map[string]VoiceParticipant    `json:"participants"`
	LastFrameBy  map[string]time.Time           `json:"last_frame_by,omitempty"`
}

type ServerRecord struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	OwnerPeerID string                   `json:"owner_peer_id"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Members     []string                 `json:"members"`
	Channels    map[string]ChannelRecord `json:"channels"`
	Manifest    Manifest                 `json:"manifest"`
	Invite      string                   `json:"invite,omitempty"`
}

type Event struct {
	Type    string         `json:"type"`
	Time    time.Time      `json:"time"`
	Payload map[string]any `json:"payload,omitempty"`
}

type Snapshot struct {
	Role            Role                    `json:"role"`
	PeerID          string                  `json:"peer_id"`
	ListenAddresses []string                `json:"listen_addresses"`
	ControlEndpoint string                  `json:"control_endpoint"`
	Identity        Identity                `json:"identity"`
	KnownPeers      []PeerRecord            `json:"known_peers"`
	Servers         []ServerRecord          `json:"servers"`
	DMs             []DMRecord              `json:"dms"`
	Messages        []MessageRecord         `json:"messages"`
	VoiceSessions   []VoiceSession          `json:"voice_sessions"`
	Settings        map[string]string       `json:"settings,omitempty"`
	Telemetry       []string                `json:"telemetry,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

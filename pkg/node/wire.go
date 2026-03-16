package node

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	ControlAPIVersion = "v1"
	joinScheme        = "aether"
	joinHost          = "join"
)

type Manifest struct {
	ServerID        string    `json:"server_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	OwnerPeerID     string    `json:"owner_peer_id"`
	OwnerPublicKey  string    `json:"owner_public_key"`
	OwnerAddresses  []string  `json:"owner_addresses"`
	BootstrapAddrs  []string  `json:"bootstrap_addrs,omitempty"`
	RelayAddrs      []string  `json:"relay_addrs,omitempty"`
	Capabilities    []string  `json:"capabilities"`
	IssuedAt        time.Time `json:"issued_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	Signature       string    `json:"signature"`
}

type Invite struct {
	ServerID       string    `json:"server_id"`
	OwnerPeerID    string    `json:"owner_peer_id"`
	OwnerPublicKey string    `json:"owner_public_key"`
	ServerAddrs    []string  `json:"server_addrs"`
	BootstrapAddrs []string  `json:"bootstrap_addrs,omitempty"`
	RelayAddrs     []string  `json:"relay_addrs,omitempty"`
	ManifestHash   string    `json:"manifest_hash"`
	ExpiresAt      time.Time `json:"expires_at"`
	Signature      string    `json:"signature"`
}

type Delivery struct {
	ID               string    `json:"id"`
	Kind             string    `json:"kind"`
	ScopeID          string    `json:"scope_id"`
	ScopeType        string    `json:"scope_type"`
	ServerID         string    `json:"server_id,omitempty"`
	SenderPeerID     string    `json:"sender_peer_id"`
	SenderPublicKey  string    `json:"sender_public_key"`
	RecipientPeerIDs []string  `json:"recipient_peer_ids"`
	Body             string    `json:"body,omitempty"`
	Data             []byte    `json:"data,omitempty"`
	Muted            bool      `json:"muted,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	Signature        string    `json:"signature"`
}

type JoinRequest struct {
	Invite       Invite   `json:"invite"`
	Requester    PeerInfo `json:"requester"`
	Capabilities []string `json:"capabilities"`
}

type JoinResponse struct {
	Manifest Manifest        `json:"manifest"`
	Server   ServerRecord    `json:"server"`
	Channels []ChannelRecord `json:"channels"`
	History  []MessageRecord `json:"history"`
}

type DrainRequest struct {
	PeerID string `json:"peer_id"`
}

type PeerInfo struct {
	PeerID    string   `json:"peer_id"`
	Role      Role     `json:"role"`
	Addresses []string `json:"addresses"`
	PublicKey string   `json:"public_key"`
}

type ManualPeerRequest struct {
	Address string `json:"address"`
}

type CreateIdentityRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
}

type CreateServerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type JoinServerRequest struct {
	Deeplink string `json:"deeplink"`
}

type CreateChannelRequest struct {
	Name  string `json:"name"`
	Voice bool   `json:"voice"`
}

type SendMessageRequest struct {
	Body string `json:"body"`
}

type EditMessageRequest struct {
	Body string `json:"body"`
}

type CreateDMRequest struct {
	PeerID string `json:"peer_id"`
}

type VoiceJoinRequest struct {
	Muted bool `json:"muted"`
}

type VoiceFrameRequest struct {
	Data []byte `json:"data"`
}

type BackupResponse struct {
	Identity Identity `json:"identity"`
}

func (m Manifest) canonicalBytes() ([]byte, error) {
	copy := m
	copy.Signature = ""
	copy.OwnerAddresses = dedupeSorted(copy.OwnerAddresses)
	copy.BootstrapAddrs = dedupeSorted(copy.BootstrapAddrs)
	copy.RelayAddrs = dedupeSorted(copy.RelayAddrs)
	copy.Capabilities = dedupeSorted(copy.Capabilities)
	return json.Marshal(copy)
}

func (m *Manifest) Sign(identity Identity) error {
	payload, err := m.canonicalBytes()
	if err != nil {
		return err
	}
	sig, err := identity.SignCanonical(payload)
	if err != nil {
		return err
	}
	m.Signature = sig
	return nil
}

func (m Manifest) Verify() error {
	if m.ServerID == "" || m.OwnerPeerID == "" || m.OwnerPublicKey == "" {
		return errors.New("manifest is incomplete")
	}
	payload, err := m.canonicalBytes()
	if err != nil {
		return err
	}
	if err := VerifyCanonical(m.OwnerPublicKey, payload, m.Signature); err != nil {
		return fmt.Errorf("manifest signature: %w", err)
	}
	return nil
}

func (m Manifest) Hash() string {
	payload, _ := m.canonicalBytes()
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	if len(encoded) > 32 {
		return encoded[:32]
	}
	return encoded
}

func (i Invite) canonicalBytes() ([]byte, error) {
	copy := i
	copy.Signature = ""
	copy.ServerAddrs = dedupeSorted(copy.ServerAddrs)
	copy.BootstrapAddrs = dedupeSorted(copy.BootstrapAddrs)
	copy.RelayAddrs = dedupeSorted(copy.RelayAddrs)
	return json.Marshal(copy)
}

func (i *Invite) Sign(identity Identity) error {
	payload, err := i.canonicalBytes()
	if err != nil {
		return err
	}
	sig, err := identity.SignCanonical(payload)
	if err != nil {
		return err
	}
	i.Signature = sig
	return nil
}

func (i Invite) Verify() error {
	if i.ServerID == "" || i.OwnerPublicKey == "" {
		return errors.New("invite is incomplete")
	}
	if !i.ExpiresAt.IsZero() && time.Now().UTC().After(i.ExpiresAt) {
		return errors.New("invite expired")
	}
	payload, err := i.canonicalBytes()
	if err != nil {
		return err
	}
	if err := VerifyCanonical(i.OwnerPublicKey, payload, i.Signature); err != nil {
		return fmt.Errorf("invite signature: %w", err)
	}
	return nil
}

func (i Invite) Deeplink() (string, error) {
	raw, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("invite", base64.RawURLEncoding.EncodeToString(raw))
	return fmt.Sprintf("%s://%s/%s?%s", joinScheme, joinHost, i.ServerID, q.Encode()), nil
}

func ParseDeeplink(raw string) (Invite, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Invite{}, fmt.Errorf("parse deeplink: %w", err)
	}
	if u.Scheme != joinScheme || u.Host != joinHost {
		return Invite{}, errors.New("invalid deeplink scheme or host")
	}
	if strings.TrimSpace(u.Fragment) != "" {
		return Invite{}, errors.New("deeplink fragments are not allowed")
	}
	if len(u.Query()) != 1 || u.Query().Get("invite") == "" {
		return Invite{}, errors.New("deeplink requires only invite query parameter")
	}
	serverID := strings.TrimPrefix(u.Path, "/")
	if serverID == "" {
		return Invite{}, errors.New("deeplink missing server id")
	}
	blob, err := base64.RawURLEncoding.DecodeString(u.Query().Get("invite"))
	if err != nil {
		return Invite{}, fmt.Errorf("decode invite: %w", err)
	}
	var invite Invite
	if err := json.Unmarshal(blob, &invite); err != nil {
		return Invite{}, fmt.Errorf("decode invite payload: %w", err)
	}
	if invite.ServerID != serverID {
		return Invite{}, errors.New("deeplink server id mismatch")
	}
	if err := invite.Verify(); err != nil {
		return Invite{}, err
	}
	return invite, nil
}

func (d Delivery) canonicalBytes() ([]byte, error) {
	copy := d
	copy.Signature = ""
	copy.RecipientPeerIDs = dedupeSorted(copy.RecipientPeerIDs)
	return json.Marshal(copy)
}

func (d *Delivery) Sign(identity Identity) error {
	payload, err := d.canonicalBytes()
	if err != nil {
		return err
	}
	sig, err := identity.SignCanonical(payload)
	if err != nil {
		return err
	}
	d.Signature = sig
	return nil
}

func (d Delivery) Verify() error {
	if d.ID == "" || d.Kind == "" || d.SenderPeerID == "" || d.SenderPublicKey == "" {
		return errors.New("delivery is incomplete")
	}
	payload, err := d.canonicalBytes()
	if err != nil {
		return err
	}
	if err := VerifyCanonical(d.SenderPublicKey, payload, d.Signature); err != nil {
		return fmt.Errorf("delivery signature: %w", err)
	}
	return nil
}

func dedupeSorted[T ~string](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	set := make(map[T]struct{}, len(in))
	for _, item := range in {
		if strings.TrimSpace(string(item)) == "" {
			continue
		}
		set[item] = struct{}{}
	}
	out := make([]T, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

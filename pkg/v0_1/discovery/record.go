// Package discovery implements the layered peer discovery system (spec 31).
package discovery

import "time"

// PeerRecord is a discovered peer advertisement (spec 31 §7).
type PeerRecord struct {
	PeerID    string    `json:"peer_id"`
	Addresses []string  `json:"addresses"`
	Role      string    `json:"role"`
	Caps      []string  `json:"capabilities"`
	LastSeen  int64     `json:"last_seen"` // unix milliseconds
	Source    string    `json:"source"`    // "mdns", "dht", "bootstrap", "pex", "manual"
	ExpiresAt time.Time `json:"-"`

	// Signature fields (spec 31 §7.1 — hybrid signing).
	SigningPublicKey []byte `json:"signing_public_key,omitempty"` // Ed25519 public key (32 bytes)
	MLDSA65PublicKey []byte `json:"mldsa65_public_key,omitempty"` // ML-DSA-65 public key (1952 bytes)
	SignedAt         int64  `json:"signed_at,omitempty"`          // unix seconds
	Signature        string `json:"signature,omitempty"`          // base64url-no-pad hybrid sig
}

// IsExpired reports whether this record's TTL has passed.
func (r *PeerRecord) IsExpired() bool {
	return !r.ExpiresAt.IsZero() && time.Now().After(r.ExpiresAt)
}

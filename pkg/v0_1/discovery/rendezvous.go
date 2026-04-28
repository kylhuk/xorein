package discovery

import (
	"crypto/sha256"
	"encoding/hex"
)

// ServerRendezvousCID computes the Kademlia CID for a server's rendezvous point
// per spec 31 §3.5: SHA-256("xorein/server/" || server_id), hex-encoded.
func ServerRendezvousCID(serverID string) string {
	h := sha256.Sum256([]byte("xorein/server/" + serverID))
	return hex.EncodeToString(h[:])
}

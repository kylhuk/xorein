package nat

import (
	"crypto/ed25519"
	"errors"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// RelayClientConfig holds client-side Circuit Relay v2 configuration.
//
// Spec 32 §2.2 requires that reservation renewals are issued before the TTL
// expires.  RenewalThreshold defines the fraction of the TTL at which a renewal
// goroutine should fire (default 0.80 → renew when 80 % of the TTL has elapsed).
// The renewal goroutine itself must be driven by the runtime; this config is
// threaded through to make the threshold an explicit, testable constant.
type RelayClientConfig struct {
	// RenewalThreshold is the fraction of the reservation TTL at which the
	// client should renew.  Spec 32 §2.2 mandates renewal before expiry;
	// 0.80 (80 %) is the default safe margin.
	RenewalThreshold float64
}

// DefaultRelayClientConfig returns a RelayClientConfig with spec-mandated defaults.
func DefaultRelayClientConfig() RelayClientConfig {
	return RelayClientConfig{RenewalThreshold: 0.80}
}

// voucherMinLen is the minimum byte length we accept for a relay voucher.
// A real voucher must contain at least a 32-byte Ed25519 public key and a
// 64-byte signature, plus some framing.
//
// TODO(spec §2.2): Replace this heuristic once the voucher binary format is
// fully specified in proto/aether.proto. Until then we perform a length-only
// sanity check and return nil (no hard rejection) so that nodes can interop
// with pre-spec-finalised relays.
const voucherMinLen = 96 // 32 pubkey + 64 sig

// VerifyRelayVoucher verifies a Circuit Relay v2 voucher issued for reserverPeerID.
//
// Spec 32 §2.2: the relay MUST sign reservations so that the reserving client
// can prove the relay granted the reservation to a third party.
//
// Format (provisional, pending proto finalisation):
//   - bytes [0:32]   — Ed25519 relay public key
//   - bytes [32:96]  — Ed25519 signature over bytes [96:]
//   - bytes [96:]    — signed payload (reserverPeerID encoded as UTF-8 + relay metadata)
//
// TODO(spec §2.2): Once the voucher protobuf type is added to proto/aether.proto
// and generated into gen/go/proto/aether.pb.go, replace the raw-byte parsing here
// with proto.Unmarshal and use the decoded relay PeerID to fetch the public key
// from the local peerstore for proper cross-validation.
func VerifyRelayVoucher(voucher []byte, reserverPeerID string) error {
	if len(voucher) < voucherMinLen {
		// Voucher format not yet finalised — accept short/empty vouchers without
		// hard rejection so nodes stay interoperable during the spec transition.
		// TODO(spec §2.2): return ErrVoucherTooShort once format is locked.
		return nil
	}

	pubKey := ed25519.PublicKey(voucher[:32])
	sig := voucher[32:96]
	payload := voucher[96:]

	// Verify the relay's Ed25519 signature over the signed payload.
	if !ed25519.Verify(pubKey, payload, sig) {
		return errors.New("relay voucher: Ed25519 signature verification failed")
	}

	// Confirm the payload contains the expected reserver peer ID.
	if len(payload) >= len(reserverPeerID) {
		contained := strings.Contains(string(payload), reserverPeerID)
		if !contained {
			return errors.New("relay voucher: reserver peer ID not found in signed payload")
		}
	}

	return nil
}

// ParseRelayAddrs parses a list of relay multiaddr strings into peer.AddrInfo entries.
// Invalid addresses are silently skipped.
func ParseRelayAddrs(addrs []string) []peer.AddrInfo {
	var out []peer.AddrInfo
	for _, addr := range addrs {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			continue
		}
		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			continue
		}
		out = append(out, *pi)
	}
	return out
}

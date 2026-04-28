package envelope

import (
	"encoding/binary"
	"fmt"
	"sort"

	"google.golang.org/protobuf/proto"
)

// BuildCanonicalPayload serialises the canonical payload bytes that are signed
// per spec 02 §2.1:
//
//	canonical_payload = varint(payload_type) || deterministic_proto(payload) || uint64BE(signed_at_ms)
//
// signed_at_ms is unix milliseconds. The output is stored in SignedEnvelope.canonical_payload
// and is what both Ed25519 and ML-DSA-65 sign.
func BuildCanonicalPayload(payloadType uint32, msg proto.Message, signedAtMS int64) ([]byte, error) {
	// Encode payload_type as a protobuf varint.
	var varintBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(varintBuf[:], uint64(payloadType))

	// Deterministic protobuf serialization.
	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("canonical payload: marshal: %w", err)
	}

	// uint64 big-endian signed_at.
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(signedAtMS))

	// Concatenate.
	out := make([]byte, n+len(b)+8)
	copy(out[:n], varintBuf[:n])
	copy(out[n:], b)
	copy(out[n+len(b):], ts[:])
	return out, nil
}

// sortStringSliceField sorts a []any (string) slice stored at key in m, if present.
func sortStringSliceField(m map[string]any, key string) {
	raw, ok := m[key]
	if !ok {
		return
	}
	slice, ok := raw.([]any)
	if !ok {
		return
	}
	sort.Slice(slice, func(i, j int) bool {
		si, oki := slice[i].(string)
		sj, okj := slice[j].(string)
		if !oki || !okj {
			return false
		}
		return si < sj
	})
	m[key] = slice
}

// CanonicalManifestJSON produces the canonical JSON for a Manifest per spec 02 §3.1.
//
// The following slice fields are sorted lexicographically before serialisation:
// owner_addresses, bootstrap_addrs, relay_addrs, capabilities.
// offered_security_modes is NOT sorted (order is preserved).
// The signature key is excluded from the output.
func CanonicalManifestJSON(manifest map[string]any) ([]byte, error) {
	// Deep-copy to avoid mutating the caller's map.
	m := make(map[string]any, len(manifest))
	for k, v := range manifest {
		m[k] = v
	}
	// Exclude signature key.
	delete(m, "signature")
	// Sort the specified array fields.
	for _, field := range []string{"owner_addresses", "bootstrap_addrs", "relay_addrs", "capabilities"} {
		sortStringSliceField(m, field)
	}
	return CanonicalJSON(m)
}

// CanonicalInviteJSON produces the canonical JSON for an Invite per spec 02 §3.2.
//
// The capabilities slice is sorted lexicographically.
// The signature key is excluded from the output.
func CanonicalInviteJSON(invite map[string]any) ([]byte, error) {
	m := make(map[string]any, len(invite))
	for k, v := range invite {
		m[k] = v
	}
	delete(m, "signature")
	sortStringSliceField(m, "capabilities")
	return CanonicalJSON(m)
}

// CanonicalDeliveryJSON produces the canonical JSON for a Delivery per spec 02 §3.3.
//
// The signature, data, and muted keys are excluded from the output.
// The capabilities slice, if present, is sorted lexicographically.
func CanonicalDeliveryJSON(delivery map[string]any) ([]byte, error) {
	m := make(map[string]any, len(delivery))
	for k, v := range delivery {
		m[k] = v
	}
	delete(m, "signature")
	delete(m, "data")
	delete(m, "muted")
	sortStringSliceField(m, "capabilities")
	return CanonicalJSON(m)
}

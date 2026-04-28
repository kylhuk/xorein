package envelope

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
)

// CanonicalJSON serialises v as deterministic JSON with sorted object keys,
// no whitespace, and UTF-8 encoding. Byte slice fields must already be encoded
// as base64url-no-padding strings by the caller.
// Source: spec 02 §3 — used for Manifest, Invite, Delivery signing.
func CanonicalJSON(v any) ([]byte, error) {
	// Marshal to standard JSON first, then round-trip through map to sort keys.
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonical json: marshal: %w", err)
	}
	var m any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("canonical json: unmarshal: %w", err)
	}
	return sortedJSON(m)
}

func sortedJSON(v any) ([]byte, error) {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			buf.Write(kb)
			buf.WriteByte(':')
			vb, err := sortedJSON(val[k])
			if err != nil {
				return nil, err
			}
			buf.Write(vb)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	case []any:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			ib, err := sortedJSON(item)
			if err != nil {
				return nil, err
			}
			buf.Write(ib)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	default:
		return json.Marshal(val)
	}
}

// EncodeHybridSig returns the base64url-no-padding encoding of Ed25519_sig || ML-DSA-65_sig.
// Source: spec 01 §6.1.
func EncodeHybridSig(edSig, mldsaSig []byte) string {
	combined := make([]byte, len(edSig)+len(mldsaSig))
	copy(combined, edSig)
	copy(combined[len(edSig):], mldsaSig)
	return base64.RawURLEncoding.EncodeToString(combined)
}

// DecodeHybridSig decodes a base64url-no-padding combined signature and splits
// it at byte 64 (Ed25519 sig) + remainder (ML-DSA-65 sig).
func DecodeHybridSig(s string) (edSig, mldsaSig []byte, err error) {
	combined, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, nil, fmt.Errorf("decode hybrid sig: %w", err)
	}
	if len(combined) < 64 {
		return nil, nil, fmt.Errorf("decode hybrid sig: too short (%d bytes)", len(combined))
	}
	return combined[:64], combined[64:], nil
}

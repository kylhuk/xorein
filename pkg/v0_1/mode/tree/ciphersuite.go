// Package tree — ciphersuite constants for the Xorein hybrid MLS ciphersuite.
// Source: docs/spec/v0.1/12-mode-tree.md §1
package tree

// CiphersuiteID is the Xorein private-use hybrid MLS ciphersuite per RFC 9420 §17.1.
const CiphersuiteID = uint16(0xFF01)

// CiphersuiteLabel is the IANA-style human-readable ciphersuite name.
const CiphersuiteLabel = "XoreinMLS_128_HYBRID_DHKEMX25519MLKEM768_AES128GCM_SHA256_Ed25519MLDSA65"

// MLSVersion is the MLS protocol version supported.
const MLSVersion = uint8(1)

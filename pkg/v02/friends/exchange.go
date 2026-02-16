package friends

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const (
	publicKeyShareNamespace = "aether-friendkey"
	publicKeyShareVersion   = "v1"
	qrPayloadPrefix         = "aether-friend-qr:"
	deepLinkScheme          = "aether"
	deepLinkHost            = "friend"
)

type ExchangeReason string

const (
	ExchangeReasonValid            ExchangeReason = "exchange-valid"
	ExchangeReasonInvalidFormat    ExchangeReason = "exchange-invalid-format"
	ExchangeReasonInvalidEncoding  ExchangeReason = "exchange-invalid-encoding"
	ExchangeReasonChecksumMismatch ExchangeReason = "exchange-checksum-mismatch"
	ExchangeReasonMissingSignature ExchangeReason = "exchange-missing-signature"
	ExchangeReasonExpired          ExchangeReason = "exchange-expired"
	ExchangeReasonInvalidScheme    ExchangeReason = "exchange-invalid-scheme"
	ExchangeReasonMissingPayload   ExchangeReason = "exchange-missing-payload"
)

type ExchangePath string

const (
	ExchangePathPublicKey ExchangePath = "public-key"
	ExchangePathQR        ExchangePath = "qr"
	ExchangePathDeepLink  ExchangePath = "deep-link"
)

type RequestSignatureAlgorithm string

const (
	RequestSignatureAlgorithmUnspecified RequestSignatureAlgorithm = "unspecified"
	RequestSignatureAlgorithmEd25519     RequestSignatureAlgorithm = "ed25519"
)

type PublicKeyShare struct {
	IdentityID string
	PublicKey  []byte
	Checksum   string
	Encoded    string
}

type InvitePayload struct {
	Version            string                    `json:"version"`
	IdentityID         string                    `json:"identity_id"`
	PublicKey          string                    `json:"public_key"`
	PublicKeyChecksum  string                    `json:"public_key_checksum"`
	Nonce              string                    `json:"nonce"`
	ExpiresAtUnix      int64                     `json:"expires_at_unix"`
	SignatureAlgorithm RequestSignatureAlgorithm `json:"signature_algorithm"`
	Signature          string                    `json:"signature"`
}

type CanonicalRequestInput struct {
	Path               ExchangePath
	IdentityID         string
	PublicKey          []byte
	PublicKeyChecksum  string
	Nonce              string
	ExpiresAtUnix      int64
	SignatureAlgorithm RequestSignatureAlgorithm
	Signature          []byte
}

func EncodePublicKeyShare(identityID string, publicKey []byte) (string, ExchangeReason) {
	if identityID == "" || len(publicKey) == 0 {
		return "", ExchangeReasonInvalidFormat
	}
	key := base64.RawURLEncoding.EncodeToString(publicKey)
	checksum := computePublicKeyChecksum(publicKey)
	encoded := strings.Join([]string{publicKeyShareNamespace, publicKeyShareVersion, identityID, key, checksum}, ":")
	return encoded, ExchangeReasonValid
}

func ParsePublicKeyShare(raw string) (PublicKeyShare, ExchangeReason) {
	parts := strings.Split(raw, ":")
	if len(parts) != 5 || parts[0] != publicKeyShareNamespace || parts[1] != publicKeyShareVersion || parts[2] == "" {
		return PublicKeyShare{}, ExchangeReasonInvalidFormat
	}
	keyBytes, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return PublicKeyShare{}, ExchangeReasonInvalidEncoding
	}
	checksum := computePublicKeyChecksum(keyBytes)
	if checksum != parts[4] {
		return PublicKeyShare{}, ExchangeReasonChecksumMismatch
	}
	return PublicKeyShare{
		IdentityID: parts[2],
		PublicKey:  keyBytes,
		Checksum:   checksum,
		Encoded:    raw,
	}, ExchangeReasonValid
}

func BuildInvitePayload(identityID string, publicKey []byte, nonce string, expiresAtUnix int64, algorithm RequestSignatureAlgorithm, signature []byte) (InvitePayload, ExchangeReason) {
	if identityID == "" || len(publicKey) == 0 || nonce == "" || expiresAtUnix <= 0 {
		return InvitePayload{}, ExchangeReasonInvalidFormat
	}
	if algorithm == RequestSignatureAlgorithmUnspecified || len(signature) == 0 {
		return InvitePayload{}, ExchangeReasonMissingSignature
	}
	return InvitePayload{
		Version:            publicKeyShareVersion,
		IdentityID:         identityID,
		PublicKey:          base64.RawURLEncoding.EncodeToString(publicKey),
		PublicKeyChecksum:  computePublicKeyChecksum(publicKey),
		Nonce:              nonce,
		ExpiresAtUnix:      expiresAtUnix,
		SignatureAlgorithm: algorithm,
		Signature:          base64.RawURLEncoding.EncodeToString(signature),
	}, ExchangeReasonValid
}

func EncodeQRPayload(payload InvitePayload) (string, ExchangeReason) {
	if _, reason := canonicalizeInvitePayload(payload, time.Unix(payload.ExpiresAtUnix, 0), false); reason != ExchangeReasonValid {
		return "", reason
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", ExchangeReasonInvalidFormat
	}
	return qrPayloadPrefix + base64.RawURLEncoding.EncodeToString(blob), ExchangeReasonValid
}

func EncodeDeepLinkPayload(payload InvitePayload) (string, ExchangeReason) {
	if _, reason := canonicalizeInvitePayload(payload, time.Unix(payload.ExpiresAtUnix, 0), false); reason != ExchangeReasonValid {
		return "", reason
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", ExchangeReasonInvalidFormat
	}
	values := url.Values{}
	values.Set("payload", base64.RawURLEncoding.EncodeToString(blob))
	return (&url.URL{Scheme: deepLinkScheme, Host: deepLinkHost, RawQuery: values.Encode()}).String(), ExchangeReasonValid
}

func ParseInvitePayload(raw string, now time.Time) (CanonicalRequestInput, ExchangeReason) {
	payloadBlob, path, reason := extractInvitePayloadBlob(raw)
	if reason != ExchangeReasonValid {
		return CanonicalRequestInput{}, reason
	}
	var payload InvitePayload
	if err := json.Unmarshal(payloadBlob, &payload); err != nil {
		return CanonicalRequestInput{}, ExchangeReasonInvalidFormat
	}
	canonical, reason := canonicalizeInvitePayload(payload, now, true)
	if reason != ExchangeReasonValid {
		return CanonicalRequestInput{}, reason
	}
	canonical.Path = path
	return canonical, ExchangeReasonValid
}

func canonicalizeInvitePayload(payload InvitePayload, now time.Time, checkExpiry bool) (CanonicalRequestInput, ExchangeReason) {
	if payload.Version != publicKeyShareVersion || payload.IdentityID == "" || payload.Nonce == "" || payload.ExpiresAtUnix <= 0 {
		return CanonicalRequestInput{}, ExchangeReasonInvalidFormat
	}
	if payload.SignatureAlgorithm == RequestSignatureAlgorithmUnspecified || payload.Signature == "" {
		return CanonicalRequestInput{}, ExchangeReasonMissingSignature
	}
	publicKey, err := base64.RawURLEncoding.DecodeString(payload.PublicKey)
	if err != nil {
		return CanonicalRequestInput{}, ExchangeReasonInvalidEncoding
	}
	if computePublicKeyChecksum(publicKey) != payload.PublicKeyChecksum {
		return CanonicalRequestInput{}, ExchangeReasonChecksumMismatch
	}
	signature, err := base64.RawURLEncoding.DecodeString(payload.Signature)
	if err != nil {
		return CanonicalRequestInput{}, ExchangeReasonInvalidEncoding
	}
	if checkExpiry && now.Unix() > payload.ExpiresAtUnix {
		return CanonicalRequestInput{}, ExchangeReasonExpired
	}
	return CanonicalRequestInput{
		IdentityID:         payload.IdentityID,
		PublicKey:          publicKey,
		PublicKeyChecksum:  payload.PublicKeyChecksum,
		Nonce:              payload.Nonce,
		ExpiresAtUnix:      payload.ExpiresAtUnix,
		SignatureAlgorithm: payload.SignatureAlgorithm,
		Signature:          signature,
	}, ExchangeReasonValid
}

func extractInvitePayloadBlob(raw string) ([]byte, ExchangePath, ExchangeReason) {
	if strings.HasPrefix(raw, qrPayloadPrefix) {
		blob, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(raw, qrPayloadPrefix))
		if err != nil {
			return nil, "", ExchangeReasonInvalidEncoding
		}
		return blob, ExchangePathQR, ExchangeReasonValid
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, "", ExchangeReasonInvalidFormat
	}
	if u.Scheme != deepLinkScheme || u.Host != deepLinkHost {
		return nil, "", ExchangeReasonInvalidScheme
	}
	encoded := u.Query().Get("payload")
	if encoded == "" {
		return nil, "", ExchangeReasonMissingPayload
	}
	blob, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", ExchangeReasonInvalidEncoding
	}
	return blob, ExchangePathDeepLink, ExchangeReasonValid
}

func computePublicKeyChecksum(publicKey []byte) string {
	sum := sha256.Sum256(publicKey)
	return hex.EncodeToString(sum[:8])
}

package friends

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPublicKeyShareRoundTrip(t *testing.T) {
	publicKey := []byte{1, 2, 3, 4, 5}
	encoded, reason := EncodePublicKeyShare("alice", publicKey)
	if reason != ExchangeReasonValid {
		t.Fatalf("EncodePublicKeyShare reason=%q", reason)
	}
	parsed, reason := ParsePublicKeyShare(encoded)
	if reason != ExchangeReasonValid {
		t.Fatalf("ParsePublicKeyShare reason=%q", reason)
	}
	if parsed.IdentityID != "alice" || !bytes.Equal(parsed.PublicKey, publicKey) || parsed.Encoded != encoded {
		t.Fatalf("unexpected parsed share: %+v", parsed)
	}
}

func TestParsePublicKeyShareFailures(t *testing.T) {
	encoded, _ := EncodePublicKeyShare("alice", []byte{1, 2, 3})
	parts := strings.Split(encoded, ":")
	parts[4] = "0000000000000000"
	checksumMismatch := strings.Join(parts, ":")

	tests := []struct {
		name   string
		raw    string
		reason ExchangeReason
	}{
		{name: "invalid format", raw: "aether-friendkey:v1:alice", reason: ExchangeReasonInvalidFormat},
		{name: "invalid encoding", raw: "aether-friendkey:v1:alice:@@@@:abcd", reason: ExchangeReasonInvalidEncoding},
		{name: "checksum mismatch", raw: checksumMismatch, reason: ExchangeReasonChecksumMismatch},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, reason := ParsePublicKeyShare(tc.raw)
			if reason != tc.reason {
				t.Fatalf("ParsePublicKeyShare reason=%q want %q", reason, tc.reason)
			}
		})
	}
}

func TestQRAndDeepLinkProduceEquivalentCanonicalInput(t *testing.T) {
	now := time.Unix(1000, 0)
	payload, reason := BuildInvitePayload("alice", []byte{9, 8, 7, 6}, "nonce-1", now.Add(time.Hour).Unix(), RequestSignatureAlgorithmEd25519, []byte{4, 5, 6})
	if reason != ExchangeReasonValid {
		t.Fatalf("BuildInvitePayload reason=%q", reason)
	}
	qr, reason := EncodeQRPayload(payload)
	if reason != ExchangeReasonValid {
		t.Fatalf("EncodeQRPayload reason=%q", reason)
	}
	link, reason := EncodeDeepLinkPayload(payload)
	if reason != ExchangeReasonValid {
		t.Fatalf("EncodeDeepLinkPayload reason=%q", reason)
	}

	fromQR, reason := ParseInvitePayload(qr, now)
	if reason != ExchangeReasonValid {
		t.Fatalf("ParseInvitePayload(qr) reason=%q", reason)
	}
	fromLink, reason := ParseInvitePayload(link, now)
	if reason != ExchangeReasonValid {
		t.Fatalf("ParseInvitePayload(link) reason=%q", reason)
	}

	if fromQR.Path != ExchangePathQR || fromLink.Path != ExchangePathDeepLink {
		t.Fatalf("unexpected paths qr=%q link=%q", fromQR.Path, fromLink.Path)
	}
	if fromQR.IdentityID != fromLink.IdentityID || fromQR.Nonce != fromLink.Nonce || fromQR.ExpiresAtUnix != fromLink.ExpiresAtUnix || fromQR.PublicKeyChecksum != fromLink.PublicKeyChecksum || fromQR.SignatureAlgorithm != fromLink.SignatureAlgorithm || !bytes.Equal(fromQR.PublicKey, fromLink.PublicKey) || !bytes.Equal(fromQR.Signature, fromLink.Signature) {
		t.Fatalf("canonical mismatch qr=%+v link=%+v", fromQR, fromLink)
	}
}

func TestInvitePayloadValidationFailures(t *testing.T) {
	now := time.Unix(1000, 0)
	payload, reason := BuildInvitePayload("alice", []byte{9, 8, 7, 6}, "nonce-1", now.Add(time.Hour).Unix(), RequestSignatureAlgorithmEd25519, []byte{4, 5, 6})
	if reason != ExchangeReasonValid {
		t.Fatalf("BuildInvitePayload reason=%q", reason)
	}

	missingSignature := payload
	missingSignature.Signature = ""
	if _, reason = EncodeQRPayload(missingSignature); reason != ExchangeReasonMissingSignature {
		t.Fatalf("EncodeQRPayload(missing sig) reason=%q", reason)
	}

	expired := payload
	expired.ExpiresAtUnix = now.Add(-time.Second).Unix()
	qr, reason := EncodeQRPayload(expired)
	if reason != ExchangeReasonValid {
		t.Fatalf("EncodeQRPayload(expired) reason=%q", reason)
	}
	if _, reason = ParseInvitePayload(qr, now); reason != ExchangeReasonExpired {
		t.Fatalf("ParseInvitePayload(expired) reason=%q", reason)
	}

	if _, reason = ParseInvitePayload("https://example.com", now); reason != ExchangeReasonInvalidScheme {
		t.Fatalf("ParseInvitePayload(invalid scheme) reason=%q", reason)
	}
	if _, reason = ParseInvitePayload("aether://friend", now); reason != ExchangeReasonMissingPayload {
		t.Fatalf("ParseInvitePayload(missing payload) reason=%q", reason)
	}
}

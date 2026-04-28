package mediashield_test

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"

	ms "github.com/aether/code_aether/pkg/v0_1/mode/mediashield"
	"github.com/aether/code_aether/pkg/v0_1/spectest"
)

func vectorDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test path")
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
	return filepath.Join(root, "docs", "spec", "v0.1", "91-test-vectors")
}

func TestPins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

// TestW4MediaShieldNonce drives the mediashield nonce KAT.
func TestW4MediaShieldNonce(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "mediashield_nonce.json"))

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			ikm := spectest.Hex(v.Inputs["ikm"])
			salt := spectest.Hex(v.Inputs["salt"])
			expected := spectest.Hex(v.ExpectedOutput["nonce"])

			// frameCounter from salt (8 bytes BE)
			var ctr uint64
			for i, b := range salt {
				ctr = (ctr << 8) | uint64(b)
				_ = i
			}
			got, err := ms.DeriveNonce(ikm, ctr)
			if err != nil {
				t.Fatalf("DeriveNonce: %v", err)
			}
			if !bytes.Equal(got, expected) {
				t.Fatalf("nonce mismatch:\n  want %x\n  got  %x", expected, got)
			}
		})
	}
}

// TestW4MediaShieldSFrame drives the SFrame encrypt KAT.
func TestW4MediaShieldSFrame(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "mediashield_sframe.json"))

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			key := spectest.Hex(v.Inputs["mediashield_key"])
			rtpHeader := spectest.Hex(v.Inputs["rtp_header_bytes"])
			plaintext := spectest.Hex(v.Inputs["media_frame_plaintext"])
			peerID := "alice@xorein"
			expectedCT := spectest.Hex(v.ExpectedOutput["ciphertext_with_tag"])
			expectedPayload := spectest.Hex(v.ExpectedOutput["sframe_payload"])

			pk := &ms.PeerKey{PeerID: peerID, Key: key, FrameCounter: 0}
			hdr, ct, err := ms.EncryptFrame(pk, rtpHeader, plaintext)
			if err != nil {
				t.Fatalf("EncryptFrame: %v", err)
			}
			if !bytes.Equal(ct, expectedCT) {
				t.Fatalf("ciphertext_with_tag mismatch:\n  want %x\n  got  %x", expectedCT, ct)
			}
			payload := append(hdr, ct...)
			if !bytes.Equal(payload, expectedPayload) {
				t.Fatalf("sframe_payload mismatch:\n  want %x\n  got  %x", expectedPayload, payload)
			}

			// Verify decrypt round-trip.
			pk2 := &ms.PeerKey{PeerID: peerID, Key: key}
			got, err := ms.DecryptFrame(pk2, rtpHeader, hdr, ct)
			if err != nil {
				t.Fatalf("DecryptFrame: %v", err)
			}
			if !bytes.Equal(got, plaintext) {
				t.Fatalf("decrypt mismatch: want %x got %x", plaintext, got)
			}
		})
	}
}

// TestW4MediaShieldRoundtrip verifies encrypt/decrypt round-trip with frame counter increment.
func TestW4MediaShieldRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = 0xAB
	}
	pk := &ms.PeerKey{PeerID: "bob@xorein", Key: key}
	pkDecrypt := &ms.PeerKey{PeerID: "bob@xorein", Key: key}

	for i := 0; i < 5; i++ {
		pt := []byte("frame payload " + string(rune('0'+i)))
		rtp := []byte{0x80, 0x00, byte(i), 0x01}
		hdr, ct, err := ms.EncryptFrame(pk, rtp, pt)
		if err != nil {
			t.Fatalf("EncryptFrame %d: %v", i, err)
		}
		got, err := ms.DecryptFrame(pkDecrypt, rtp, hdr, ct)
		if err != nil {
			t.Fatalf("DecryptFrame %d: %v", i, err)
		}
		if !bytes.Equal(got, pt) {
			t.Fatalf("frame %d mismatch: want %q got %q", i, pt, got)
		}
	}
}

// TestW4MediaShieldPeerKID verifies KID derivation.
func TestW4MediaShieldPeerKID(t *testing.T) {
	expected := spectest.Hex("176cd6a27cee7df1")
	got := ms.PeerKID("alice@xorein")
	if !bytes.Equal(got, expected) {
		t.Fatalf("PeerKID mismatch:\n  want %x\n  got  %x", expected, got)
	}
}

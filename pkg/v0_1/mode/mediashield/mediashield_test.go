package mediashield_test

import (
	"bytes"
	"testing"

	ms "github.com/aether/code_aether/pkg/v0_1/mode/mediashield"
)

// TestDecryptFrameKIDValidation verifies that DecryptFrame rejects frames whose
// SFrame KID does not match the PeerKey.PeerID.
func TestDecryptFrameKIDValidation(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = 0xAB
	}

	pkEnc := &ms.PeerKey{PeerID: "alice@xorein", Key: key}
	// Encrypt as alice.
	rtp := []byte{0x80, 0x01, 0x00, 0x01}
	pt := []byte("test frame")
	hdr, ct, err := ms.EncryptFrame(pkEnc, rtp, pt)
	if err != nil {
		t.Fatalf("EncryptFrame: %v", err)
	}

	// Decrypt as alice (correct KID) → should succeed.
	pkDec := &ms.PeerKey{PeerID: "alice@xorein", Key: key}
	if _, err := ms.DecryptFrame(pkDec, rtp, hdr, ct); err != nil {
		t.Fatalf("DecryptFrame (correct KID): %v", err)
	}

	// Decrypt as bob (wrong KID) → should return ErrKIDMismatch.
	pkBob := &ms.PeerKey{PeerID: "bob@xorein", Key: key}
	_, err = ms.DecryptFrame(pkBob, rtp, hdr, ct)
	if err != ms.ErrKIDMismatch {
		t.Fatalf("want ErrKIDMismatch, got %v", err)
	}
}

// TestDecryptFrameReplayProtection verifies that frames with a counter ≤ the
// highest decrypted counter are rejected with ErrReplay.
func TestDecryptFrameReplayProtection(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = 0xCD
	}
	peerID := "relay-peer@xorein"
	rtp := []byte{0x80, 0x01}

	// Encrypt frames 0, 1, 2.
	pkEnc := &ms.PeerKey{PeerID: peerID, Key: key}
	var hdrs [3][]byte
	var cts [3][]byte
	for i := 0; i < 3; i++ {
		h, c, err := ms.EncryptFrame(pkEnc, rtp, []byte{byte(i)})
		if err != nil {
			t.Fatalf("EncryptFrame %d: %v", i, err)
		}
		hdrs[i] = h
		cts[i] = c
	}

	pkDec := &ms.PeerKey{PeerID: peerID, Key: key}

	// Decrypt frame 0.
	if _, err := ms.DecryptFrame(pkDec, rtp, hdrs[0], cts[0]); err != nil {
		t.Fatalf("DecryptFrame 0: %v", err)
	}

	// Decrypt frame 2 (out-of-order but forward, should be fine).
	if _, err := ms.DecryptFrame(pkDec, rtp, hdrs[2], cts[2]); err != nil {
		t.Fatalf("DecryptFrame 2: %v", err)
	}

	// Replay frame 0 → should be rejected.
	_, err := ms.DecryptFrame(pkDec, rtp, hdrs[0], cts[0])
	if err != ms.ErrReplay {
		t.Fatalf("want ErrReplay for replayed frame 0, got %v", err)
	}

	// Replay frame 2 → should also be rejected.
	_, err = ms.DecryptFrame(pkDec, rtp, hdrs[2], cts[2])
	if err != ms.ErrReplay {
		t.Fatalf("want ErrReplay for replayed frame 2, got %v", err)
	}
}

// TestDecryptFrameFirstFrameAllowed verifies the first frame (counter=0) is
// accepted even when HasDecrypted is false.
func TestDecryptFrameFirstFrameAllowed(t *testing.T) {
	key := make([]byte, 32)
	peerID := "first@xorein"
	rtp := []byte{0x00}

	pkEnc := &ms.PeerKey{PeerID: peerID, Key: key}
	hdr, ct, err := ms.EncryptFrame(pkEnc, rtp, []byte("hello"))
	if err != nil {
		t.Fatalf("EncryptFrame: %v", err)
	}

	pkDec := &ms.PeerKey{PeerID: peerID, Key: key}
	got, err := ms.DecryptFrame(pkDec, rtp, hdr, ct)
	if err != nil {
		t.Fatalf("DecryptFrame first frame: %v", err)
	}
	if !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("plaintext mismatch: %q", got)
	}
}

// TestDecryptFrameCounterStateUpdated verifies MaxDecryptedCounter and HasDecrypted
// are updated after successful decryption.
func TestDecryptFrameCounterStateUpdated(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = 0xEF
	}
	peerID := "state@xorein"
	rtp := []byte{0x01}

	pkEnc := &ms.PeerKey{PeerID: peerID, Key: key, FrameCounter: 5}
	hdr, ct, err := ms.EncryptFrame(pkEnc, rtp, []byte("state check"))
	if err != nil {
		t.Fatalf("EncryptFrame: %v", err)
	}
	if pkEnc.FrameCounter != 6 {
		t.Fatalf("EncryptFrame should increment counter: want 6 got %d", pkEnc.FrameCounter)
	}

	pkDec := &ms.PeerKey{PeerID: peerID, Key: key}
	if pkDec.HasDecrypted {
		t.Fatal("HasDecrypted should start as false")
	}
	if _, err := ms.DecryptFrame(pkDec, rtp, hdr, ct); err != nil {
		t.Fatalf("DecryptFrame: %v", err)
	}
	if !pkDec.HasDecrypted {
		t.Fatal("HasDecrypted should be true after first decrypt")
	}
	if pkDec.MaxDecryptedCounter != 5 {
		t.Fatalf("MaxDecryptedCounter: want 5, got %d", pkDec.MaxDecryptedCounter)
	}
}

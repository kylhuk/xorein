package mediashield_test

import (
	"testing"

	ms "github.com/aether/code_aether/pkg/v0_1/mode/mediashield"
)

// FuzzSFrameDecrypt feeds arbitrary bytes to SFrame decrypt to verify no panics.
func FuzzSFrameDecrypt(f *testing.F) {
	// Build a valid key and encrypt one frame to seed the corpus.
	key := make([]byte, 32)
	for i := range key {
		key[i] = 0xAB
	}
	pk := &ms.PeerKey{PeerID: "alice@xorein", Key: key, FrameCounter: 0}
	rtpHeader := []byte{0x80, 0x00, 0x01, 0x00}
	plaintext := []byte("fuzz seed frame")

	hdr, ct, err := ms.EncryptFrame(pk, rtpHeader, plaintext)
	if err != nil {
		f.Fatal(err)
	}

	// Seed with valid inputs.
	f.Add(key, rtpHeader, hdr, ct)
	// Seed with some invalid ciphertexts.
	f.Add(key, rtpHeader, hdr, []byte{})
	f.Add(key, rtpHeader, hdr, []byte{0x00, 0x01, 0x02, 0x03})
	f.Add(key, []byte{}, hdr, ct)
	f.Add(make([]byte, 32), rtpHeader, hdr, ct)

	f.Fuzz(func(t *testing.T, keyBytes []byte, rtpHdr []byte, sframeHdr []byte, ciphertextBytes []byte) {
		if len(keyBytes) != 32 {
			// PeerKey requires a 32-byte key.
			return
		}
		pk2 := &ms.PeerKey{PeerID: "alice@xorein", Key: keyBytes}
		// Must not panic — only return an error.
		_, _ = ms.DecryptFrame(pk2, rtpHdr, sframeHdr, ciphertextBytes)
	})
}

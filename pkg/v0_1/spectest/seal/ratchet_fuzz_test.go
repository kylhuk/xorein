package seal_test

import (
	"testing"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	seal "github.com/aether/code_aether/pkg/v0_1/mode/seal"
)

// FuzzDoubleRatchetDecrypt feeds arbitrary ciphertext to the ratchet decrypt path
// to verify it never panics (only returns errors).
func FuzzDoubleRatchetDecrypt(f *testing.F) {
	// Seed corpus: a valid ciphertext from a known ratchet state.
	ikm := make([]byte, 32)
	ikm[31] = 0x42
	okm, err := v0crypto.DeriveKey(ikm, nil, v0crypto.LabelSealRootKey, 64)
	if err != nil {
		f.Fatal(err)
	}
	rPriv, rPub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		f.Fatal(err)
	}

	sender := &seal.RatchetState{SkipList: make(map[seal.SkipKey][32]byte)}
	copy(sender.RootKey[:], okm[:32])
	copy(sender.SendChainKey[:], okm[32:])
	sender.SendRatchetPriv = rPriv
	sender.SendRatchetPub = rPub

	receiver := &seal.RatchetState{SkipList: make(map[seal.SkipKey][32]byte)}
	copy(receiver.RootKey[:], okm[:32])
	copy(receiver.RecvChainKey[:], okm[32:])
	receiver.RemoteRatchetPub = rPub

	hdr, ct, err := seal.Encrypt(sender, []byte("seed message"))
	if err != nil {
		f.Fatal(err)
	}

	// Add valid ciphertext as corpus seed.
	f.Add(hdr[:], ct)
	// Add some invalid seeds.
	f.Add(hdr[:], []byte{})
	f.Add(hdr[:], []byte{0x00, 0x01, 0x02})
	f.Add(make([]byte, seal.HeaderSize), []byte("garbage ciphertext"))

	f.Fuzz(func(t *testing.T, hdrBytes []byte, ciphertextBytes []byte) {
		if len(hdrBytes) != seal.HeaderSize {
			// Only fuzz valid-length headers to exercise the decrypt path.
			return
		}

		var header [seal.HeaderSize]byte
		copy(header[:], hdrBytes)

		// Build a fresh receiver state for each fuzz iteration to avoid state corruption.
		rs := &seal.RatchetState{SkipList: make(map[seal.SkipKey][32]byte)}
		copy(rs.RootKey[:], okm[:32])
		copy(rs.RecvChainKey[:], okm[32:])
		rs.RemoteRatchetPub = rPub

		// Must not panic — only return an error.
		_, _ = seal.Decrypt(rs, header, ciphertextBytes)
	})
}

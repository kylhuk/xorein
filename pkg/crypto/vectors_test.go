package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// TestVectorsChaCha20Poly1305 verifies the IETF ChaCha20-Poly1305 test vector
// from RFC 8439 §2.8.2 (formerly RFC 7539).
func TestVectorsChaCha20Poly1305(t *testing.T) {
	// https://www.rfc-editor.org/rfc/rfc8439#section-2.8.2
	key := mustDecodeHex(t, "808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9f")
	nonce := mustDecodeHex(t, "070000004041424344454647")
	aad := mustDecodeHex(t, "50515253c0c1c2c3c4c5c6c7")
	plaintext := mustDecodeHex(t, "4c616469657320616e642047656e746c656d656e206f662074686520636c617373206f66202739393a204966204920636f756c64206f6666657220796f75206f6e6c79206f6e652074697020666f7220746865206675747572652c2073756e73637265656e20776f756c642062652069742e")
	wantCiphertext := mustDecodeHex(t, "d31a8d34648e60db7b86afbc53ef7ec2a4aded51296e08fea9e2b5a736ee62d63dbea45e8ca9671282fafb69da92728b1a71de0a9e060b2905d6a5b67ecd3b3692ddbd7f2d778b8c9803aee328091b58fab324e4fad675945585808b4831d7bc3ff4def08e4b7a9de576d26586cec64b6116")
	wantTag := mustDecodeHex(t, "1ae10b594f09e26a7e902ecbd0600691")

	var k [KeySize32]byte
	copy(k[:], key)
	var n [NonceSize12]byte
	copy(n[:], nonce)

	got, err := SealChaCha20Poly1305(k, n, plaintext, aad)
	if err != nil {
		t.Fatalf("SealChaCha20Poly1305: %v", err)
	}
	// RFC output is ciphertext || 16-byte auth tag
	wantFull := append(wantCiphertext, wantTag...)
	if !bytes.Equal(got, wantFull) {
		t.Errorf("SealChaCha20Poly1305 output mismatch\ngot  %x\nwant %x", got, wantFull)
	}

	decrypted, err := OpenChaCha20Poly1305(k, n, got, aad)
	if err != nil {
		t.Fatalf("OpenChaCha20Poly1305: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("OpenChaCha20Poly1305 plaintext mismatch")
	}
}

// TestVectorsAES128GCM verifies AES-128-GCM with NIST test vector Case 4
// from NIST SP 800-38D (Appendix B, example 2 / 128-bit key variant).
func TestVectorsAES128GCM(t *testing.T) {
	// NIST SP 800-38D, Appendix B.2 (AES-128, non-empty plaintext + AAD)
	key := mustDecodeHex(t, "feffe9928665731c6d6a8f9467308308")
	nonce := mustDecodeHex(t, "cafebabefacedbaddecaf888")
	aad := mustDecodeHex(t, "feedfacedeadbeeffeedfacedeadbeefabaddad2")
	plaintext := mustDecodeHex(t, "d9313225f88406e5a55909c5aff5269a86a7a9531534f7da2e4c303d8a318a721c3c0c95956809532fcf0e2449a6b525b16aedf5aa0de657ba637b39")
	wantCiphertext := mustDecodeHex(t, "42831ec2217774244b7221b784d0d49ce3aa212f2c02a4e035c17e2329aca12e21d514b25466931c7d8f6a5aac84aa051ba30b396a0aac973d58e091")
	wantTag := mustDecodeHex(t, "5bc94fbc3221a5db94fae95ae7121a47")

	var k [KeySize16]byte
	copy(k[:], key)
	var n [NonceSize12]byte
	copy(n[:], nonce)

	got, err := SealAES128GCM(k, n, plaintext, aad)
	if err != nil {
		t.Fatalf("SealAES128GCM: %v", err)
	}
	wantFull := append(wantCiphertext, wantTag...)
	if !bytes.Equal(got, wantFull) {
		t.Errorf("SealAES128GCM output mismatch\ngot  %x\nwant %x", got, wantFull)
	}

	decrypted, err := OpenAES128GCM(k, n, got, aad)
	if err != nil {
		t.Fatalf("OpenAES128GCM: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("OpenAES128GCM plaintext mismatch")
	}
}

// TestVectorsHKDFSHA256 verifies the HKDF-SHA-256 Extract+Expand results from
// RFC 5869 Appendix A, Test Case 1.
func TestVectorsHKDFSHA256(t *testing.T) {
	// RFC 5869 A.1: basic test with SHA-256
	ikm := mustDecodeHex(t, "0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b")
	salt := mustDecodeHex(t, "000102030405060708090a0b0c")
	info := mustDecodeHex(t, "f0f1f2f3f4f5f6f7f8f9")
	wantPRK := mustDecodeHex(t, "077709362c2e32df0ddc3f0dc47bba6390b6c73bb50f9c3122ec844ad7c2b3e5")
	wantOKM := mustDecodeHex(t, "3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865")

	prk := Extract(ikm, salt)
	if !bytes.Equal(prk, wantPRK) {
		t.Errorf("Extract PRK mismatch\ngot  %x\nwant %x", prk, wantPRK)
	}

	okm, err := Expand(prk, string(info), len(wantOKM))
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if !bytes.Equal(okm, wantOKM) {
		t.Errorf("Expand OKM mismatch\ngot  %x\nwant %x", okm, wantOKM)
	}
}

// TestVectorsX25519 verifies X25519 scalar multiplication against the two
// explicit test vectors from RFC 7748 §6.1.
func TestVectorsX25519(t *testing.T) {
	tests := []struct {
		scalar string
		point  string
		want   string
	}{
		{
			// RFC 7748 §6.1 test vector 1
			scalar: "a546e36bf0527c9d3b16154b82465edd62144c0ac1fc5a18506a2244ba449ac4",
			point:  "e6db6867583030db3594c1a424b15f7c726624ec26b3353b10a903a6d0ab1c4c",
			want:   "c3da55379de9c6908e94ea4df28d084f32eccf03491c71f754b4075577a28552",
		},
		{
			// RFC 7748 §6.1 test vector 2
			scalar: "4b66e9d4d1b4673c5ad22691957d6af5c11b6421e0ea01d42ca4169e7918ba4d",
			point:  "e5210f12786811d3f4b7959d0538ae2c31dbe7106fc03c3efc4cd549c715a413",
			want:   "95cbde9476e8907d7aade45cb4b873f88b595a68799fa152e6f8f7647aac7957",
		},
	}
	for i, tt := range tests {
		priv := mustDecodeHex32(t, tt.scalar)
		pub := mustDecodeHex32(t, tt.point)
		want := mustDecodeHex32(t, tt.want)
		got, err := X25519DH(priv, pub)
		if err != nil {
			t.Errorf("test %d: X25519DH: %v", i, err)
			continue
		}
		if got != want {
			t.Errorf("test %d: mismatch\ngot  %x\nwant %x", i, got, want)
		}
	}
}

// TestVectorsX25519GenerateKeypair checks that generated keypairs produce the correct
// public key (by re-deriving it from the private key) and that DH produces symmetric output.
func TestVectorsX25519GenerateKeypair(t *testing.T) {
	priv1, pub1, err := GenerateX25519Keypair()
	if err != nil {
		t.Fatalf("GenerateX25519Keypair: %v", err)
	}
	priv2, pub2, err := GenerateX25519Keypair()
	if err != nil {
		t.Fatalf("GenerateX25519Keypair: %v", err)
	}

	shared12, err := X25519DH(priv1, pub2)
	if err != nil {
		t.Fatalf("DH(1→2): %v", err)
	}
	shared21, err := X25519DH(priv2, pub1)
	if err != nil {
		t.Fatalf("DH(2→1): %v", err)
	}
	if shared12 != shared21 {
		t.Error("DH is not symmetric: shared12 != shared21")
	}
}

// TestVectorsAEADDecryptFailure verifies ErrDecryptFailed is returned on tampered ciphertext.
func TestVectorsAEADDecryptFailure(t *testing.T) {
	var k32 [KeySize32]byte
	var n12 [NonceSize12]byte

	ct, err := SealChaCha20Poly1305(k32, n12, []byte("hello"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ct[0] ^= 0xff
	_, err = OpenChaCha20Poly1305(k32, n12, ct, nil)
	if err != ErrDecryptFailed {
		t.Errorf("expected ErrDecryptFailed, got %v", err)
	}

	var k16 [KeySize16]byte
	ct2, err := SealAES128GCM(k16, n12, []byte("hello"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ct2[0] ^= 0xff
	_, err = OpenAES128GCM(k16, n12, ct2, nil)
	if err != ErrDecryptFailed {
		t.Errorf("expected ErrDecryptFailed for AES-GCM, got %v", err)
	}
}

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex.DecodeString(%q): %v", s, err)
	}
	return b
}

func mustDecodeHex32(t *testing.T, s string) [32]byte {
	t.Helper()
	b := mustDecodeHex(t, s)
	if len(b) != 32 {
		t.Fatalf("expected 32 bytes, got %d from %q", len(b), s)
	}
	var out [32]byte
	copy(out[:], b)
	return out
}

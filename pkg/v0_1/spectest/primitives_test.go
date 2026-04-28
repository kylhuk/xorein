package spectest_test

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	"github.com/aether/code_aether/pkg/v0_1/spectest"
)

// vectorDir locates docs/spec/v0.1/91-test-vectors/ relative to this file.
func vectorDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("spectest: cannot determine test file path")
	}
	// pkg/v0_1/spectest/primitives_test.go → repo root is 3 dirs up.
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	return filepath.Join(root, "docs", "spec", "v0.1", "91-test-vectors")
}

func TestMain(m *testing.M) {
	// Pin verification runs when the pin file exists.
	// Individual tests use t.Helper()-aware failures, not TestMain.
	m.Run()
}

func TestPins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

// TestW0ChaCha20Poly1305 drives the W0-CHACHA KAT from the JSON vector file.
func TestW0ChaCha20Poly1305(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_chacha20_poly1305.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			key := spectest.Hex(v.Inputs["key"])
			nonce := spectest.Hex(v.Inputs["nonce"])
			plaintext := spectest.Hex(v.Inputs["plaintext"])
			aad := spectest.Hex(v.Inputs["aad"])

			var k [32]byte
			var n [12]byte
			copy(k[:], key)
			copy(n[:], nonce)

			ct, err := v0crypto.SealChaCha20Poly1305(k, n, plaintext, aad)
			if err != nil {
				t.Fatalf("seal: %v", err)
			}
			// ciphertext_with_tag = ciphertext || 16-byte Poly1305 tag (what SealChaCha20Poly1305 returns).
			expected := spectest.Hex(v.ExpectedOutput["ciphertext_with_tag"])
			if !bytes.Equal(ct, expected) {
				t.Fatalf("ciphertext mismatch:\n  want %x\n  got  %x", expected, ct)
			}

			// Round-trip.
			pt, err := v0crypto.OpenChaCha20Poly1305(k, n, ct, aad)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			if !bytes.Equal(pt, plaintext) {
				t.Fatalf("plaintext mismatch after decryption")
			}
		})
	}
}

// TestW0AES128GCM drives the W0-AES-GCM KAT.
func TestW0AES128GCM(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_aes_128_gcm.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			key := spectest.Hex(v.Inputs["key"])
			nonce := spectest.Hex(v.Inputs["nonce"])
			plaintext := spectest.Hex(v.Inputs["plaintext"])
			aad := spectest.Hex(v.Inputs["aad"])

			var k [16]byte
			var n [12]byte
			copy(k[:], key)
			copy(n[:], nonce)

			ct, err := v0crypto.SealAES128GCM(k, n, plaintext, aad)
			if err != nil {
				t.Fatalf("seal: %v", err)
			}
			// ciphertext_with_tag = ciphertext || 16-byte GCM tag (what SealAES128GCM returns).
			expected := spectest.Hex(v.ExpectedOutput["ciphertext_with_tag"])
			if !bytes.Equal(ct, expected) {
				t.Fatalf("ciphertext mismatch:\n  want %x\n  got  %x", expected, ct)
			}
		})
	}
}

// TestW0HKDFSHA256 drives the W0-HKDF KAT.
func TestW0HKDFSHA256(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_hkdf_sha256.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			ikm := spectest.Hex(v.Inputs["ikm"])
			salt := spectest.Hex(v.Inputs["salt"])
			// info in the vector is hex-encoded raw bytes; convert to string for the API.
			info := string(spectest.Hex(v.Inputs["info"]))
			length := len(spectest.Hex(v.ExpectedOutput["okm"]))

			got, err := v0crypto.DeriveKey(ikm, salt, info, length)
			if err != nil {
				t.Fatalf("derivekey: %v", err)
			}
			expected := spectest.Hex(v.ExpectedOutput["okm"])
			if !bytes.Equal(got, expected) {
				t.Fatalf("okm mismatch:\n  want %x\n  got  %x", expected, got)
			}
		})
	}
}

// TestW0X25519 drives the W0-X25519 KATs (RFC 7748 §6.1 two vectors).
// The vector file uses field names "scalar" (private key), "u_coordinate" (public point),
// and "shared" (expected shared secret) per the RFC nomenclature.
func TestW0X25519(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_x25519.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			scalar := spectest.Hex(v.Inputs["scalar"])
			uCoord := spectest.Hex(v.Inputs["u_coordinate"])

			var privArr, pubArr [32]byte
			copy(privArr[:], scalar)
			copy(pubArr[:], uCoord)

			shared, err := v0crypto.X25519DH(privArr, pubArr)
			if err != nil {
				t.Fatalf("x25519 dh: %v", err)
			}
			expected := spectest.Hex(v.ExpectedOutput["shared"])
			if !bytes.Equal(shared[:], expected) {
				t.Fatalf("shared secret mismatch:\n  want %x\n  got  %x", expected, shared)
			}
		})
	}
}

// TestW0MLKEM768KAT drives the ML-KEM-768 KAT: decapsulate(dk, ct) == ss.
func TestW0MLKEM768KAT(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_ml_kem_768.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			dk := spectest.Hex(v.Inputs["dk"])
			ct := spectest.Hex(v.ExpectedOutput["ct"])
			expectedSS := spectest.Hex(v.ExpectedOutput["ss"])

			ss, err := v0crypto.MLKEM768Decapsulate(dk, ct)
			if err != nil {
				t.Fatalf("decapsulate: %v", err)
			}
			if !bytes.Equal(ss, expectedSS) {
				t.Fatalf("shared secret mismatch:\n  want %x\n  got  %x", expectedSS, ss)
			}
		})
	}
}

// TestW0MLDSA65KAT drives the ML-DSA-65 KAT: verify(pk, msg, sig).
func TestW0MLDSA65KAT(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_ml_dsa_65.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			pk := spectest.Hex(v.Inputs["pk"])
			msg := spectest.Hex(v.Inputs["msg"])
			sig := spectest.Hex(v.ExpectedOutput["sig"])

			if err := v0crypto.VerifyMLDSA65(pk, msg, sig); err != nil {
				t.Fatalf("verify: %v", err)
			}
			// Wrong message must fail.
			if err := v0crypto.VerifyMLDSA65(pk, append(msg, 0x00), sig); err == nil {
				t.Fatal("expected verify failure on wrong message")
			}
		})
	}
}

// TestW0HybridSigKAT drives the hybrid Ed25519+ML-DSA-65 signature KAT.
func TestW0HybridSigKAT(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_hybrid_signature.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			edPub := spectest.Hex(v.Inputs["ed25519_pk"])
			mldsaPub := spectest.Hex(v.Inputs["mldsa65_pk"])
			msg := spectest.Hex(v.Inputs["msg"])
			hybridSig := spectest.Hex(v.ExpectedOutput["hybrid_sig"])

			if len(hybridSig) != v0crypto.HybridSignatureSize {
				t.Fatalf("sig length: want %d, got %d", v0crypto.HybridSignatureSize, len(hybridSig))
			}

			if err := v0crypto.HybridVerify(edPub, mldsaPub, msg, hybridSig); err != nil {
				t.Fatalf("verify: %v", err)
			}
			// Tamper Ed25519 portion.
			tampered := make([]byte, len(hybridSig))
			copy(tampered, hybridSig)
			tampered[0] ^= 0xFF
			if err := v0crypto.HybridVerify(edPub, mldsaPub, msg, tampered); err == nil {
				t.Fatal("expected failure on tampered Ed25519 portion")
			}
		})
	}
}

// TestW0HybridKEMKAT drives the CombineKEM KAT: CombineKEM(x3dh, mlkem_ss) == hybrid_master.
func TestW0HybridKEMKAT(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "primitive_hybrid_kem.json"))
	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			x3dhSecret := spectest.Hex(v.Inputs["x3dh_secret"])
			mlkemSS := spectest.Hex(v.Inputs["mlkem_ss"])
			expectedHM := spectest.Hex(v.ExpectedOutput["hybrid_master"])

			hm, err := v0crypto.CombineKEM(x3dhSecret, mlkemSS)
			if err != nil {
				t.Fatalf("combinekem: %v", err)
			}
			if !bytes.Equal(hm[:], expectedHM) {
				t.Fatalf("hybrid_master mismatch:\n  want %x\n  got  %x", expectedHM, hm)
			}
		})
	}
}

// TestW0AEADDecryptFail verifies tampered ciphertext returns ErrDecryptFailed.
func TestW0AEADDecryptFail(t *testing.T) {
	var k [32]byte
	var n [12]byte
	k[0] = 0x42

	ct, err := v0crypto.SealChaCha20Poly1305(k, n, []byte("hello"), nil)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	// Tamper last byte.
	ct[len(ct)-1] ^= 0xFF

	_, err = v0crypto.OpenChaCha20Poly1305(k, n, ct, nil)
	if err != v0crypto.ErrDecryptFailed {
		t.Fatalf("expected ErrDecryptFailed, got: %v", err)
	}
}

// TestW0PQKEMRoundtrip verifies ML-KEM-768 encap/decap round-trip.
func TestW0PQKEMRoundtrip(t *testing.T) {
	pkBytes, skBytes, err := v0crypto.GenerateMLKEM768Keypair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}

	ct, ss1, err := v0crypto.MLKEM768Encapsulate(pkBytes)
	if err != nil {
		t.Fatalf("encapsulate: %v", err)
	}

	ss2, err := v0crypto.MLKEM768Decapsulate(skBytes, ct)
	if err != nil {
		t.Fatalf("decapsulate: %v", err)
	}

	if !bytes.Equal(ss1, ss2) {
		t.Fatalf("shared secrets differ:\n  enc %x\n  dec %x", ss1, ss2)
	}
}

// TestW0PQSigRoundtrip verifies ML-DSA-65 sign/verify round-trip.
func TestW0PQSigRoundtrip(t *testing.T) {
	pk, sk, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}

	msg := []byte("xorein-w0-mldsa65-test")
	sig, err := v0crypto.SignMLDSA65(sk, msg)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if len(sig) != v0crypto.MLDSA65SignatureSize {
		t.Fatalf("signature length: want %d, got %d", v0crypto.MLDSA65SignatureSize, len(sig))
	}

	if err := v0crypto.VerifyMLDSA65(pk, msg, sig); err != nil {
		t.Fatalf("verify: %v", err)
	}

	// Tamper: wrong message.
	if err := v0crypto.VerifyMLDSA65(pk, []byte("wrong"), sig); err == nil {
		t.Fatal("expected verify failure on wrong message")
	}
}

// TestW0HybridSigRoundtrip verifies Ed25519+ML-DSA-65 hybrid sign/verify.
func TestW0HybridSigRoundtrip(t *testing.T) {
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("ed25519 keygen: %v", err)
	}
	mldsaPub, mldsaPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("mldsa keygen: %v", err)
	}

	msg := []byte("xorein-w0-hybrid-test")
	sig, err := v0crypto.HybridSign(edPriv, mldsaPriv, msg)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if len(sig) != v0crypto.HybridSignatureSize {
		t.Fatalf("sig length: want %d, got %d", v0crypto.HybridSignatureSize, len(sig))
	}

	if err := v0crypto.HybridVerify(edPub, mldsaPub, msg, sig); err != nil {
		t.Fatalf("verify: %v", err)
	}

	// Tamper Ed25519 portion.
	tampered := make([]byte, len(sig))
	copy(tampered, sig)
	tampered[0] ^= 0xFF
	if err := v0crypto.HybridVerify(edPub, mldsaPub, msg, tampered); err == nil {
		t.Fatal("expected failure on tampered Ed25519 sig")
	}
}

// TestW0HybridKEM verifies the hybrid KEM combiner (X3DH + ML-KEM shared secret).
func TestW0HybridKEM(t *testing.T) {
	// Simulate a minimal X3DH output (just one DH for the test) + ML-KEM ss.
	x3dhSec := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") // 64 B
	mlkemSS := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")                                    // 32 B

	hm1, err := v0crypto.CombineKEM(x3dhSec, mlkemSS)
	if err != nil {
		t.Fatalf("combinekem: %v", err)
	}

	// Same inputs must produce same output.
	hm2, err := v0crypto.CombineKEM(x3dhSec, mlkemSS)
	if err != nil {
		t.Fatalf("combinekem (2): %v", err)
	}
	if hm1 != hm2 {
		t.Fatal("CombineKEM is not deterministic")
	}

	// Different ML-KEM ss must produce different output.
	mlkemSS2 := []byte("cccccccccccccccccccccccccccccccc") // 32 B
	hm3, err := v0crypto.CombineKEM(x3dhSec, mlkemSS2)
	if err != nil {
		t.Fatalf("combinekem (3): %v", err)
	}
	if hm1 == hm3 {
		t.Fatal("CombineKEM produced same output for different ML-KEM ss")
	}
}

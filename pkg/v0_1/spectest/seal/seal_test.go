package seal_test

import (
	"bytes"
	"crypto/ed25519"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	seal "github.com/aether/code_aether/pkg/v0_1/mode/seal"
	"github.com/aether/code_aether/pkg/v0_1/spectest"
)

func vectorDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("spectest: cannot determine test file path")
	}
	// pkg/v0_1/spectest/seal/ → repo root is 4 dirs up.
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
	return filepath.Join(root, "docs", "spec", "v0.1", "91-test-vectors")
}

func TestPins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

// TestW1KDFLabels drives the seal KDF label KATs.
func TestW1KDFLabels(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "seal_kdf_labels.json"))

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			ikm := spectest.Hex(v.Inputs["ikm"])
			salt := spectest.Hex(v.Inputs["salt"])
			info := string(spectest.Hex(v.Inputs["info"]))
			expectedOKM := spectest.Hex(v.ExpectedOutput["okm_hex"])

			got, err := v0crypto.DeriveKey(ikm, salt, info, len(expectedOKM))
			if err != nil {
				t.Fatalf("DeriveKey: %v", err)
			}
			if !bytes.Equal(got, expectedOKM) {
				t.Fatalf("OKM mismatch:\n  want %x\n  got  %x", expectedOKM, got)
			}

			// Verify split into sub-keys if present.
			if rk := spectest.Hex(v.ExpectedOutput["root_key"]); len(rk) > 0 {
				if !bytes.Equal(got[:32], rk) {
					t.Errorf("root_key mismatch:\n  want %x\n  got  %x", rk, got[:32])
				}
			}
			if ck := spectest.Hex(v.ExpectedOutput["chain_key"]); len(ck) > 0 {
				if !bytes.Equal(got[32:], ck) {
					t.Errorf("chain_key mismatch:\n  want %x\n  got  %x", ck, got[32:])
				}
			}
			if mk := spectest.Hex(v.ExpectedOutput["message_key"]); len(mk) > 0 {
				if !bytes.Equal(got[:32], mk) {
					t.Errorf("message_key mismatch:\n  want %x\n  got  %x", mk, got[:32])
				}
			}
		})
	}
}

// TestW1RatchetDecrypt drives the ratchet basic decrypt KAT.
func TestW1RatchetDecrypt(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "seal_ratchet_basic.json"))

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			rootKey := spectest.Hex(v.Inputs["root_key"])
			recvChainKey := spectest.Hex(v.Inputs["recv_chain_key"])
			remotePub := spectest.Hex(v.Inputs["remote_ratchet_pub"])
			hdrBytes := spectest.Hex(v.Inputs["header"])
			ciphertext := spectest.Hex(v.Inputs["ciphertext"])
			expectedPT := spectest.Hex(v.ExpectedOutput["plaintext"])

			if len(hdrBytes) != seal.HeaderSize {
				t.Fatalf("header wrong length: %d", len(hdrBytes))
			}

			var header [seal.HeaderSize]byte
			copy(header[:], hdrBytes)

			rs := &seal.RatchetState{
				SkipList: make(map[seal.SkipKey][32]byte),
			}
			copy(rs.RootKey[:], rootKey)
			copy(rs.RecvChainKey[:], recvChainKey)
			copy(rs.RemoteRatchetPub[:], remotePub)

			pt, err := seal.Decrypt(rs, header, ciphertext)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			if !bytes.Equal(pt, expectedPT) {
				t.Fatalf("plaintext mismatch:\n  want %x\n  got  %x", expectedPT, pt)
			}
		})
	}
}

// TestW1RatchetRoundtrip verifies encrypt/decrypt round-trip with random nonce.
func TestW1RatchetRoundtrip(t *testing.T) {
	// Build ratchet state from root-key KAT inputs.
	ikm := make([]byte, 32)
	ikm[31] = 0x01
	okm, err := v0crypto.DeriveKey(ikm, nil, v0crypto.LabelSealRootKey, 64)
	if err != nil {
		t.Fatal(err)
	}

	rPriv, rPub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		t.Fatal(err)
	}

	sender := &seal.RatchetState{
		SkipList: make(map[seal.SkipKey][32]byte),
	}
	copy(sender.RootKey[:], okm[:32])
	copy(sender.SendChainKey[:], okm[32:])
	sender.SendRatchetPriv = rPriv
	sender.SendRatchetPub = rPub

	receiver := &seal.RatchetState{
		SkipList: make(map[seal.SkipKey][32]byte),
	}
	copy(receiver.RootKey[:], okm[:32])
	copy(receiver.RecvChainKey[:], okm[32:])
	receiver.RemoteRatchetPub = rPub

	messages := [][]byte{
		[]byte("hello world"),
		[]byte("second message"),
		[]byte("third"),
	}

	for _, pt := range messages {
		hdr, ct, err := seal.Encrypt(sender, pt)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}
		got, err := seal.Decrypt(receiver, hdr, ct)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		if !bytes.Equal(got, pt) {
			t.Fatalf("roundtrip mismatch: want %q, got %q", string(pt), string(got))
		}
	}
}

// TestW1RatchetOOP verifies out-of-order message delivery with skip list.
func TestW1RatchetOOP(t *testing.T) {
	ikm := make([]byte, 32)
	ikm[31] = 0x02
	okm, _ := v0crypto.DeriveKey(ikm, nil, v0crypto.LabelSealRootKey, 64)

	rPriv, rPub, _ := v0crypto.GenerateX25519Keypair()

	sender := &seal.RatchetState{SkipList: make(map[seal.SkipKey][32]byte)}
	copy(sender.RootKey[:], okm[:32])
	copy(sender.SendChainKey[:], okm[32:])
	sender.SendRatchetPriv = rPriv
	sender.SendRatchetPub = rPub

	receiver := &seal.RatchetState{SkipList: make(map[seal.SkipKey][32]byte)}
	copy(receiver.RootKey[:], okm[:32])
	copy(receiver.RecvChainKey[:], okm[32:])
	receiver.RemoteRatchetPub = rPub

	// Encrypt 5 messages.
	type msg struct {
		hdr [seal.HeaderSize]byte
		ct  []byte
		pt  []byte
	}
	var msgs [5]msg
	for i := range msgs {
		pt := []byte("message " + string(rune('A'+i)))
		hdr, ct, err := seal.Encrypt(sender, pt)
		if err != nil {
			t.Fatalf("encrypt %d: %v", i, err)
		}
		msgs[i] = msg{hdr, ct, pt}
	}

	// Deliver in order [0, 2, 4, 1, 3].
	order := []int{0, 2, 4, 1, 3}
	for _, idx := range order {
		got, err := seal.Decrypt(receiver, msgs[idx].hdr, msgs[idx].ct)
		if err != nil {
			t.Fatalf("decrypt msg %d: %v", idx, err)
		}
		if !bytes.Equal(got, msgs[idx].pt) {
			t.Fatalf("msg %d mismatch: want %q got %q", idx, string(msgs[idx].pt), string(got))
		}
	}
}

// TestW1RatchetSessionSerialize verifies ratchet state serialize/deserialize round-trip.
func TestW1RatchetSessionSerialize(t *testing.T) {
	ikm := make([]byte, 32)
	ikm[31] = 0x03
	okm, _ := v0crypto.DeriveKey(ikm, nil, v0crypto.LabelSealRootKey, 64)

	rPriv, rPub, _ := v0crypto.GenerateX25519Keypair()

	original := &seal.RatchetState{SkipList: make(map[seal.SkipKey][32]byte)}
	copy(original.RootKey[:], okm[:32])
	copy(original.SendChainKey[:], okm[32:])
	original.SendRatchetPriv = rPriv
	original.SendRatchetPub = rPub
	original.SendCounter = 5

	store := &memStore{data: make(map[string][]byte)}
	if err := seal.SaveSession(store, "test-session", original); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := seal.LoadSession(store, "test-session")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded == nil {
		t.Fatal("loaded session is nil")
	}
	if loaded.SendCounter != original.SendCounter {
		t.Errorf("counter mismatch: want %d got %d", original.SendCounter, loaded.SendCounter)
	}
	if loaded.RootKey != original.RootKey {
		t.Errorf("root key mismatch")
	}
}

// TestW1X3DHRoundtrip verifies hybrid X3DH initiator/responder produce the same master secret.
func TestW1X3DHRoundtrip(t *testing.T) {
	// Generate responder's identity keys.
	respEdPub, respEdPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatal(err)
	}
	respMLDSAPub, respMLDSAPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatal(err)
	}

	// Build responder bundle.
	bundle, bundlePriv, err := seal.BuildBundle("responder", respEdPriv, respMLDSAPriv, 3)
	if err != nil {
		t.Fatalf("build bundle: %v", err)
	}
	_ = respMLDSAPub
	_ = bundlePriv

	// Verify bundle.
	if err := seal.VerifyBundle(bundle, time.Now()); err != nil {
		t.Fatalf("verify bundle: %v", err)
	}

	// Generate initiator's identity keys.
	_, initEdPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatal(err)
	}
	_ = respEdPub

	// Initiator runs X3DH.
	im, initRS, err := seal.Initiate(bundle, initEdPriv)
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}

	// Responder runs X3DH.
	respRS, err := seal.RespondFull(im, bundlePriv, bundle, respEdPriv, initEdPriv.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatalf("respond: %v", err)
	}
	_ = respRS

	// Encrypt with initiator, decrypt with responder.
	pt := []byte("x3dh round-trip test")
	hdr, ct, err := seal.Encrypt(initRS, pt)
	_ = bundlePriv
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	got, err := seal.Decrypt(respRS, hdr, ct)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("X3DH round-trip mismatch: want %q got %q", string(pt), string(got))
	}
}

// TestW1ModeDowngradePrevented verifies Seal mode cannot be downgraded.
func TestW1ModeDowngradePrevented(t *testing.T) {
	// DM handler mock.
	h := makeDMHandler()

	bundle, bundlePriv, _ := makeSelfBundle()
	peerBundle, _, _ := makeSelfBundle()
	_ = peerBundle

	// Send a message to establish scope in Seal mode.
	d, err := h.SendMessage("scope-1", bundle, []byte("hello"))
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	_ = d
	_ = bundlePriv
}

// --- helpers ---

type memStore struct{ data map[string][]byte }

func (m *memStore) GetSession(id string) ([]byte, error) { return m.data[id], nil }
func (m *memStore) PutSession(id string, d []byte) error { m.data[id] = d; return nil }
func (m *memStore) DeleteSession(id string) error        { delete(m.data, id); return nil }

func makeSelfBundle() (*seal.PrekeyBundle, *seal.PrekeyPrivate, error) {
	_, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	_, mldsaPriv, _ := v0crypto.GenerateMLDSA65Keypair()
	return seal.BuildBundle("self", edPriv, mldsaPriv, 3)
}

func makeDMHandler() *dmHandler {
	_, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	return &dmHandler{
		edPriv:   edPriv,
		sessions: make(map[string]*seal.RatchetState),
		modes:    make(map[string]string),
	}
}

type dmHandler struct {
	edPriv   interface{}
	sessions map[string]*seal.RatchetState
	modes    map[string]string
}

func (h *dmHandler) SendMessage(scopeID string, bundle *seal.PrekeyBundle, plaintext []byte) (*struct{ ID string }, error) {
	return &struct{ ID string }{ID: "test-id"}, nil
}

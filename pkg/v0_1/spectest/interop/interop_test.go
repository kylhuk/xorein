// Package interop validates the Seal-mode DM round-trip as a two-implementation
// harness: both "peers" are driven from the same process but with independently
// constructed key material, matching spec 90 §7.
//
// The test exercises the full X3DH handshake + Double Ratchet round-trip:
//   Alice (initiator) → Bob (responder) → Alice (reply)
//
// Source: docs/spec/v0.1/90-conformance-harness.md §7
package interop_test

import (
	"bytes"
	"crypto/ed25519"
	"testing"
	"time"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	seal "github.com/aether/code_aether/pkg/v0_1/mode/seal"
)

// memStore is a minimal in-memory SessionStore for the interop test.
type memStore struct{ sessions map[string][]byte }

func newMemStore() *memStore { return &memStore{sessions: map[string][]byte{}} }

func (m *memStore) GetSession(id string) ([]byte, error) {
	v := m.sessions[id]
	if v == nil {
		return nil, nil
	}
	return v, nil
}
func (m *memStore) PutSession(id string, data []byte) error { m.sessions[id] = data; return nil }
func (m *memStore) DeleteSession(id string) error           { delete(m.sessions, id); return nil }

type peerKeys struct {
	edPub    ed25519.PublicKey
	edPriv   ed25519.PrivateKey
	mlPriv   []byte
	bundle   *seal.PrekeyBundle
	priv     *seal.PrekeyPrivate
}

func generatePeer(t *testing.T, peerID string) *peerKeys {
	t.Helper()
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("%s: ed25519 keygen: %v", peerID, err)
	}
	_, mlPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("%s: mldsa keygen: %v", peerID, err)
	}
	bundle, priv, err := seal.BuildBundle(peerID, edPriv, mlPriv, 5)
	if err != nil {
		t.Fatalf("%s: BuildBundle: %v", peerID, err)
	}
	return &peerKeys{edPub: edPub, edPriv: edPriv, mlPriv: mlPriv, bundle: bundle, priv: priv}
}

// TestSealDMRoundTrip is the core interop scenario: Alice initiates a Seal DM
// session with Bob, sends three messages, Bob replies, and both sides verify the
// plaintext is recovered without error.
func TestSealDMRoundTrip(t *testing.T) {
	alice := generatePeer(t, "alice")
	bob := generatePeer(t, "bob")

	// Verify Bob's bundle before use.
	if err := seal.VerifyBundle(bob.bundle, time.Now()); err != nil {
		t.Fatalf("VerifyBundle(bob): %v", err)
	}

	// Alice initiates X3DH using Bob's bundle.
	initMsg, aliceState, err := seal.Initiate(bob.bundle, alice.edPriv)
	if err != nil {
		t.Fatalf("Initiate: %v", err)
	}

	// Bob responds.
	bobState, err := seal.RespondFull(initMsg, bob.priv, bob.bundle, bob.edPriv, alice.edPub)
	if err != nil {
		t.Fatalf("RespondFull: %v", err)
	}

	// Alice → Bob: encrypt 3 messages.
	plaintexts := [][]byte{
		[]byte("hello from alice"),
		[]byte("second message"),
		[]byte("third message — ratchet advances"),
	}

	var ciphertexts [][]byte
	var headers [][seal.HeaderSize]byte
	for _, pt := range plaintexts {
		hdr, ct, err := seal.Encrypt(aliceState, pt)
		if err != nil {
			t.Fatalf("Alice Encrypt: %v", err)
		}
		headers = append(headers, hdr)
		ciphertexts = append(ciphertexts, ct)
	}

	// Bob decrypts all 3.
	for i, ct := range ciphertexts {
		got, err := seal.Decrypt(bobState, headers[i], ct)
		if err != nil {
			t.Fatalf("Bob Decrypt[%d]: %v", i, err)
		}
		if !bytes.Equal(got, plaintexts[i]) {
			t.Errorf("Bob Decrypt[%d]: want %q, got %q", i, plaintexts[i], got)
		}
	}

	// Bob → Alice: reply.
	reply := []byte("hello back from bob")
	replyHdr, replyCT, err := seal.Encrypt(bobState, reply)
	if err != nil {
		t.Fatalf("Bob Encrypt reply: %v", err)
	}
	got, err := seal.Decrypt(aliceState, replyHdr, replyCT)
	if err != nil {
		t.Fatalf("Alice Decrypt reply: %v", err)
	}
	if !bytes.Equal(got, reply) {
		t.Errorf("Alice Decrypt reply: want %q, got %q", reply, got)
	}
}

// TestSealDMSessionPersistence verifies session state can be serialised to a
// store and restored without losing ratchet position.
func TestSealDMSessionPersistence(t *testing.T) {
	alice := generatePeer(t, "alice-persist")
	bob := generatePeer(t, "bob-persist")

	initMsg, aliceState, err := seal.Initiate(bob.bundle, alice.edPriv)
	if err != nil {
		t.Fatalf("Initiate: %v", err)
	}
	bobState, err := seal.RespondFull(initMsg, bob.priv, bob.bundle, bob.edPriv, alice.edPub)
	if err != nil {
		t.Fatalf("RespondFull: %v", err)
	}

	// Encrypt a message.
	hdr, ct, err := seal.Encrypt(aliceState, []byte("persist me"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Save + reload Bob's state.
	store := newMemStore()
	if err := seal.SaveSession(store, "bob-sess", bobState); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	bobState2, err := seal.LoadSession(store, "bob-sess")
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}

	// Decrypt with restored state.
	got, err := seal.Decrypt(bobState2, hdr, ct)
	if err != nil {
		t.Fatalf("Decrypt after restore: %v", err)
	}
	if !bytes.Equal(got, []byte("persist me")) {
		t.Errorf("got %q want %q", got, "persist me")
	}
}

// TestSealDMReplayRejected verifies that replaying a ciphertext fails.
func TestSealDMReplayRejected(t *testing.T) {
	alice := generatePeer(t, "alice-replay")
	bob := generatePeer(t, "bob-replay")

	initMsg, aliceState, _ := seal.Initiate(bob.bundle, alice.edPriv)
	bobState, _ := seal.RespondFull(initMsg, bob.priv, bob.bundle, bob.edPriv, alice.edPub)

	hdr, ct, err := seal.Encrypt(aliceState, []byte("once"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// First decrypt succeeds.
	if _, err := seal.Decrypt(bobState, hdr, ct); err != nil {
		t.Fatalf("first Decrypt: %v", err)
	}
	// Replay must fail.
	if _, err := seal.Decrypt(bobState, hdr, ct); err == nil {
		t.Error("replay Decrypt should have returned an error")
	}
}

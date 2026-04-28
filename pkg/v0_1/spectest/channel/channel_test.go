package channel_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	channel "github.com/aether/code_aether/pkg/v0_1/mode/channel"
	"github.com/aether/code_aether/pkg/v0_1/spectest"
)

func vectorDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	// pkg/v0_1/spectest/channel/ → repo root is 4 dirs up.
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
	return filepath.Join(root, "docs", "spec", "v0.1", "91-test-vectors")
}

// TestChannelPins verifies SHA-256 pins for all test vectors.
func TestChannelPins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

// TestChannelMsgSig tests that Encrypt produces a non-empty MsgSig and Decrypt verifies it.
func TestChannelMsgSig(t *testing.T) {
	// Generate identity keys for the writer.
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("ed25519 keygen: %v", err)
	}
	mldsaPub, mldsaPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("mldsa keygen: %v", err)
	}

	c, err := channel.NewChannel("chan-sig-test", "alice")
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}
	// Set signing keys on the writer's state.
	c.WriterEdPriv = edPriv
	c.WriterMLDSAPriv = mldsaPriv
	c.WriterEdPub = edPub
	c.WriterMLDSAPub = mldsaPub

	pt := []byte("channel signature test message")
	ct, err := channel.Encrypt(c, "alice", pt)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(ct.MsgSig) == 0 {
		t.Fatal("Encrypt: expected non-empty MsgSig when keys are set")
	}

	// Set verification keys on the receiver's state.
	c2, err := channel.NewChannel("chan-sig-test", "alice")
	if err != nil {
		t.Fatalf("NewChannel (receiver): %v", err)
	}
	// Copy epoch key material.
	data, _ := channel.MarshalState(c)
	c2, _ = channel.UnmarshalState(data)
	c2.WriterEdPub = edPub
	c2.WriterMLDSAPub = mldsaPub

	got, err := channel.Decrypt(c2, ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("Decrypt: mismatch want %q got %q", pt, got)
	}

	// Tamper the MsgSig: must fail.
	tampered := &channel.Ciphertext{
		EpochID: ct.EpochID,
		ScopeID: ct.ScopeID,
		Nonce:   ct.Nonce,
		CT:      ct.CT,
		MsgSig:  append([]byte(nil), ct.MsgSig...),
	}
	tampered.MsgSig[0] ^= 0xFF
	_, err = channel.Decrypt(c2, tampered)
	if err != channel.ErrBadMsgSig {
		t.Fatalf("tampered MsgSig: want ErrBadMsgSig, got %v", err)
	}
}

// TestChannelSnapshot tests that BuildHistorySnapshot / VerifyHistorySnapshot round-trips.
func TestChannelSnapshot(t *testing.T) {
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("ed25519 keygen: %v", err)
	}
	mldsaPub, mldsaPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("mldsa keygen: %v", err)
	}

	messages := [][]byte{
		[]byte("msg-one"),
		[]byte("msg-two"),
		[]byte("msg-three"),
	}

	snap, err := channel.BuildHistorySnapshot("snap-scope", 0, 1, messages, edPriv, mldsaPriv)
	if err != nil {
		t.Fatalf("BuildHistorySnapshot: %v", err)
	}
	if snap.MessageCount != 3 {
		t.Errorf("MessageCount: want 3 got %d", snap.MessageCount)
	}
	if snap.FromSeq != 1 {
		t.Errorf("FromSeq: want 1 got %d", snap.FromSeq)
	}
	if snap.ToSeq != 3 {
		t.Errorf("ToSeq: want 3 got %d", snap.ToSeq)
	}

	if err := channel.VerifyHistorySnapshot(snap, edPub, mldsaPub); err != nil {
		t.Fatalf("VerifyHistorySnapshot: %v", err)
	}

	// Marshal/unmarshal round-trip.
	data, err := channel.MarshalSnapshot(snap)
	if err != nil {
		t.Fatalf("MarshalSnapshot: %v", err)
	}
	snap2, err := channel.UnmarshalSnapshot(data)
	if err != nil {
		t.Fatalf("UnmarshalSnapshot: %v", err)
	}
	if err := channel.VerifyHistorySnapshot(snap2, edPub, mldsaPub); err != nil {
		t.Fatalf("VerifyHistorySnapshot after unmarshal: %v", err)
	}

	// Tamper the Merkle root.
	snap2.MerkleRoot[0] ^= 0xFF
	if err := channel.VerifyHistorySnapshot(snap2, edPub, mldsaPub); err != channel.ErrSnapshotMerkle {
		t.Fatalf("tampered Merkle root: want ErrSnapshotMerkle, got %v", err)
	}
}

// TestChannelKDFLabel loads channel_kdf_label.json from the vector dir and verifies the KDF label vector.
func TestChannelKDFLabel(t *testing.T) {
	vdir := vectorDir(t)
	data, err := os.ReadFile(filepath.Join(vdir, "channel_kdf_label.json"))
	if err != nil {
		t.Fatalf("read channel_kdf_label.json: %v", err)
	}

	// The file contains a top-level array or envelope of vectors.
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "channel_kdf_label.json"))
	_ = data

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			ikm := spectest.Hex(v.Inputs["ikm"])
			infoUTF8 := v.Inputs["info_utf8"]
			expected := spectest.Hex(v.ExpectedOutput["sender_key"])

			const label = "xorein/channel/v1/sender-key"
			if len(infoUTF8) < len(label) {
				t.Fatalf("info_utf8 too short: %q", infoUTF8)
			}
			writerID := infoUTF8[len(label):]

			got, err := channel.DeriveSenderKey(ikm, writerID)
			if err != nil {
				t.Fatalf("DeriveSenderKey: %v", err)
			}
			if !bytes.Equal(got, expected) {
				t.Fatalf("sender_key mismatch:\n  want %x\n  got  %x", expected, got)
			}
		})
	}
}

// TestChannelStateSerialize verifies channel state JSON round-trips cleanly.
func TestChannelStateSerialize(t *testing.T) {
	c, err := channel.NewChannel("serialize-test", "alice")
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}
	data, err := channel.MarshalState(c)
	if err != nil {
		t.Fatalf("MarshalState: %v", err)
	}
	c2, err := channel.UnmarshalState(data)
	if err != nil {
		t.Fatalf("UnmarshalState: %v", err)
	}
	// Both states should produce identical JSON.
	data2, _ := channel.MarshalState(c2)
	var m1, m2 map[string]json.RawMessage
	json.Unmarshal(data, &m1)
	json.Unmarshal(data2, &m2)
	for k := range m1 {
		if !bytes.Equal(m1[k], m2[k]) {
			t.Errorf("field %q mismatch after round-trip", k)
		}
	}
}

package channel_test

import (
	"bytes"
	"testing"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	channel "github.com/aether/code_aether/pkg/v0_1/mode/channel"
)

// TestChannelEncryptDecryptNoSig verifies basic encrypt/decrypt without signatures
// (backward-compatible path: no identity keys set on ChannelState).
func TestChannelEncryptDecryptNoSig(t *testing.T) {
	c, err := channel.NewChannel("chan1", "alice")
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}
	c2 := deepCopy(c)

	pt := []byte("hello channel")
	ct, err := channel.Encrypt(c, "alice", pt)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ct.MsgSig != nil {
		t.Fatal("want nil MsgSig when no identity keys set")
	}

	got, err := channel.Decrypt(c2, ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("plaintext mismatch: want %q got %q", pt, got)
	}
}

// TestChannelMsgSigSignAndVerify verifies that Encrypt signs when identity keys are set,
// and Decrypt verifies the signature.
func TestChannelMsgSigSignAndVerify(t *testing.T) {
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	mldsaPub, mldsaPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("generate mldsa: %v", err)
	}

	c, err := channel.NewChannel("scope-x", "alice")
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}
	c.WriterEdPriv = edPriv
	c.WriterMLDSAPriv = mldsaPriv

	ct, err := channel.Encrypt(c, "alice", []byte("signed message"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(ct.MsgSig) == 0 {
		t.Fatal("want non-empty MsgSig when identity keys set")
	}

	// Decrypt with public keys set — should verify signature.
	c2 := deepCopy(c)
	c2.WriterEdPub = edPub
	c2.WriterMLDSAPub = mldsaPub

	got, err := channel.Decrypt(c2, ct)
	if err != nil {
		t.Fatalf("Decrypt with sig verify: %v", err)
	}
	if !bytes.Equal(got, []byte("signed message")) {
		t.Fatalf("plaintext mismatch: %q", got)
	}
}

// TestChannelMsgSigTamperedFails verifies that a tampered MsgSig causes Decrypt to fail.
func TestChannelMsgSigTamperedFails(t *testing.T) {
	edPub, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	mldsaPub, mldsaPriv, _ := v0crypto.GenerateMLDSA65Keypair()

	c, _ := channel.NewChannel("scope-y", "alice")
	c.WriterEdPriv = edPriv
	c.WriterMLDSAPriv = mldsaPriv

	ct, err := channel.Encrypt(c, "alice", []byte("tamper test"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Tamper the signature.
	ct.MsgSig[0] ^= 0xFF

	c2 := deepCopy(c)
	c2.WriterEdPub = edPub
	c2.WriterMLDSAPub = mldsaPub

	_, err = channel.Decrypt(c2, ct)
	if err != channel.ErrBadMsgSig {
		t.Fatalf("want ErrBadMsgSig, got %v", err)
	}
}

// TestChannelLegacyMsgNoSig verifies legacy messages (no MsgSig) still decrypt
// successfully even when public keys are set.
func TestChannelLegacyMsgNoSig(t *testing.T) {
	edPub, _, _ := v0crypto.GenerateEd25519Keypair()
	mldsaPub, _, _ := v0crypto.GenerateMLDSA65Keypair()

	// Create channel without signing keys (legacy mode).
	c, _ := channel.NewChannel("scope-z", "alice")
	ct, err := channel.Encrypt(c, "alice", []byte("legacy msg"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ct.MsgSig != nil {
		t.Fatal("want nil MsgSig for legacy path")
	}

	// Decrypt with public keys set: should still succeed (backward compat).
	c2 := deepCopy(c)
	c2.WriterEdPub = edPub
	c2.WriterMLDSAPub = mldsaPub

	got, err := channel.Decrypt(c2, ct)
	if err != nil {
		t.Fatalf("Decrypt legacy msg: %v", err)
	}
	if !bytes.Equal(got, []byte("legacy msg")) {
		t.Fatalf("plaintext mismatch: %q", got)
	}
}

// --- snapshot tests ---

// TestHistorySnapshotBuildVerify verifies snapshot round-trip.
func TestHistorySnapshotBuildVerify(t *testing.T) {
	edPub, edPriv, err := v0crypto.GenerateEd25519Keypair()
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	mldsaPub, mldsaPriv, err := v0crypto.GenerateMLDSA65Keypair()
	if err != nil {
		t.Fatalf("generate mldsa: %v", err)
	}

	msgs := [][]byte{
		[]byte("msg body 1"),
		[]byte("msg body 2"),
		[]byte("msg body 3"),
	}
	snap, err := channel.BuildHistorySnapshot("chan-abc", 7, 100, msgs, edPriv, mldsaPriv)
	if err != nil {
		t.Fatalf("BuildHistorySnapshot: %v", err)
	}
	if len(snap.MerkleRoot) != 32 {
		t.Fatalf("MerkleRoot size: want 32, got %d", len(snap.MerkleRoot))
	}
	if snap.MessageCount != 3 {
		t.Fatalf("MessageCount: want 3, got %d", snap.MessageCount)
	}
	if snap.FromSeq != 100 || snap.ToSeq != 102 {
		t.Fatalf("seq range: want 100..102, got %d..%d", snap.FromSeq, snap.ToSeq)
	}

	if err := channel.VerifyHistorySnapshot(snap, edPub, mldsaPub); err != nil {
		t.Fatalf("VerifyHistorySnapshot: %v", err)
	}
}

// TestHistorySnapshotTamperedFails verifies tampered snapshots are rejected.
func TestHistorySnapshotTamperedFails(t *testing.T) {
	edPub, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	mldsaPub, mldsaPriv, _ := v0crypto.GenerateMLDSA65Keypair()

	msgs := [][]byte{[]byte("only message")}
	snap, _ := channel.BuildHistorySnapshot("chan-xyz", 1, 0, msgs, edPriv, mldsaPriv)

	// Tamper the signature.
	snap.Signature[0] ^= 0xFF
	if err := channel.VerifyHistorySnapshot(snap, edPub, mldsaPub); err == nil {
		t.Fatal("want verification failure for tampered signature")
	}
}

// TestHistorySnapshotMerkleRootMismatch verifies that a modified message hash is detected.
func TestHistorySnapshotMerkleRootMismatch(t *testing.T) {
	edPub, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	mldsaPub, mldsaPriv, _ := v0crypto.GenerateMLDSA65Keypair()

	msgs := [][]byte{[]byte("msg a"), []byte("msg b")}
	snap, _ := channel.BuildHistorySnapshot("chan-xyz", 2, 0, msgs, edPriv, mldsaPriv)

	// Tamper a message hash.
	snap.Messages[0].Hash[0] ^= 0xFF
	if err := channel.VerifyHistorySnapshot(snap, edPub, mldsaPub); err == nil {
		t.Fatal("want ErrSnapshotMerkle for tampered message hash")
	}
}

func deepCopy(c *channel.ChannelState) *channel.ChannelState {
	data, _ := channel.MarshalState(c)
	c2, _ := channel.UnmarshalState(data)
	return c2
}

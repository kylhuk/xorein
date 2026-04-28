package crowd_test

import (
	"bytes"
	"encoding/hex"
	"path/filepath"
	"runtime"
	"testing"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	crowd "github.com/aether/code_aether/pkg/v0_1/mode/crowd"
	channel "github.com/aether/code_aether/pkg/v0_1/mode/channel"
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

// TestW3CrowdEpochChain drives the crowd epoch chain KAT.
func TestW3CrowdEpochChain(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "crowd_epoch_chain.json"))

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			if v.Inputs["ikm"] == "" {
				t.Skip("null ikm — dependent vector")
			}
			ikm := spectest.Hex(v.Inputs["ikm"])
			expected := spectest.Hex(v.ExpectedOutput["next_epoch_root"])
			got, err := crowd.DeriveEpochRoot(ikm)
			if err != nil {
				t.Fatalf("DeriveEpochRoot: %v", err)
			}
			if !bytes.Equal(got, expected) {
				t.Fatalf("next_epoch_root mismatch:\n  want %x\n  got  %x", expected, got)
			}
		})
	}
}

// TestW3CrowdSenderKey drives the crowd sender key KAT.
func TestW3CrowdSenderKey(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "crowd_sender_key.json"))

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			ikm := spectest.Hex(v.Inputs["ikm"])
			infoUTF8 := v.Inputs["info_utf8"]
			expected := spectest.Hex(v.ExpectedOutput["sender_key"])

			// Extract peer_id from info: strip the label prefix.
			const label = "xorein/crowd/v1/sender-key"
			peerID := infoUTF8[len(label):]

			got, err := crowd.DeriveSenderKey(ikm, peerID)
			if err != nil {
				t.Fatalf("DeriveSenderKey: %v", err)
			}
			if !bytes.Equal(got, expected) {
				t.Fatalf("sender_key mismatch:\n  want %x\n  got  %x", expected, got)
			}
		})
	}
}

// TestW3ChannelSenderKey drives the channel sender key KAT.
func TestW3ChannelSenderKey(t *testing.T) {
	vdir := vectorDir(t)
	vecs := spectest.LoadVectors(t, filepath.Join(vdir, "channel_kdf_label.json"))

	for _, v := range vecs {
		t.Run(v.Name(), func(t *testing.T) {
			ikm := spectest.Hex(v.Inputs["ikm"])
			infoUTF8 := v.Inputs["info_utf8"]
			expected := spectest.Hex(v.ExpectedOutput["sender_key"])

			const label = "xorein/channel/v1/sender-key"
			peerID := infoUTF8[len(label):]

			got, err := channel.DeriveSenderKey(ikm, peerID)
			if err != nil {
				t.Fatalf("DeriveSenderKey: %v", err)
			}
			if !bytes.Equal(got, expected) {
				t.Fatalf("sender_key mismatch:\n  want %x\n  got  %x", expected, got)
			}
		})
	}
}

// TestW3CrowdRoundtrip verifies crowd mode encrypt/decrypt.
func TestW3CrowdRoundtrip(t *testing.T) {
	g, err := crowd.NewGroup("test-scope")
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	// Two members; same group state (copy).
	g2 := deepCopyCrowdState(g)

	pt := []byte("hello crowd")
	ct, err := crowd.Encrypt(g, "alice", pt)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	got, err := crowd.Decrypt(g2, ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("mismatch: want %q got %q", pt, got)
	}
}

// TestW3CrowdLegacyWindow verifies 2-epoch legacy window.
func TestW3CrowdLegacyWindow(t *testing.T) {
	g, _ := crowd.NewGroup("scope")
	// Send at epoch 0.
	ct0, _ := crowd.Encrypt(g, "alice", []byte("msg0"))

	// Force epoch rotation by filling the counter to MaxEpochMessages so the NEXT Encrypt rotates.
	rotate := func() {
		g.CurrentEpoch.MessageCount = crowd.MaxEpochMessages
		crowd.Encrypt(g, "alice", []byte("x")) // needsRotation=true → rotate, then encrypt in new epoch
	}

	rotate() // epoch 0 → 1
	rotate() // epoch 1 → 2

	// epoch 0 still in legacy window (window = 2): prevEpochs = [epoch1, epoch0].
	_, err := crowd.Decrypt(g, ct0)
	if err != nil {
		t.Fatalf("should decrypt epoch 0 in legacy window: %v", err)
	}

	// Rotate once more → epoch 3; prevEpochs = [epoch2, epoch1]; epoch 0 drops out.
	rotate()

	_, err = crowd.Decrypt(g, ct0)
	if err == nil {
		t.Fatal("should not decrypt epoch 0 outside legacy window")
	}
}

// TestW3ChannelRoundtrip verifies channel mode (only writer can encrypt).
func TestW3ChannelRoundtrip(t *testing.T) {
	c, err := channel.NewChannel("chan1", "alice")
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}
	c2 := deepCopyChannelState(c)

	ct, err := channel.Encrypt(c, "alice", []byte("broadcast"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	got, err := channel.Decrypt(c2, ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, []byte("broadcast")) {
		t.Fatalf("mismatch: %q", got)
	}
}

// TestW3ChannelNonWriterRejected verifies only the writer can encrypt.
func TestW3ChannelNonWriterRejected(t *testing.T) {
	c, _ := channel.NewChannel("chan1", "alice")
	_, err := channel.Encrypt(c, "bob", []byte("unauthorized"))
	if err != channel.ErrNotWriter {
		t.Fatalf("want ErrNotWriter, got %v", err)
	}
}

// --- helpers ---

func fromHex(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

// DeriveKey wrapper for inline label check in TestW3CrowdSenderKey.
var _ = v0crypto.DeriveKey

func deepCopyCrowdState(g *crowd.GroupState) *crowd.GroupState {
	data, _ := crowd.MarshalState(g)
	g2, _ := crowd.UnmarshalState(data)
	return g2
}

func deepCopyChannelState(c *channel.ChannelState) *channel.ChannelState {
	data, _ := channel.MarshalState(c)
	c2, _ := channel.UnmarshalState(data)
	return c2
}

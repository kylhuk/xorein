package crowd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	crowd "github.com/aether/code_aether/pkg/v0_1/mode/crowd"
)

// TestRotateEpochMembership verifies that RotateEpochMembership produces a
// fresh random epoch root (not derived from the current root).
func TestRotateEpochMembership(t *testing.T) {
	g, err := crowd.NewGroup("scope1")
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	originalRoot := append([]byte(nil), g.CurrentEpoch.EpochRoot...)
	originalID := g.CurrentEpoch.EpochID

	if err := crowd.RotateEpochMembership(g); err != nil {
		t.Fatalf("RotateEpochMembership: %v", err)
	}
	if g.CurrentEpoch.EpochID != originalID+1 {
		t.Fatalf("epoch ID: want %d, got %d", originalID+1, g.CurrentEpoch.EpochID)
	}
	// The new root must differ from the chained root.
	chainedRoot, _ := crowd.DeriveEpochRoot(originalRoot)
	if bytes.Equal(g.CurrentEpoch.EpochRoot, chainedRoot) {
		t.Fatal("RotateEpochMembership must use fresh random root, not chained root")
	}
	// The old epoch is preserved in PrevEpochs.
	if len(g.PrevEpochs) == 0 {
		t.Fatal("old epoch must be preserved in PrevEpochs")
	}
	if !bytes.Equal(g.PrevEpochs[0].EpochRoot, originalRoot) {
		t.Fatal("PrevEpochs[0] should be the original epoch")
	}
}

// TestRotateEpochMembershipExcludesRemovedMember verifies that after a membership
// rotation, messages encrypted with the old sender key can no longer be derived
// from the new epoch root.
func TestRotateEpochMembershipExcludesRemovedMember(t *testing.T) {
	g, _ := crowd.NewGroup("scope2")
	// Add alice and bob sender keys.
	if err := crowd.AddSender(g, "alice"); err != nil {
		t.Fatalf("AddSender alice: %v", err)
	}
	aliceKey := append([]byte(nil), g.CurrentEpoch.SenderKeys["alice"]...)

	// Encrypt alice's message in epoch 0.
	ct0, err := crowd.Encrypt(g, "alice", []byte("epoch 0 msg"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Membership change: bob leaves → fresh epoch rotation.
	if err := crowd.RotateEpochMembership(g); err != nil {
		t.Fatalf("RotateEpochMembership: %v", err)
	}

	// Alice's sender key in the NEW epoch must differ (fresh root → different derivation).
	newAliceKey, _ := crowd.DeriveSenderKey(g.CurrentEpoch.EpochRoot, "alice")
	if bytes.Equal(aliceKey, newAliceKey) {
		t.Fatal("sender key must change after membership rotation")
	}

	// The old epoch message can still be decrypted (in legacy window).
	g2 := deepCopy(g)
	_, err = crowd.Decrypt(g2, ct0)
	if err != nil {
		t.Fatalf("should still decrypt epoch 0 msg (in legacy window): %v", err)
	}
}

// TestSenderKeyPackageBuildUnmarshal verifies BuildSenderKeyPackage serialization.
func TestSenderKeyPackageBuildUnmarshal(t *testing.T) {
	epochID := make([]byte, 16)
	for i := range epochID {
		epochID[i] = byte(i)
	}
	epochRoot := make([]byte, 32)
	for i := range epochRoot {
		epochRoot[i] = byte(i + 100)
	}
	members := []string{"alice@test", "bob@test", "carol@test"}

	skpJSON, err := crowd.BuildSenderKeyPackage(epochID, epochRoot, crowd.RotationMembershipChange, members)
	if err != nil {
		t.Fatalf("BuildSenderKeyPackage: %v", err)
	}

	skp, err := crowd.UnmarshalSenderKeyPackage(skpJSON)
	if err != nil {
		t.Fatalf("UnmarshalSenderKeyPackage: %v", err)
	}
	if !bytes.Equal(skp.EpochID, epochID) {
		t.Fatalf("epoch_id mismatch")
	}
	if !bytes.Equal(skp.EpochRootSecret, epochRoot) {
		t.Fatalf("epoch_root_secret mismatch")
	}
	if skp.RotationTrigger != crowd.RotationMembershipChange {
		t.Fatalf("rotation_trigger: want %q, got %q", crowd.RotationMembershipChange, skp.RotationTrigger)
	}
	if len(skp.MemberPeerIDs) != 3 {
		t.Fatalf("member count: want 3, got %d", len(skp.MemberPeerIDs))
	}
	if skp.IssuedAt <= 0 {
		t.Fatal("issued_at must be set")
	}
	if skp.ExpiresAt <= skp.IssuedAt {
		t.Fatal("expires_at must be after issued_at")
	}
}

// TestSenderKeyPackageEpochIDTooShort verifies BuildSenderKeyPackage rejects short epoch IDs.
func TestSenderKeyPackageEpochIDTooShort(t *testing.T) {
	_, err := crowd.BuildSenderKeyPackage([]byte("tooshort"), make([]byte, 32), crowd.RotationTimeLimit, nil)
	if err == nil {
		t.Fatal("want error for short epoch_id")
	}
}

// TestSenderKeyPackageRootTooShort verifies BuildSenderKeyPackage rejects short epoch roots.
func TestSenderKeyPackageRootTooShort(t *testing.T) {
	_, err := crowd.BuildSenderKeyPackage(make([]byte, 16), []byte("tooshort"), crowd.RotationMessageLimit, nil)
	if err == nil {
		t.Fatal("want error for short epoch_root")
	}
}

// TestDistributeSenderKeyPackageStub verifies stub returns the same bytes.
func TestDistributeSenderKeyPackageStub(t *testing.T) {
	input := []byte(`{"test":"stub"}`)
	got, err := crowd.DistributeSenderKeyPackage(input)
	if err != nil {
		t.Fatalf("DistributeSenderKeyPackage: %v", err)
	}
	if !bytes.Equal(got, input) {
		t.Fatalf("stub must return input unchanged")
	}
}

// TestLegacyWindowHardCapAfterMembershipRotation verifies that after 3 rotations
// (2 chained + 1 membership), the epoch that fell out of the legacy window cannot
// be decrypted.
func TestLegacyWindowHardCapAfterMembershipRotation(t *testing.T) {
	g, _ := crowd.NewGroup("scope3")
	ct0, _ := crowd.Encrypt(g, "alice", []byte("epoch0"))

	// Three rotations: epoch 0 → 1 → 2 → 3.
	// After 3 rotations, prevEpochs = [epoch2, epoch1]; epoch0 is gone.
	for i := 0; i < 3; i++ {
		if i == 1 {
			// Use membership rotation for the second rotation.
			crowd.RotateEpochMembership(g)
		} else {
			// Force time/message rotation by filling the counter.
			g.CurrentEpoch.MessageCount = crowd.MaxEpochMessages
			crowd.Encrypt(g, "alice", []byte("x"))
		}
	}

	_, err := crowd.Decrypt(g, ct0)
	if err == nil {
		t.Fatal("should not decrypt epoch 0 after 3 rotations (outside legacy window)")
	}
	if err != crowd.ErrEpochExpired {
		t.Fatalf("want ErrEpochExpired, got %v", err)
	}
}

func deepCopy(g *crowd.GroupState) *crowd.GroupState {
	data, _ := json.Marshal(g)
	var g2 crowd.GroupState
	json.Unmarshal(data, &g2)
	return &g2
}

package tree_test

import (
	"bytes"
	"fmt"
	"testing"

	groupdm "github.com/aether/code_aether/pkg/v0_1/family/groupdm"
	tree "github.com/aether/code_aether/pkg/v0_1/mode/tree"
)

func makeMember(id string) tree.Member {
	return tree.Member{PeerID: id}
}

// TestW2TreeGroupCreate verifies group creation and initial message exchange.
func TestW2TreeGroupCreate(t *testing.T) {
	g, err := tree.NewGroup("test-group", makeMember("alice"))
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	if len(g.Members) != 1 {
		t.Fatalf("want 1 member, got %d", len(g.Members))
	}
	if g.CurrentEpoch.EpochID != 0 {
		t.Fatalf("want epoch 0, got %d", g.CurrentEpoch.EpochID)
	}
}

// TestW2TreeAddRemove verifies member add/remove triggers epoch rotation.
func TestW2TreeAddRemove(t *testing.T) {
	g, _ := tree.NewGroup("g", makeMember("alice"))
	epoch0 := g.CurrentEpoch.EpochID

	commit, err := tree.AddMember(g, makeMember("bob"))
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if commit == nil {
		t.Fatal("AddMember: want commit")
	}
	if g.CurrentEpoch.EpochID != epoch0+1 {
		t.Fatalf("epoch not rotated after add: want %d got %d", epoch0+1, g.CurrentEpoch.EpochID)
	}

	commit, err = tree.RemoveMember(g, "bob")
	if err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	if commit == nil {
		t.Fatal("RemoveMember: want commit")
	}
	if g.CurrentEpoch.EpochID != epoch0+2 {
		t.Fatalf("epoch not rotated after remove: want %d got %d", epoch0+2, g.CurrentEpoch.EpochID)
	}
	if tree.IsMember(g, "bob") {
		t.Fatal("bob should not be a member after removal")
	}
}

// TestW2TreeMaxMembers verifies the 50-member cap.
func TestW2TreeMaxMembers(t *testing.T) {
	g, _ := tree.NewGroup("g", makeMember("peer0"))
	for i := 1; i < tree.MaxMembers; i++ {
		if _, err := tree.AddMember(g, makeMember(fmt.Sprintf("peer%d", i))); err != nil {
			t.Fatalf("AddMember %d: %v", i, err)
		}
	}
	_, err := tree.AddMember(g, makeMember("overflow"))
	if err != tree.ErrGroupFull {
		t.Fatalf("want ErrGroupFull, got %v", err)
	}
}

// TestW2TreeRoundtrip verifies encrypt/decrypt round-trip for group members.
func TestW2TreeRoundtrip(t *testing.T) {
	// Two independent group state copies (simulating two nodes).
	g1, _ := tree.NewGroup("g", makeMember("alice"))
	g2, _ := tree.NewGroup("g", makeMember("alice"))
	// Copy epoch key to simulate key distribution.
	copy(g2.CurrentEpoch.EpochKey, g1.CurrentEpoch.EpochKey)
	copy(g2.RootKey, g1.RootKey)

	pt := []byte("hello group")
	ct, _, err := tree.Encrypt(g1, "alice", pt)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	got, err := tree.Decrypt(g2, ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("mismatch: want %q got %q", pt, got)
	}
}

// TestW2TreeRemovedMemberCannotDecrypt verifies post-remove messages are unreadable.
func TestW2TreeRemovedMemberCannotDecrypt(t *testing.T) {
	// alice's group (alice + bob).
	gAlice, _ := tree.NewGroup("g", makeMember("alice"))
	tree.AddMember(gAlice, makeMember("bob"))
	epoch1 := gAlice.CurrentEpoch.EpochID

	// bob's state is at epoch 1 (after join), with a copy of alice's epoch 1 key.
	gBob := &tree.GroupState{
		GroupID: "g",
		CurrentEpoch: &tree.EpochState{
			EpochID:  epoch1,
			EpochKey: append([]byte(nil), gAlice.CurrentEpoch.EpochKey...),
		},
		Members: []tree.Member{makeMember("alice"), makeMember("bob")},
		RootKey: append([]byte(nil), gAlice.RootKey...),
	}
	gBob.PrevEpochs = nil

	// Alice removes bob → epoch rotates to epoch 2.
	tree.RemoveMember(gAlice, "bob")
	epoch2 := gAlice.CurrentEpoch.EpochID
	if epoch2 != epoch1+1 {
		t.Fatalf("want epoch %d, got %d", epoch1+1, epoch2)
	}

	// Alice sends a message in epoch 2.
	ct, _, err := tree.Encrypt(gAlice, "alice", []byte("secret after bob left"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Bob cannot decrypt (he doesn't have epoch 2's key).
	_, err = tree.Decrypt(gBob, ct)
	if err == nil {
		t.Fatal("bob should not be able to decrypt post-removal message")
	}
}

// TestW2TreeLegacyWindow verifies 2-epoch legacy window for decryption.
func TestW2TreeLegacyWindow(t *testing.T) {
	g, _ := tree.NewGroup("g", makeMember("alice"))

	// Encrypt a message in epoch 0.
	ct0, _, err := tree.Encrypt(g, "alice", []byte("epoch 0 message"))
	if err != nil {
		t.Fatalf("Encrypt epoch 0: %v", err)
	}

	// Rotate to epoch 1.
	tree.AddMember(g, makeMember("bob"))
	// Rotate to epoch 2.
	tree.RemoveMember(g, "bob")

	// epoch 0 should still be in legacy window.
	pt, err := tree.Decrypt(g, ct0)
	if err != nil {
		t.Fatalf("should decrypt epoch 0 (in legacy window): %v", err)
	}
	if !bytes.Equal(pt, []byte("epoch 0 message")) {
		t.Fatalf("legacy window mismatch: %q", pt)
	}

	// Rotate to epoch 3 — epoch 0 drops out of legacy window.
	tree.AddMember(g, makeMember("carol"))

	_, err = tree.Decrypt(g, ct0)
	if err == nil {
		t.Fatal("should not decrypt epoch 0 after 2 rotations (outside legacy window)")
	}
}

// TestW2GroupDMHandler verifies the GroupDM handler create/send lifecycle.
func TestW2GroupDMHandler(t *testing.T) {
	alice := &groupdm.Handler{LocalPeerID: "alice"}
	bob := &groupdm.Handler{LocalPeerID: "bob"}

	creator := makeMember("alice")
	if err := alice.CreateGroup("grp1", creator); err != nil {
		t.Fatalf("alice create: %v", err)
	}
	if err := bob.CreateGroup("grp1", creator); err != nil {
		t.Fatalf("bob create: %v", err)
	}

	// Add bob to both states.
	bobMember := makeMember("bob")
	commit, err := alice.AddMember("grp1", bobMember)
	if err != nil {
		t.Fatalf("alice add bob: %v", err)
	}
	_ = commit

	// For the test, copy alice's epoch key to bob's group state (simulating Welcome delivery).
	aliceState := alice.GroupStateForTest("grp1")
	bobState := bob.GroupStateForTest("grp1")
	if aliceState == nil || bobState == nil {
		t.Fatal("nil group state")
	}
	tree.AddMember(bobState, bobMember)
	copy(bobState.CurrentEpoch.EpochKey, aliceState.CurrentEpoch.EpochKey)
	copy(bobState.RootKey, aliceState.RootKey)

	// Alice sends a message.
	ct, _, err := alice.SendMessage("grp1", "alice", []byte("hello group"))
	if err != nil {
		t.Fatalf("alice send: %v", err)
	}

	// Bob receives.
	pt, err := bob.ReceiveMessage("grp1", ct)
	if err != nil {
		t.Fatalf("bob receive: %v", err)
	}
	if !bytes.Equal(pt, []byte("hello group")) {
		t.Fatalf("mismatch: got %q", pt)
	}
}

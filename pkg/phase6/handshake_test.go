package phase6

import (
	"errors"
	"testing"
	"time"
)

func TestHandshakeMachineJoinSuccess(t *testing.T) {
	store := NewManifestStore(time.Hour)
	manifest := mustSignManifest(t, &Manifest{
		ServerID:     "handshake-success",
		Version:      3,
		UpdatedAt:    time.Now().UTC(),
		Capabilities: Capabilities{Chat: true, Voice: true},
	}, "remote-signer")
	if err := store.Publish(manifest); err != nil {
		t.Fatalf("publish manifest: %v", err)
	}
	machine := NewHandshakeMachine(store, "local-client")
	link := &DeepLink{ServerID: manifest.ServerID}
	state, err := machine.Join(link)
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	if state.Status != MembershipStatusActive {
		t.Fatalf("expected active status, got %s", state.Status)
	}
	if state.Manifest == nil || state.Manifest.ServerID != manifest.ServerID {
		t.Fatalf("manifest not retained in state")
	}
	if !state.ChatFlowEnabled() || !state.VoiceFlowEnabled() {
		t.Fatalf("chat/voice readiness flags not set as expected")
	}
	if state.LastHandshake.IsZero() {
		t.Fatalf("expected last handshake timestamp to be set")
	}

	state.Manifest.Description = "mutated externally"
	if machine.state[link.ServerID].Manifest.Description == "mutated externally" {
		t.Fatalf("returned state leaked internal manifest pointer")
	}

	state2, err := machine.Join(link)
	var hsErr *HandshakeError
	if !errors.As(err, &hsErr) || hsErr.Kind != HandshakeErrAlreadyJoined {
		t.Fatalf("expected already joined error, got %v", err)
	}
	if state2 == state {
		t.Fatalf("rejoining returned the same state pointer; expected detached clone")
	}
	if state2.ServerID != state.ServerID || state2.Status != state.Status {
		t.Fatalf("rejoining returned inconsistent state clone")
	}
}

func TestHandshakeMachineJoinFailureScenarios(t *testing.T) {
	store := NewManifestStore(time.Hour)
	machine := NewHandshakeMachine(store, "phase6")
	link := &DeepLink{ServerID: "missing-server"}
	if _, err := machine.Join(link); err == nil {
		t.Fatalf("expected error when manifest missing")
	}
	state := machine.state[link.ServerID]
	if state.Status != MembershipStatusFailed {
		t.Fatalf("status not failed after missing manifest, got %s", state.Status)
	}
	if state.LastError == "" {
		t.Fatalf("expected last error populated on failure")
	}
	if state.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", state.RetryCount)
	}
	if _, err := machine.Join(link); err == nil {
		t.Fatalf("expected error on second attempt missing manifest")
	}
	if state.RetryCount != 2 {
		t.Fatalf("retry count should increment on retries, got %d", state.RetryCount)
	}

	present := mustSignManifest(t, &Manifest{ServerID: "bad-signature", Version: ManifestVersionV1, UpdatedAt: time.Now().UTC(), Capabilities: Capabilities{Chat: true}}, "attacker")
	present.Signature = "tampered"
	store.mu.Lock()
	store.entries[present.ServerID] = &manifestEntry{manifest: present.Clone(), expiresAt: time.Now().UTC().Add(time.Hour)}
	store.mu.Unlock()
	link.ServerID = present.ServerID
	stateRetry, err := machine.Join(link)
	if err == nil {
		t.Fatalf("expected invalid signature error")
	}
	if stateRetry.Status != MembershipStatusFailed {
		t.Fatalf("status not failed after signature error")
	}
	if stateRetry.LastError != "invalid signature" {
		t.Fatalf("expected invalid signature error message, got %q", stateRetry.LastError)
	}
	var hsErr *HandshakeError
	if !errors.As(err, &hsErr) || hsErr.Kind != HandshakeErrInvalidSignature {
		t.Fatalf("expected invalid signature handshake error, got %v", err)
	}
}

package phase7

import "testing"

func TestBootstrapperRotationAndCompatibility(t *testing.T) {
	b := NewBootstrapper()
	id := ParticipantID("alice")
	state, err := b.Bootstrap(id)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	if state.Rotation != 1 {
		t.Fatalf("expected initial rotation 1, got %d", state.Rotation)
	}
	second, err := b.Bootstrap(id)
	if err != nil {
		t.Fatalf("second bootstrap failed: %v", err)
	}
	if state != second {
		t.Fatalf("expected repeated bootstrap to return cached state")
	}
	legacy := cloneKey(state.SenderKey)
	oldRotation := state.Rotation
	rotated, err := b.Rotate(id)
	if err != nil {
		t.Fatalf("rotation failed: %v", err)
	}
	if rotated.Rotation != oldRotation+1 {
		t.Fatalf("expected rotation %d, got %d", oldRotation+1, rotated.Rotation)
	}
	if len(rotated.LegacySender) == 0 {
		t.Fatalf("expected legacy senders after rotation")
	}
	if !b.SenderCompatible(id, legacy) {
		t.Fatalf("expected legacy sender compatibility")
	}
	if b.SenderCompatible(id, []byte("unknown")) {
		t.Fatalf("expected unknown candidate to be incompatible")
	}
}

func TestBootstrapperRekeyOnMismatch(t *testing.T) {
	tests := []struct {
		name      string
		candidate func(*KeyState) []byte
		delta     uint64
	}{
		{
			name: "match",
			candidate: func(s *KeyState) []byte {
				return cloneKey(s.SenderKey)
			},
			delta: 0,
		},
		{
			name: "mismatch",
			candidate: func(*KeyState) []byte {
				return []byte("mismatch-send-key-candidate-000000")
			},
			delta: 1,
		},
		{
			name:      "empty",
			candidate: func(*KeyState) []byte { return nil },
			delta:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := ParticipantID("bob")
			b := NewBootstrapper()
			state, err := b.Bootstrap(id)
			if err != nil {
				t.Fatalf("bootstrap failed: %v", err)
			}
			candidate := tt.candidate(state)
			newState, err := b.RekeyOnMismatch(id, candidate)
			if err != nil {
				t.Fatalf("rekey failed: %v", err)
			}
			if newState.Rotation != state.Rotation+tt.delta {
				t.Fatalf("unexpected rotation delta, want %d got %d", tt.delta, newState.Rotation-state.Rotation)
			}
		})
	}
}

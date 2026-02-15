package phase7

import (
	"bytes"
	"errors"
	"testing"
	"testing/quick"
)

func TestPipelineRoundTripQuick(t *testing.T) {
	cfg := &quick.Config{MaxCount: 64}
	property := func(payload []byte, seq uint16) bool {
		b := NewBootstrapper()
		state, err := b.Bootstrap(ParticipantID("quick-sender"))
		if err != nil {
			return false
		}

		sender := NewPipeline(ParticipantID("quick-sender"), state)
		receiver := NewPipeline(ParticipantID("quick-receiver"), state)

		msg, err := sender.Send(uint64(seq)+1, payload)
		if err != nil {
			return false
		}
		got, err := receiver.Receive(msg, state.MLSSecret, state.Verifier)
		if err != nil {
			return false
		}
		return bytes.Equal(got, payload)
	}

	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("quick round-trip property failed: %v", err)
	}
}

func TestPipelineTwoPeerBidirectionalIntegration(t *testing.T) {
	b := NewBootstrapper()
	aliceState, err := b.Bootstrap(ParticipantID("alice"))
	if err != nil {
		t.Fatalf("alice bootstrap failed: %v", err)
	}
	bobState, err := b.Bootstrap(ParticipantID("bob"))
	if err != nil {
		t.Fatalf("bob bootstrap failed: %v", err)
	}

	// Simulate channel key distribution by aligning the MLS secret across peers.
	bobState.MLSSecret = cloneKey(aliceState.MLSSecret)

	alice := NewPipeline(ParticipantID("alice"), aliceState)
	bob := NewPipeline(ParticipantID("bob"), bobState)

	aliceMsg, err := alice.Send(1, []byte("hello-from-alice"))
	if err != nil {
		t.Fatalf("alice send failed: %v", err)
	}
	bobPayload, err := bob.Receive(aliceMsg, bobState.MLSSecret, aliceState.Verifier)
	if err != nil {
		t.Fatalf("bob receive failed: %v", err)
	}
	if string(bobPayload) != "hello-from-alice" {
		t.Fatalf("unexpected bob payload: %q", bobPayload)
	}

	bobMsg, err := bob.Send(1, []byte("hello-from-bob"))
	if err != nil {
		t.Fatalf("bob send failed: %v", err)
	}
	alicePayload, err := alice.Receive(bobMsg, aliceState.MLSSecret, bobState.Verifier)
	if err != nil {
		t.Fatalf("alice receive failed: %v", err)
	}
	if string(alicePayload) != "hello-from-bob" {
		t.Fatalf("unexpected alice payload: %q", alicePayload)
	}
}

func TestPipelineMultiPeerOutOfOrderAndDuplicateBehavior(t *testing.T) {
	b := NewBootstrapper()
	aliceState, err := b.Bootstrap(ParticipantID("alice"))
	if err != nil {
		t.Fatalf("alice bootstrap failed: %v", err)
	}
	charlieState, err := b.Bootstrap(ParticipantID("charlie"))
	if err != nil {
		t.Fatalf("charlie bootstrap failed: %v", err)
	}

	charlieState.MLSSecret = cloneKey(aliceState.MLSSecret)

	alice := NewPipeline(ParticipantID("alice"), aliceState)
	charlie := NewPipeline(ParticipantID("charlie"), charlieState)

	second, err := charlie.Send(2, []byte("second"))
	if err != nil {
		t.Fatalf("charlie send(2) failed: %v", err)
	}
	first, err := charlie.Send(1, []byte("first"))
	if err != nil {
		t.Fatalf("charlie send(1) failed: %v", err)
	}

	payload2, err := alice.Receive(second, aliceState.MLSSecret, charlieState.Verifier)
	if err != nil {
		t.Fatalf("receive second failed: %v", err)
	}
	if string(payload2) != "second" {
		t.Fatalf("unexpected payload for second message: %q", payload2)
	}

	payload1, err := alice.Receive(first, aliceState.MLSSecret, charlieState.Verifier)
	if err != nil {
		t.Fatalf("receive first failed: %v", err)
	}
	if string(payload1) != "first" {
		t.Fatalf("unexpected payload for first message: %q", payload1)
	}

	if _, err := alice.Receive(second, aliceState.MLSSecret, charlieState.Verifier); !errors.Is(err, ErrDuplicateMessage) {
		t.Fatalf("expected duplicate rejection for repeated message, got %v", err)
	}
}

func TestPipelineReceiveFailsWithWrongSecret(t *testing.T) {
	b := NewBootstrapper()
	aliceState, err := b.Bootstrap(ParticipantID("alice"))
	if err != nil {
		t.Fatalf("alice bootstrap failed: %v", err)
	}
	bobState, err := b.Bootstrap(ParticipantID("bob"))
	if err != nil {
		t.Fatalf("bob bootstrap failed: %v", err)
	}

	bobState.MLSSecret = cloneKey(aliceState.MLSSecret)

	alice := NewPipeline(ParticipantID("alice"), aliceState)
	bob := NewPipeline(ParticipantID("bob"), bobState)
	msg, err := alice.Send(5, []byte("confidential"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	wrongSecret := bytes.Repeat([]byte{0x42}, 32)
	if _, err := bob.Receive(msg, wrongSecret, aliceState.Verifier); err == nil {
		t.Fatalf("expected decryption failure with wrong secret")
	}
}

package phase7

import "testing"

func TestPipelineSendReceive(t *testing.T) {
	b := NewBootstrapper()
	id := ParticipantID("alice-p7")
	state, err := b.Bootstrap(id)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	sender := NewPipeline(id, state)
	plaintext := []byte("phase7 pipeline payload")
	msg, err := sender.Send(42, plaintext)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	receiver := NewPipeline(ParticipantID("receiver"), state)
	got, err := receiver.Receive(msg, state.MLSSecret, state.Verifier)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("unexpected plaintext, want %q got %q", plaintext, got)
	}
}

func TestPipelineReceiveInvalidSignature(t *testing.T) {
	b := NewBootstrapper()
	id := ParticipantID("alice-p7")
	state, err := b.Bootstrap(id)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	sender := NewPipeline(id, state)
	msg, err := sender.Send(1, []byte("payload"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	msg.Signature[0] ^= 0xFF
	receiver := NewPipeline(ParticipantID("receiver"), state)
	if _, err := receiver.Receive(msg, state.MLSSecret, state.Verifier); err != ErrInvalidSignature {
		t.Fatalf("expected invalid signature error, got %v", err)
	}
}

func TestPipelineReceiveRejectsDuplicates(t *testing.T) {
	b := NewBootstrapper()
	id := ParticipantID("alice-p7")
	state, err := b.Bootstrap(id)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	sender := NewPipeline(id, state)
	msg, err := sender.Send(7, []byte("duplicate check"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	receiver := NewPipeline(ParticipantID("receiver"), state)
	if _, err := receiver.Receive(msg, state.MLSSecret, state.Verifier); err != nil {
		t.Fatalf("first receive failed: %v", err)
	}
	if _, err := receiver.Receive(msg, state.MLSSecret, state.Verifier); err != ErrDuplicateMessage {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

package voice

import "testing"

func TestVoiceSessionFallbackAndBackoff(t *testing.T) {
	s := NewSession()

	if got := s.CurrentFallback(); got != FallbackDirect {
		t.Fatalf("expected direct fallback, got %q", got)
	}

	for i := 1; i < len(s.Ladder()); i++ {
		s.AdvanceFallback()
	}

	if got := s.CurrentFallback(); got != FallbackTURN {
		t.Fatalf("expected last fallback turn, got %q", got)
	}

	backoff1 := s.Reconnect()
	backoff2 := s.Reconnect()
	if backoff1 >= backoff2 {
		t.Fatalf("backoff should increase, got %v >= %v", backoff1, backoff2)
	}

	s.Reset()
	if s.CurrentFallback() != FallbackDirect {
		t.Fatalf("expected fallback reset to direct")
	}
}

func TestSessionNegotiation(t *testing.T) {
	s := NewSession()
	codec, transport, err := s.Negotiate([]string{"opus/16000/2", "opus/48000/2"}, []string{"tcp", "udp"})
	if err != nil {
		t.Fatalf("unexpected negotiation error: %v", err)
	}
	if codec != "opus/48000/2" {
		t.Fatalf("expected best codec, got %q", codec)
	}
	if transport != "udp" {
		t.Fatalf("expected udp transport, got %q", transport)
	}
}

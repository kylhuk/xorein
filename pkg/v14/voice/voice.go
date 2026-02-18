package voice

import (
	"fmt"
	"time"
)

// FallbackStep enumerates deterministic voice path priorities.
type FallbackStep string

const (
	FallbackDirect FallbackStep = "direct"
	FallbackMesh   FallbackStep = "mesh"
	FallbackSFU    FallbackStep = "sfu"
	FallbackTURN   FallbackStep = "turn"
)

var (
	codecPreference     = []string{"opus/48000/2", "opus/24000/2", "opus/16000/2"}
	transportPreference = []string{"udp", "tcp", "webrtc"}
)

// VoiceSession keeps deterministic negotiation state.
type VoiceSession struct {
	fallbackSteps []FallbackStep
	fallbackIndex int
	backoff       []time.Duration
	attempt       int
	codec         string
	transport     string
}

// NewSession builds a session with deterministic ladder/backoff.
func NewSession() *VoiceSession {
	return &VoiceSession{
		fallbackSteps: []FallbackStep{FallbackDirect, FallbackMesh, FallbackSFU, FallbackTURN},
		backoff:       []time.Duration{250 * time.Millisecond, 500 * time.Millisecond, 750 * time.Millisecond, 1 * time.Second},
	}
}

// Negotiate returns the highest priority codec and transport intersection.
func (s *VoiceSession) Negotiate(codecs, transports []string) (string, string, error) {
	codec := intersectPreferred(codecPreference, codecs)
	if codec == "" && len(codecs) > 0 {
		codec = codecs[0]
	}
	transport := intersectPreferred(transportPreference, transports)
	if transport == "" && len(transports) > 0 {
		transport = transports[0]
	}

	if codec == "" || transport == "" {
		return "", "", fmt.Errorf("no compatible codec (%q) or transport (%q)", codec, transport)
	}

	s.codec = codec
	s.transport = transport
	s.attempt = 0
	return codec, transport, nil
}

func intersectPreferred(preferred, candidates []string) string {
	for _, pref := range preferred {
		for _, candidate := range candidates {
			if pref == candidate {
				return candidate
			}
		}
	}
	return ""
}

// CurrentFallback reports the currently selected fallback step.
func (s *VoiceSession) CurrentFallback() FallbackStep {
	if s.fallbackIndex >= len(s.fallbackSteps) {
		return s.fallbackSteps[len(s.fallbackSteps)-1]
	}
	return s.fallbackSteps[s.fallbackIndex]
}

// AdvanceFallback moves to the next fallback and returns it.
func (s *VoiceSession) AdvanceFallback() FallbackStep {
	if s.fallbackIndex < len(s.fallbackSteps)-1 {
		s.fallbackIndex++
	}
	return s.CurrentFallback()
}

// Reconnect calculates the next backoff while incrementing attempt counter.
func (s *VoiceSession) Reconnect() time.Duration {
	s.attempt++
	idx := s.attempt - 1
	if idx >= len(s.backoff) {
		idx = len(s.backoff) - 1
	}
	return s.backoff[idx]
}

// Reset reconnect/backoff counters when the session stabilizes.
func (s *VoiceSession) Reset() {
	s.attempt = 0
	s.fallbackIndex = 0
}

// Ladder returns a copy of the fallback order.
func (s *VoiceSession) Ladder() []FallbackStep {
	ladder := make([]FallbackStep, len(s.fallbackSteps))
	copy(ladder, s.fallbackSteps)
	return ladder
}

package push

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	MaxPayloadBytes = 256 * 1024
	AuthMetadataKey = "push.auth_token"
)

type Registration struct {
	DeviceID  string
	ClientID  string
	Token     string
	CreatedAt time.Time
	Metadata  map[string]string
}

type PayloadEnvelope struct {
	ID        string
	Recipient string
	Payload   []byte
	Metadata  map[string]string
	CreatedAt time.Time
}

type RelayBlind interface {
	Register(ctx context.Context, reg Registration) error
	Forward(ctx context.Context, env PayloadEnvelope) error
}

type Provider interface {
	Send(ctx context.Context, env PayloadEnvelope) error
}

type Service struct {
	adapter   Provider
	mu        sync.Mutex
	regs      map[string]Registration
	forwarded map[string]struct{}
}

func NewService(adapter Provider) *Service {
	return &Service{adapter: adapter, regs: make(map[string]Registration), forwarded: make(map[string]struct{})}
}

func (s *Service) Register(ctx context.Context, reg Registration) error {
	if reg.DeviceID == "" || reg.ClientID == "" {
		return fmt.Errorf("push.registration.invalid")
	}
	if reg.Token == "" && !hasAuthToken(reg.Metadata, "") {
		return fmt.Errorf("push.registration.unauthorized")
	}
	reg.Metadata = cloneMetadata(reg.Metadata)
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.regs[reg.DeviceID]; ok {
		if existing.ClientID == reg.ClientID && existing.Token == reg.Token && metadataEqual(existing.Metadata, reg.Metadata) {
			return nil
		}
	}
	s.regs[reg.DeviceID] = reg
	return nil
}

func (s *Service) Forward(ctx context.Context, env PayloadEnvelope) error {
	if s.adapter == nil {
		return fmt.Errorf("push.adapter.missing")
	}
	if env.ID == "" || env.Recipient == "" {
		return fmt.Errorf("push.payload.invalid")
	}
	if len(env.Payload) > MaxPayloadBytes {
		return fmt.Errorf("push.payload.too_large")
	}
	s.mu.Lock()
	if _, ok := s.forwarded[env.ID]; ok {
		s.mu.Unlock()
		return nil
	}
	reg, ok := s.regs[env.Recipient]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("push.recipient.unknown")
	}
	if !hasAuthToken(env.Metadata, reg.Token) {
		s.mu.Unlock()
		return fmt.Errorf("push.payload.unauthorized")
	}
	s.forwarded[env.ID] = struct{}{}
	s.mu.Unlock()
	return s.adapter.Send(ctx, env)
}

func BuildPayload(id, recipient string, payload []byte, metadata map[string]string) PayloadEnvelope {
	copied := make(map[string]string)
	for k, v := range metadata {
		copied[k] = v
	}
	return PayloadEnvelope{ID: id, Recipient: recipient, Payload: payload, Metadata: copied, CreatedAt: time.Now().UTC()}
}

func hasAuthToken(metadata map[string]string, expected string) bool {
	if metadata == nil {
		return false
	}
	token := metadata[AuthMetadataKey]
	if token == "" {
		return false
	}
	if expected == "" {
		return true
	}
	return token == expected
}

func metadataEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func cloneMetadata(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}
	clone := make(map[string]string, len(metadata))
	for k, v := range metadata {
		clone[k] = v
	}
	return clone
}

type MockAdapter struct {
	Sent []PayloadEnvelope
}

func (m *MockAdapter) Send(ctx context.Context, env PayloadEnvelope) error {
	m.Sent = append(m.Sent, env)
	return nil
}

type RelayAdapter struct {
	Name     string
	Provider Provider
}

func (r *RelayAdapter) Send(ctx context.Context, env PayloadEnvelope) error {
	return r.Provider.Send(ctx, env)
}

func NewAdapter(name string, provider Provider) *RelayAdapter {
	return &RelayAdapter{Name: name, Provider: provider}
}

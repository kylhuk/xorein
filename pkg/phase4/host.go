package phase4

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// TransportID identifies the transport flavor the host can launch.
type TransportID string

const (
	TransportQUIC TransportID = "quic"
	TransportTCP  TransportID = "tcp"
)

// SecurityConfig describes an optional or required security stack element.
type SecurityConfig struct {
	Name     string
	Required bool
}

// MuxerConfig describes the multiplexer negotiation expectations.
type MuxerConfig struct {
	Name     string
	Required bool
}

// HostConfig drives deterministic baseline host creation for P4-T1.
type HostConfig struct {
	IdentityKey    []byte
	TransportOrder []TransportID
	Security       []SecurityConfig
	Multiplexers   []MuxerConfig
}

// TransportProbe lets the caller simulate transport availability during startup.
type TransportProbe func(context.Context, TransportID) bool

// HostService owns the deterministic startup path and exposes startup reports.
type HostService struct {
	cfg      HostConfig
	identity ed25519.PrivateKey
	probe    TransportProbe
}

// TransportReport describes how a single transport evaluation completed.
type TransportReport struct {
	ID        TransportID
	Available bool
}

// StartupReport contains the negotiations and logs emitted during host startup.
type StartupReport struct {
	TransportDetails    []TransportReport
	SecurityStack       []SecurityConfig
	SelectedMultiplexer MuxerConfig
	Logs                []string
	IdentityFingerprint string
}

// NewHostService constructs the deterministic host module from configuration.
func NewHostService(cfg HostConfig, probe TransportProbe) (*HostService, error) {
	if len(cfg.IdentityKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("phase4: identity private key must be %d bytes", ed25519.PrivateKeySize)
	}
	identity := make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	copy(identity, cfg.IdentityKey)
	cfgCopy := cfg
	cfgCopy.TransportOrder = append([]TransportID(nil), cfg.TransportOrder...)
	cfgCopy.Security = append([]SecurityConfig(nil), cfg.Security...)
	cfgCopy.Multiplexers = append([]MuxerConfig(nil), cfg.Multiplexers...)
	return &HostService{
		cfg:      cfgCopy,
		identity: identity,
		probe:    probe,
	}, nil
}

// Start attempts to bring up transports, security, and mux stacks, returning logs.
func (s *HostService) Start(ctx context.Context) (StartupReport, error) {
	var logs []string
	transports := s.transportOrder()
	reports := make([]TransportReport, 0, len(transports))
	successCount := 0
	for _, tid := range transports {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			logs = append(logs, fmt.Sprintf("startup canceled before checking %s: %v", tid, err))
			return StartupReport{TransportDetails: reports, Logs: logs}, err
		default:
		}

		available := true
		if s.probe != nil {
			available = s.probe(ctx, tid)
		}
		reports = append(reports, TransportReport{ID: tid, Available: available})
		if available {
			successCount++
			logs = append(logs, fmt.Sprintf("transport %s available", tid))
		} else {
			logs = append(logs, fmt.Sprintf("transport %s unavailable", tid))
		}
	}
	if successCount == 0 {
		err := fmt.Errorf("phase4: no transports available")
		logs = append(logs, fmt.Sprintf("host startup failed: %s", err))
		return StartupReport{TransportDetails: reports, Logs: logs}, err
	}

	securityStack := s.securityStack()
	logs = append(logs, fmt.Sprintf("security stack %s", formatSecurityStack(securityStack)))
	muxer := s.selectMuxer()
	if muxer.Name == "" {
		logs = append(logs, "no multiplexer configured")
	} else {
		req := "optional"
		if muxer.Required {
			req = "required"
		}
		logs = append(logs, fmt.Sprintf("multiplexer selected %s (%s)", muxer.Name, req))
	}

	fingerprint := identityFingerprint(s.identity)
	logs = append(logs, fmt.Sprintf("identity fingerprint %s", fingerprint))

	report := StartupReport{
		TransportDetails:    reports,
		SecurityStack:       securityStack,
		SelectedMultiplexer: muxer,
		Logs:                append([]string(nil), logs...),
		IdentityFingerprint: fingerprint,
	}
	return report, nil
}

func (s *HostService) transportOrder() []TransportID {
	if len(s.cfg.TransportOrder) > 0 {
		order := make([]TransportID, len(s.cfg.TransportOrder))
		copy(order, s.cfg.TransportOrder)
		return order
	}
	return []TransportID{TransportQUIC, TransportTCP}
}

func (s *HostService) securityStack() []SecurityConfig {
	if len(s.cfg.Security) > 0 {
		stack := make([]SecurityConfig, len(s.cfg.Security))
		copy(stack, s.cfg.Security)
		return stack
	}
	return []SecurityConfig{{Name: "noise", Required: true}}
}

func (s *HostService) selectMuxer() MuxerConfig {
	if len(s.cfg.Multiplexers) == 0 {
		return MuxerConfig{Name: "yamux", Required: true}
	}
	for _, candidate := range s.cfg.Multiplexers {
		if candidate.Required {
			return candidate
		}
	}
	return s.cfg.Multiplexers[0]
}

func formatSecurityStack(stack []SecurityConfig) string {
	if len(stack) == 0 {
		return "<none>"
	}
	entries := make([]string, len(stack))
	for i, sec := range stack {
		req := "optional"
		if sec.Required {
			req = "required"
		}
		entries[i] = fmt.Sprintf("%s(%s)", sec.Name, req)
	}
	return strings.Join(entries, ", ")
}

func identityFingerprint(priv ed25519.PrivateKey) string {
	pub := priv.Public().(ed25519.PublicKey)
	digest := sha256.Sum256(pub)
	return hex.EncodeToString(digest[:])
}

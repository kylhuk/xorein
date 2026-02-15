package phase4

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"testing"
)

var testIdentitySeed = [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

func identityKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	return ed25519.NewKeyFromSeed(testIdentitySeed[:])
}

func mustStartHost(t *testing.T, cfg HostConfig, probe TransportProbe) StartupReport {
	t.Helper()
	svc, err := NewHostService(cfg, probe)
	if err != nil {
		t.Fatalf("NewHostService: %v", err)
	}
	report, err := svc.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return report
}

func assertLogContains(t *testing.T, logs []string, want string) {
	t.Helper()
	for _, log := range logs {
		if log == want {
			return
		}
	}
	t.Fatalf("logs do not contain %q: %v", want, logs)
}

func TestHostServiceDefaultBaseline(t *testing.T) {
	cfg := HostConfig{IdentityKey: identityKey(t)}
	report := mustStartHost(t, cfg, nil)

	if len(report.TransportDetails) != 2 {
		t.Fatalf("expected 2 transports, got %d", len(report.TransportDetails))
	}
	if got, want := report.TransportDetails[0].ID, TransportQUIC; got != want {
		t.Fatalf("transport[0] = %s, want %s", got, want)
	}
	if got, want := report.TransportDetails[1].ID, TransportTCP; got != want {
		t.Fatalf("transport[1] = %s, want %s", got, want)
	}
	for _, detail := range report.TransportDetails {
		if !detail.Available {
			t.Fatalf("transport %s should be available", detail.ID)
		}
	}
	wantSecurity := []SecurityConfig{{Name: "noise", Required: true}}
	if !equalSecurityStacks(report.SecurityStack, wantSecurity) {
		t.Fatalf("security stack = %v, want %v", report.SecurityStack, wantSecurity)
	}
	selected := report.SelectedMultiplexer
	wantMuxer := MuxerConfig{Name: "yamux", Required: true}
	if selected != wantMuxer {
		t.Fatalf("selected mux = %v, want %v", selected, wantMuxer)
	}
	assertLogContains(t, report.Logs, fmt.Sprintf("transport %s available", TransportQUIC))
	assertLogContains(t, report.Logs, fmt.Sprintf("transport %s available", TransportTCP))
	assertLogContains(t, report.Logs, "security stack noise(required)")
	assertLogContains(t, report.Logs, "multiplexer selected yamux (required)")
	assertLogContains(t, report.Logs, fmt.Sprintf("identity fingerprint %s", identityFingerprint(ed25519.NewKeyFromSeed(testIdentitySeed[:]))))
}

func TestHostServicePartialTransports(t *testing.T) {
	cfg := HostConfig{IdentityKey: identityKey(t)}
	probe := func(_ context.Context, tid TransportID) bool {
		switch tid {
		case TransportQUIC:
			return false
		case TransportTCP:
			return true
		default:
			return true
		}
	}
	report := mustStartHost(t, cfg, probe)

	if len(report.TransportDetails) != 2 {
		t.Fatalf("expected 2 transports, got %d", len(report.TransportDetails))
	}
	available := 0
	for _, detail := range report.TransportDetails {
		if detail.Available {
			available++
		}
	}
	if available != 1 {
		t.Fatalf("expected exactly 1 available transport, got %d", available)
	}
	assertLogContains(t, report.Logs, "transport quic unavailable")
	assertLogContains(t, report.Logs, "transport tcp available")
}

func TestHostServiceNoTransports(t *testing.T) {
	cfg := HostConfig{IdentityKey: identityKey(t)}
	svc, err := NewHostService(cfg, func(context.Context, TransportID) bool { return false })
	if err != nil {
		t.Fatalf("NewHostService: %v", err)
	}
	_, err = svc.Start(context.Background())
	if err == nil || err.Error() != "phase4: no transports available" {
		t.Fatalf("Start error = %v, want phase4: no transports available", err)
	}
}

func TestHostServiceInvalidIdentityKey(t *testing.T) {
	if _, err := NewHostService(HostConfig{IdentityKey: []byte{1, 2}}, nil); err == nil {
		t.Fatal("expected error for invalid identity key length")
	}
}

func TestHostServiceCustomMuxSecurity(t *testing.T) {
	cfg := HostConfig{
		IdentityKey: identityKey(t),
		Security: []SecurityConfig{
			{Name: "noise", Required: false},
			{Name: "tls", Required: true},
		},
		Multiplexers: []MuxerConfig{
			{Name: "mplex", Required: false},
			{Name: "yamux", Required: true},
		},
	}
	report := mustStartHost(t, cfg, nil)
	if !equalSecurityStacks(report.SecurityStack, cfg.Security) {
		t.Fatalf("security stack = %v, want %v", report.SecurityStack, cfg.Security)
	}
	if got, want := report.SelectedMultiplexer, cfg.Multiplexers[1]; got != want {
		t.Fatalf("selected mux = %v, want %v", got, want)
	}
	assertLogContains(t, report.Logs, "security stack noise(optional), tls(required)")
	assertLogContains(t, report.Logs, "multiplexer selected yamux (required)")
}

func equalSecurityStacks(a, b []SecurityConfig) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

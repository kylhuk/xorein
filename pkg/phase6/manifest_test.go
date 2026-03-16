package phase6

import (
	"errors"
	"testing"
	"time"
)

func TestManifestSerializeSignAndValidate(t *testing.T) {
	baseTime := time.Date(2025, time.December, 31, 23, 59, 59, 0, time.UTC)
	m := &Manifest{
		ServerID:    "server-serialization",
		Version:     ManifestVersionV1,
		Description: "deterministic payload",
		UpdatedAt:   baseTime,
		Capabilities: Capabilities{
			Chat:  true,
			Voice: false,
		},
	}

	if _, err := (*Manifest)(nil).Serialize(); !errors.Is(err, ErrManifestNil) {
		t.Fatalf("expected ErrManifestNil serializing nil manifest, got %v", err)
	}

	first, err := m.Serialize()
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}
	second, err := m.Serialize()
	if err != nil {
		t.Fatalf("serialize failed second time: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("manifest serialization is not deterministic\nfirst: %s\nsecond: %s", first, second)
	}

	if _, err := m.Sign(" gatekeeper "); err != nil {
		t.Fatalf("signing manifest failed: %v", err)
	}
	if m.Identity != "gatekeeper" {
		t.Fatalf("expected trimmed identity, got %q", m.Identity)
	}
	if m.Signature == "" {
		t.Fatalf("expected signature to be populated")
	}
	if !m.ValidateSignature("gatekeeper") {
		t.Fatalf("valid signature reported as invalid")
	}
	if m.ValidateSignature("intruder") {
		t.Fatalf("manifest accepted wrong identity")
	}
	if err := m.ValidateStoredSignature(); err != nil {
		t.Fatalf("stored signature validation failed: %v", err)
	}

	invalid := &Manifest{ServerID: "another"}
	if _, err := invalid.Sign(""); !errors.Is(err, ErrManifestIdentityRequired) {
		t.Fatalf("expected ErrManifestIdentityRequired when signing without identity, got %v", err)
	}
	invalid.Identity = "somebody"
	invalid.Signature = "abc"
	if err := invalid.ValidateStoredSignature(); !errors.Is(err, ErrManifestVersionInvalid) {
		t.Fatalf("expected ErrManifestVersionInvalid on incomplete manifest, got %v", err)
	}
}

func TestManifestStoredSignatureValidationFailures(t *testing.T) {
	base := mustSignManifest(t, &Manifest{
		ServerID:    "signed-server",
		Version:     ManifestVersionV1,
		Description: "signed",
		UpdatedAt:   time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}, "remote-signer")

	cases := []struct {
		name    string
		mutate  func(*Manifest)
		wantErr error
	}{
		{
			name: "missing identity",
			mutate: func(m *Manifest) {
				m.Identity = ""
			},
			wantErr: ErrManifestStoredIdentityNeeded,
		},
		{
			name: "missing signature",
			mutate: func(m *Manifest) {
				m.Signature = ""
			},
			wantErr: ErrManifestSignatureRequired,
		},
		{
			name: "tampered payload",
			mutate: func(m *Manifest) {
				m.Description = "tampered"
			},
			wantErr: ErrManifestInvalidSignature,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := base.Clone()
			tc.mutate(candidate)
			if err := candidate.ValidateStoredSignature(); !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestManifestFieldAndFreshnessValidation(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		m       *Manifest
		wantErr error
	}{
		{name: "nil manifest", m: nil, wantErr: ErrManifestNil},
		{name: "missing server id", m: &Manifest{Version: ManifestVersionV1, UpdatedAt: now}, wantErr: ErrManifestServerIDRequired},
		{name: "invalid version", m: &Manifest{ServerID: "srv", Version: 0, UpdatedAt: now}, wantErr: ErrManifestVersionInvalid},
		{name: "missing timestamp", m: &Manifest{ServerID: "srv", Version: ManifestVersionV1}, wantErr: ErrManifestUpdatedAtRequired},
		{name: "valid", m: &Manifest{ServerID: "srv", Version: ManifestVersionV1, UpdatedAt: now}, wantErr: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.m.ValidateFields()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}

	fresh := &Manifest{ServerID: "fresh", Version: ManifestVersionV1, UpdatedAt: now.Add(-2 * time.Minute)}
	if err := fresh.ValidateFreshness(now, 5*time.Minute); err != nil {
		t.Fatalf("expected fresh manifest to pass, got %v", err)
	}
	if err := fresh.ValidateFreshness(now, 30*time.Second); !errors.Is(err, ErrManifestStale) {
		t.Fatalf("expected ErrManifestStale, got %v", err)
	}
	if err := fresh.ValidateFreshness(now, 0); err != nil {
		t.Fatalf("expected zero maxAge to disable stale checks, got %v", err)
	}
}

func TestManifestClone(t *testing.T) {
	original := &Manifest{
		ServerID:     "clone-server",
		Version:      42,
		Description:  "original",
		UpdatedAt:    time.Now().UTC(),
		Capabilities: Capabilities{Chat: true, Voice: true},
	}
	clone := original.Clone()
	if clone == original {
		t.Fatalf("expected clone to not share pointer with original")
	}
	clone.Description = "cloned"
	if original.Description == clone.Description {
		t.Fatalf("modifying clone mutated the original")
	}
	if clone.Capabilities != original.Capabilities {
		t.Fatalf("capabilities were not preserved in clone")
	}
}

func TestServerStateMetadataHandling(t *testing.T) {
	manifest := &Manifest{ServerID: "state-server", Version: ManifestVersionV1, UpdatedAt: time.Now().UTC()}
	state := NewServerState(manifest)
	if state.LocalMetadata == nil {
		t.Fatalf("expected metadata map to be initialized")
	}
	if state.Manifest == manifest {
		t.Fatalf("expected NewServerState to clone the manifest")
	}
	state.Manifest.ServerID = "mutated"
	if manifest.ServerID != "state-server" {
		t.Fatalf("mutating server state manifest mutated original")
	}
	state.AddMetadata("feature", "enabled")
	if got := state.LocalMetadata["feature"]; got != "enabled" {
		t.Fatalf("metadata value mismatch, got %q", got)
	}
	state.LocalMetadata = nil
	state.AddMetadata("fresh", "value")
	if state.LocalMetadata["fresh"] != "value" {
		t.Fatalf("expected map to be reinitialized when nil")
	}
	var nilState *ServerState
	nilState.AddMetadata("nope", "ok") // should not panic
}

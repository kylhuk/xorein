package phase6

import (
	"errors"
	"testing"
	"time"
)

func TestManifestStorePublishResolve(t *testing.T) {
	store := NewManifestStore(time.Hour)
	base := &Manifest{
		ServerID:    "resolve-server",
		Version:     ManifestVersionV1,
		Description: "base",
		UpdatedAt:   time.Date(2025, time.December, 31, 23, 58, 0, 0, time.UTC),
	}
	if err := store.Publish(base); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	resolved, err := store.Resolve(base.ServerID)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if resolved == base {
		t.Fatalf("expected resolved manifest to be a clone")
	}
	resolved.Description = "mutated"
	second, err := store.Resolve(base.ServerID)
	if err != nil {
		t.Fatalf("second resolve failed: %v", err)
	}
	if second.Description != base.Description {
		t.Fatalf("expected original description preserved, got %q", second.Description)
	}
}

func TestManifestStorePublishConditionalUpdate(t *testing.T) {
	baseTime := time.Date(2025, time.November, 30, 8, 0, 0, 0, time.UTC)
	cases := []struct {
		name            string
		initial         *Manifest
		update          *Manifest
		expectedDesc    string
		expectedVersion int
	}{
		{
			name:            "newer version replaces",
			initial:         &Manifest{ServerID: "srv-newer-version", Version: ManifestVersionV1, Description: "original", UpdatedAt: baseTime},
			update:          &Manifest{ServerID: "srv-newer-version", Version: 2, Description: "new version", UpdatedAt: baseTime.Add(time.Minute)},
			expectedDesc:    "new version",
			expectedVersion: 2,
		},
		{
			name:            "reject older version",
			initial:         &Manifest{ServerID: "srv-reject-older", Version: 2, Description: "current", UpdatedAt: baseTime},
			update:          &Manifest{ServerID: "srv-reject-older", Version: ManifestVersionV1, Description: "older", UpdatedAt: baseTime.Add(-time.Minute)},
			expectedDesc:    "current",
			expectedVersion: 2,
		},
		{
			name:            "same version with newer timestamp",
			initial:         &Manifest{ServerID: "srv-same-version", Version: ManifestVersionV1, Description: "first", UpdatedAt: baseTime},
			update:          &Manifest{ServerID: "srv-same-version", Version: ManifestVersionV1, Description: "updated", UpdatedAt: baseTime.Add(time.Hour)},
			expectedDesc:    "updated",
			expectedVersion: ManifestVersionV1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewManifestStore(time.Hour)
			if err := store.Publish(tc.initial); err != nil {
				t.Fatalf("publish initial failed: %v", err)
			}
			if err := store.Publish(tc.update); err != nil {
				t.Fatalf("publish update failed: %v", err)
			}
			resolved, err := store.Resolve(tc.initial.ServerID)
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			if resolved.Description != tc.expectedDesc {
				t.Fatalf("description mismatch: want %q got %q", tc.expectedDesc, resolved.Description)
			}
			if resolved.Version != tc.expectedVersion {
				t.Fatalf("version mismatch: want %d got %d", tc.expectedVersion, resolved.Version)
			}
		})
	}
}

func TestManifestStorePublishRejectsInvalidManifest(t *testing.T) {
	store := NewManifestStore(time.Hour)
	cases := []struct {
		name    string
		m       *Manifest
		wantErr error
	}{
		{name: "nil manifest", m: nil, wantErr: ErrManifestNil},
		{name: "missing server id", m: &Manifest{Version: ManifestVersionV1, UpdatedAt: time.Now().UTC()}, wantErr: ErrManifestServerIDRequired},
		{name: "invalid version", m: &Manifest{ServerID: "srv", Version: 0, UpdatedAt: time.Now().UTC()}, wantErr: ErrManifestVersionInvalid},
		{name: "missing updated_at", m: &Manifest{ServerID: "srv", Version: ManifestVersionV1}, wantErr: ErrManifestUpdatedAtRequired},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.Publish(tc.m)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestManifestStoreTTLExpirationAndInvalidate(t *testing.T) {
	store := NewManifestStore(time.Hour)
	manifest := &Manifest{ServerID: "expiring-server", Version: 1, UpdatedAt: time.Now().UTC()}
	if err := store.Publish(manifest); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	store.mu.Lock()
	entry := store.entries[manifest.ServerID]
	entry.expiresAt = time.Now().UTC().Add(-time.Minute)
	store.mu.Unlock()
	if _, err := store.Resolve(manifest.ServerID); err != ErrManifestExpired {
		t.Fatalf("expected expired error, got %v", err)
	}
	if _, err := store.Resolve(manifest.ServerID); err != ErrManifestNotFound {
		t.Fatalf("expected not found after expiration, got %v", err)
	}
	if err := store.Publish(manifest); err != nil {
		t.Fatalf("publish before invalidate failed: %v", err)
	}
	store.Invalidate(manifest.ServerID)
	if _, err := store.Resolve(manifest.ServerID); err != ErrManifestNotFound {
		t.Fatalf("expected not found after invalidate, got %v", err)
	}
}

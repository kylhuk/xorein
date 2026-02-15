package phase5

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func deterministicKeyPair(t *testing.T, seed byte) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	seedBytes := bytes.Repeat([]byte{seed}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seedBytes)
	publicKey := make(ed25519.PublicKey, ed25519.PublicKeySize)
	copy(publicKey, privateKey[32:])
	return publicKey, privateKey
}

func identityHex(pk ed25519.PublicKey) string {
	return hex.EncodeToString(pk)
}

func profileFixture(identity, display, bio, avatar string, version int, updatedAt time.Time) *Profile {
	return &Profile{
		Identity:    identity,
		DisplayName: display,
		Bio:         bio,
		AvatarURL:   avatar,
		Version:     version,
		UpdatedAt:   updatedAt.UTC(),
	}
}

func mustSignProfile(t *testing.T, profile *Profile, private ed25519.PrivateKey) *SignedProfile {
	t.Helper()
	signed, err := SignProfile(profile, private)
	if err != nil {
		t.Fatalf("sign profile: %v", err)
	}
	return signed
}

func mustPublishAccepted(t *testing.T, store *ProfileStore, signed *SignedProfile) {
	t.Helper()
	accepted, err := store.Publish(signed)
	if err != nil {
		t.Fatalf("publish profile: %v", err)
	}
	if !accepted {
		t.Fatalf("expected publish to be accepted")
	}
}

func mustResolve(t *testing.T, store *ProfileStore, identity string) *Profile {
	t.Helper()
	resolved, err := store.Resolve(identity)
	if err != nil {
		t.Fatalf("resolve profile: %v", err)
	}
	return resolved
}

func assertProfileEqual(t *testing.T, got, want *Profile) {
	t.Helper()
	if got.Identity != want.Identity ||
		got.DisplayName != want.DisplayName ||
		got.Bio != want.Bio ||
		got.AvatarURL != want.AvatarURL ||
		got.Version != want.Version ||
		!got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Fatalf("profile mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestSignVerifyPublishResolve(t *testing.T) {
	pub, priv := deterministicKeyPair(t, 0x01)
	identity := identityHex(pub)
	updatedAt := time.Date(2024, time.March, 10, 12, 0, 0, 0, time.UTC)
	profile := profileFixture(identity, "alice", "intro", "https://example.com/a.png", 1, updatedAt)

	signed := mustSignProfile(t, profile, priv)
	if err := VerifySignedProfile(signed, pub); err != nil {
		t.Fatalf("verify signed profile: %v", err)
	}

	store := NewProfileStore()
	accepted, err := store.Publish(signed)
	if err != nil {
		t.Fatalf("publish signed profile: %v", err)
	}
	if !accepted {
		t.Fatalf("expected first publish to be accepted")
	}

	resolved := mustResolve(t, store, identity)
	assertProfileEqual(t, resolved, profile)

	cached := mustResolve(t, store, identity)
	if cached == resolved {
		t.Fatalf("resolve should return a cloned profile instance")
	}
	assertProfileEqual(t, cached, profile)
}

func TestPublishRejectsInvalidSignature(t *testing.T) {
	pub, priv := deterministicKeyPair(t, 0x02)
	identity := identityHex(pub)
	base := profileFixture(identity, "bob", "bio", "https://example.com/b.png", 1, time.Date(2024, time.March, 11, 9, 0, 0, 0, time.UTC))
	signed := mustSignProfile(t, base, priv)

	tampered := signed.Clone()
	tampered.Signature[0] ^= 0xFF

	if err := VerifySignedProfile(tampered, pub); !errors.Is(err, ErrProfileSignatureInvalid) {
		t.Fatalf("expected ErrProfileSignatureInvalid, got %v", err)
	}

	store := NewProfileStore()
	for attempt := 1; attempt <= 2; attempt++ {
		accepted, err := store.Publish(tampered)
		if !errors.Is(err, ErrProfileSignatureInvalid) {
			t.Fatalf("attempt %d: expected ErrProfileSignatureInvalid, got %v", attempt, err)
		}
		if accepted {
			t.Fatalf("attempt %d: expected publish to be rejected", attempt)
		}
	}

	if _, err := store.Resolve(identity); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound after rejected publish, got %v", err)
	}
}

func TestProfileStorePublishUpdateDeterminism(t *testing.T) {
	pub, priv := deterministicKeyPair(t, 0x03)
	identity := identityHex(pub)
	baseTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	baseProfile := profileFixture(identity, "Alice", "p5", "https://example.com/v2.png", 2, baseTime)
	baseSigned := mustSignProfile(t, baseProfile, priv)

	cases := []struct {
		name         string
		incoming     func() *Profile
		wantAccepted bool
		wantDisplay  string
		wantVersion  int
		wantUpdated  time.Time
	}{
		{
			name: "higher version wins",
			incoming: func() *Profile {
				return profileFixture(identity, "Alice v3", "p5", "https://example.com/v3.png", 3, baseTime.Add(2*time.Hour))
			},
			wantAccepted: true,
			wantDisplay:  "Alice v3",
			wantVersion:  3,
			wantUpdated:  baseTime.Add(2 * time.Hour),
		},
		{
			name: "lower version rejected",
			incoming: func() *Profile {
				return profileFixture(identity, "Alice downgraded", "p5", "https://example.com/v1.png", 1, baseTime.Add(3*time.Hour))
			},
			wantAccepted: false,
			wantDisplay:  baseProfile.DisplayName,
			wantVersion:  baseProfile.Version,
			wantUpdated:  baseProfile.UpdatedAt,
		},
		{
			name: "same version stale timestamp",
			incoming: func() *Profile {
				return profileFixture(identity, "Alice stale", "p5", "https://example.com/v2.png", 2, baseTime)
			},
			wantAccepted: false,
			wantDisplay:  baseProfile.DisplayName,
			wantVersion:  baseProfile.Version,
			wantUpdated:  baseProfile.UpdatedAt,
		},
		{
			name: "same version newer timestamp",
			incoming: func() *Profile {
				return profileFixture(identity, "Alice fresh", "p5", "https://example.com/v2.png", 2, baseTime.Add(time.Minute))
			},
			wantAccepted: true,
			wantDisplay:  "Alice fresh",
			wantVersion:  2,
			wantUpdated:  baseTime.Add(time.Minute),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewProfileStore()
			mustPublishAccepted(t, store, baseSigned)

			incomingSigned := mustSignProfile(t, tc.incoming(), priv)
			accepted, err := store.Publish(incomingSigned)
			if err != nil {
				t.Fatalf("publish update: %v", err)
			}
			if accepted != tc.wantAccepted {
				t.Fatalf("accepted mismatch: got %v want %v", accepted, tc.wantAccepted)
			}

			resolved := mustResolve(t, store, identity)
			if resolved.DisplayName != tc.wantDisplay || resolved.Version != tc.wantVersion || !resolved.UpdatedAt.Equal(tc.wantUpdated) {
				t.Fatalf("resolved profile mismatch for %s: got %+v", tc.name, resolved)
			}
		})
	}
}

func TestProfileStoreCacheInvalidationOnAcceptedUpdate(t *testing.T) {
	pub, priv := deterministicKeyPair(t, 0x04)
	identity := identityHex(pub)
	baseTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)

	store := NewProfileStore()
	initial := profileFixture(identity, "Carol", "p5", "https://example.com/c.png", 1, baseTime)
	mustPublishAccepted(t, store, mustSignProfile(t, initial, priv))

	first := mustResolve(t, store, identity)
	assertProfileEqual(t, first, initial)
	if _, ok := store.cache[identity]; !ok {
		t.Fatalf("expected cache entry after initial resolve")
	}

	updated := profileFixture(identity, "Carol refreshed", "p5", "https://example.com/c2.png", 2, baseTime.Add(time.Hour))
	accepted, err := store.Publish(mustSignProfile(t, updated, priv))
	if err != nil {
		t.Fatalf("publish updated profile: %v", err)
	}
	if !accepted {
		t.Fatalf("expected update to be accepted")
	}

	if _, ok := store.cache[identity]; ok {
		t.Fatalf("cache entry must be invalidated on accepted update")
	}

	resolved := mustResolve(t, store, identity)
	assertProfileEqual(t, resolved, updated)
	if cached, ok := store.cache[identity]; ok {
		assertProfileEqual(t, cached, updated)
	} else {
		t.Fatalf("expected cache repopulated after resolve")
	}
}

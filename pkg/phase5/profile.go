package phase5

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	maxProfileDisplayNameLen = 64
	maxProfileBioLen         = 280
	maxProfileAvatarURLLen   = 512
)

var (
	ErrProfileNil                    = errors.New("phase5: profile required")
	ErrProfileSignedNil              = errors.New("phase5: signed profile required")
	ErrProfileIdentityRequired       = errors.New("phase5: profile identity required")
	ErrProfileDisplayNameRequired    = errors.New("phase5: profile display_name required")
	ErrProfileDisplayNameTooLong     = fmt.Errorf("phase5: profile display_name exceeds %d chars", maxProfileDisplayNameLen)
	ErrProfileBioTooLong             = fmt.Errorf("phase5: profile bio exceeds %d chars", maxProfileBioLen)
	ErrProfileAvatarURLTooLong       = fmt.Errorf("phase5: profile avatar_url exceeds %d chars", maxProfileAvatarURLLen)
	ErrProfileVersionInvalid         = errors.New("phase5: profile version must be >= 1")
	ErrProfileUpdatedAtRequired      = errors.New("phase5: profile updated_at required")
	ErrProfileSignatureRequired      = errors.New("phase5: profile signature required")
	ErrProfileSignatureInvalid       = errors.New("phase5: profile signature invalid")
	ErrProfileSignerPrivateKeyLength = fmt.Errorf("phase5: expected %d byte private key", ed25519.PrivateKeySize)
	ErrProfileSignerPublicKeyLength  = fmt.Errorf("phase5: expected %d byte public key", ed25519.PublicKeySize)
	ErrProfileIdentityFormat         = errors.New("phase5: profile identity must be hex-encoded ed25519 public key")
	ErrProfileNotFound               = errors.New("phase5: profile not found")
)

// Profile captures the user-visible profile metadata published by an identity.
// The Identity field contains the canonical hex-encoded Ed25519 public key.
type Profile struct {
	Identity    string
	DisplayName string
	Bio         string
	AvatarURL   string
	Version     int
	UpdatedAt   time.Time
}

// ProfileFieldSensitivity classifies privacy impact for profile fields.
type ProfileFieldSensitivity string

const (
	ProfileSensitivityRestricted  ProfileFieldSensitivity = "restricted"
	ProfileSensitivityPersonal    ProfileFieldSensitivity = "personal"
	ProfileSensitivityPublic      ProfileFieldSensitivity = "public"
	ProfileSensitivityOperational ProfileFieldSensitivity = "operational"
)

var profileFieldSensitivity = map[string]ProfileFieldSensitivity{
	"identity":     ProfileSensitivityRestricted,
	"display_name": ProfileSensitivityPublic,
	"bio":          ProfileSensitivityPersonal,
	"avatar_url":   ProfileSensitivityPersonal,
	"version":      ProfileSensitivityOperational,
	"updated_at":   ProfileSensitivityOperational,
}

// ProfileFieldSensitivityMap returns a stable copy of profile field
// sensitivity classifications.
func ProfileFieldSensitivityMap() map[string]ProfileFieldSensitivity {
	out := make(map[string]ProfileFieldSensitivity, len(profileFieldSensitivity))
	for field, class := range profileFieldSensitivity {
		out[field] = class
	}
	return out
}

// DefaultProfileOptionalFields clears optional profile fields unless explicitly
// provided by the caller.
func DefaultProfileOptionalFields(p *Profile) {
	if p == nil {
		return
	}
	p.Bio = ""
	p.AvatarURL = ""
}

// RedactedProfileMetadata returns a deterministic, privacy-preserving metadata
// view that excludes personal optional fields.
func RedactedProfileMetadata(p *Profile) map[string]any {
	if p == nil {
		return map[string]any{}
	}
	return map[string]any{
		"identity":     p.Identity,
		"display_name": p.DisplayName,
		"version":      p.Version,
		"updated_at":   p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

// ValidateFields enforces deterministic profile field requirements.
func (p *Profile) ValidateFields() error {
	if p == nil {
		return ErrProfileNil
	}
	if p.Identity == "" {
		return ErrProfileIdentityRequired
	}
	if p.DisplayName == "" {
		return ErrProfileDisplayNameRequired
	}
	if len(p.DisplayName) > maxProfileDisplayNameLen {
		return ErrProfileDisplayNameTooLong
	}
	if len(p.Bio) > maxProfileBioLen {
		return ErrProfileBioTooLong
	}
	if len(p.AvatarURL) > maxProfileAvatarURLLen {
		return ErrProfileAvatarURLTooLong
	}
	if p.Version < 1 {
		return ErrProfileVersionInvalid
	}
	if p.UpdatedAt.IsZero() {
		return ErrProfileUpdatedAtRequired
	}
	return nil
}

// Clone returns a deep copy of the profile.
func (p *Profile) Clone() *Profile {
	if p == nil {
		return nil
	}
	clone := *p
	clone.UpdatedAt = p.UpdatedAt
	return &clone
}

// SignedProfile is the signed envelope used for publishing and resolving
// verified profiles.
type SignedProfile struct {
	Profile   *Profile
	Signature []byte
}

// Clone returns a deep copy of the signed profile.
func (sp *SignedProfile) Clone() *SignedProfile {
	if sp == nil {
		return nil
	}
	clone := &SignedProfile{Profile: sp.Profile.Clone()}
	clone.Signature = append(clone.Signature, sp.Signature...)
	return clone
}

// CanonicalPayload returns the deterministic bytes covered by the signature.
func (p *Profile) CanonicalPayload() ([]byte, error) {
	if err := p.ValidateFields(); err != nil {
		return nil, err
	}
	type canonicalProfile struct {
		Identity    string `json:"identity"`
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
		AvatarURL   string `json:"avatar_url"`
		Version     int    `json:"version"`
		UpdatedAt   string `json:"updated_at"`
	}
	payload := canonicalProfile{
		Identity:    p.Identity,
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		AvatarURL:   p.AvatarURL,
		Version:     p.Version,
		UpdatedAt:   p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	return json.Marshal(payload)
}

// SignProfile signs the canonical profile payload with Ed25519.
func SignProfile(profile *Profile, privateKey ed25519.PrivateKey) (*SignedProfile, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, ErrProfileSignerPrivateKeyLength
	}
	if err := profile.ValidateFields(); err != nil {
		return nil, err
	}
	payload, err := profile.CanonicalPayload()
	if err != nil {
		return nil, err
	}
	signature := ed25519.Sign(privateKey, payload)
	return &SignedProfile{Profile: profile.Clone(), Signature: signature}, nil
}

// VerifySignedProfile validates payload fields and signature bytes.
func VerifySignedProfile(sp *SignedProfile, publicKey ed25519.PublicKey) error {
	if sp == nil {
		return ErrProfileSignedNil
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return ErrProfileSignerPublicKeyLength
	}
	if len(sp.Signature) == 0 {
		return ErrProfileSignatureRequired
	}
	if err := sp.Profile.ValidateFields(); err != nil {
		return err
	}
	payload, err := sp.Profile.CanonicalPayload()
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, payload, sp.Signature) {
		return ErrProfileSignatureInvalid
	}
	return nil
}

func decodeIdentityPublicKey(identity string) (ed25519.PublicKey, error) {
	if identity == "" {
		return nil, ErrProfileIdentityRequired
	}
	raw, err := hex.DecodeString(identity)
	if err != nil || len(raw) != ed25519.PublicKeySize {
		return nil, ErrProfileIdentityFormat
	}
	pk := make(ed25519.PublicKey, ed25519.PublicKeySize)
	copy(pk, raw)
	return pk, nil
}

// ProfileStore publishes and resolves signed profiles with deterministic update
// behavior and cache invalidation on accepted updates.
type ProfileStore struct {
	mu        sync.RWMutex
	published map[string]*SignedProfile
	cache     map[string]*Profile
}

// NewProfileStore constructs an empty profile store.
func NewProfileStore() *ProfileStore {
	return &ProfileStore{
		published: map[string]*SignedProfile{},
		cache:     map[string]*Profile{},
	}
}

// Publish verifies and stores a signed profile. Update rules are deterministic:
// newer version wins; same version requires strictly newer UpdatedAt.
// Accepted updates invalidate the resolve cache entry for the profile identity.
func (s *ProfileStore) Publish(sp *SignedProfile) (bool, error) {
	if sp == nil {
		return false, ErrProfileSignedNil
	}
	publicKey, err := decodeIdentityPublicKey(sp.Profile.Identity)
	if err != nil {
		return false, err
	}
	if err := VerifySignedProfile(sp, publicKey); err != nil {
		return false, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.published[sp.Profile.Identity]
	if ok {
		prev := existing.Profile
		incoming := sp.Profile
		if incoming.Version < prev.Version {
			return false, nil
		}
		if incoming.Version == prev.Version && !incoming.UpdatedAt.After(prev.UpdatedAt) {
			return false, nil
		}
	}

	s.published[sp.Profile.Identity] = sp.Clone()
	delete(s.cache, sp.Profile.Identity)
	return true, nil
}

// Resolve verifies the stored signed profile before returning a cloned profile.
// Successful verification is cached until the next accepted publish update.
func (s *ProfileStore) Resolve(identity string) (*Profile, error) {
	s.mu.RLock()
	if cached, ok := s.cache[identity]; ok {
		clone := cached.Clone()
		s.mu.RUnlock()
		return clone, nil
	}
	sp, ok := s.published[identity]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrProfileNotFound
	}

	publicKey, err := decodeIdentityPublicKey(identity)
	if err != nil {
		return nil, err
	}
	if err := VerifySignedProfile(sp, publicKey); err != nil {
		return nil, err
	}

	resolved := sp.Profile.Clone()
	s.mu.Lock()
	s.cache[identity] = resolved.Clone()
	s.mu.Unlock()
	return resolved, nil
}

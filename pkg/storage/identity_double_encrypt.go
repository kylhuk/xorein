package storage

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

// DeriveIdentityKey derives the sub-encryption key for identity private key material per spec §9.1.
// storeKey is the 32-byte store key (from Argon2id).
// Returns a 32-byte key: HKDF-SHA-256(ikm=storeKey, info="xorein/storage/v1/identity-key", salt=none).
func DeriveIdentityKey(storeKey []byte) []byte {
	r := hkdf.New(sha256.New, storeKey, nil, []byte("xorein/storage/v1/identity-key"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		// hkdf.New with sha256 and 32-byte output never fails.
		panic("storage: DeriveIdentityKey: " + err.Error())
	}
	return key
}

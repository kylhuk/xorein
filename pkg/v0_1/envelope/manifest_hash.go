package envelope

import (
	"crypto/sha256"
	"encoding/base64"
)

// ManifestHash returns the 32-character base64url-no-pad prefix of the
// SHA-256 digest of the canonical manifest JSON per spec 02 §4.
//
//	hash = base64url_raw(SHA-256(canonical_manifest_JSON))[0:32]
func ManifestHash(canonicalManifestJSON []byte) string {
	sum := sha256.Sum256(canonicalManifestJSON)
	return base64.RawURLEncoding.EncodeToString(sum[:])[:32]
}

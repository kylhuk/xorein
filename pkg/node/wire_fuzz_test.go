package node

import (
	"testing"
	"time"
)

func FuzzParseDeeplink(f *testing.F) {
	identity, err := GenerateIdentity("fuzz")
	if err != nil {
		f.Fatalf("GenerateIdentity() error = %v", err)
	}
	invite := Invite{ServerID: "server-seed", OwnerPeerID: identity.PeerID, OwnerPublicKey: identity.PublicKey, ServerAddrs: []string{"127.0.0.1:1"}, ManifestHash: "seed", ExpiresAt: identity.CreatedAt.Add(24 * 365 * time.Hour)}
	if err := invite.Sign(identity); err != nil {
		f.Fatalf("invite.Sign() error = %v", err)
	}
	deeplink, err := invite.Deeplink()
	if err != nil {
		f.Fatalf("invite.Deeplink() error = %v", err)
	}
	f.Add(deeplink)
	f.Add("aether://join/server-seed?invite=broken")
	f.Add("https://example.invalid")
	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = ParseDeeplink(raw)
	})
}

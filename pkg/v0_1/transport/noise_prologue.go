package transport

import (
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
)

// aetherNoisePrologue is the mandatory Noise XX prologue per spec 30 §1.2.
// Both peers MUST use this value; mismatched prologues cause handshake failure.
var aetherNoisePrologue = []byte("/aether/noise/1.0")

// NewNoiseWithPrologue is a libp2p.Security-compatible constructor that wraps
// noise.New with the Xorein-specific prologue injected via WithSessionOptions.
func NewNoiseWithPrologue(id libp2pprotocol.ID, privKey libp2pcrypto.PrivKey, muxers []tptu.StreamMuxer) (*noise.SessionTransport, error) {
	t, err := noise.New(id, privKey, muxers)
	if err != nil {
		return nil, err
	}
	return t.WithSessionOptions(noise.Prologue(aetherNoisePrologue))
}

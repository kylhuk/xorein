package nat

import (
	"fmt"
	"time"

	relay "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
)

// RelayService wraps the Circuit Relay v2 service with spec 32 §2.5 quotas.
type RelayService struct {
	svc *relay.Relay
}

// NewRelayService starts a Circuit Relay v2 service on the given host.
// Intended for nodes running with role="relay".
func NewRelayService(h libp2phost.Host) (*RelayService, error) {
	rc := relay.Resources{
		Limit: &relay.RelayLimit{
			Duration: 2 * time.Minute,
			Data:     2 << 20, // 2 MiB per connection per spec 32 §2.5
		},
		ReservationTTL:  24 * time.Hour,
		MaxReservations: 512,
		MaxCircuits:     64,
		BufferSize:      4096,
	}

	svc, err := relay.New(h, relay.WithResources(rc))
	if err != nil {
		return nil, fmt.Errorf("relay service: %w", err)
	}
	return &RelayService{svc: svc}, nil
}

// Close shuts down the relay service.
func (rs *RelayService) Close() error {
	return rs.svc.Close()
}

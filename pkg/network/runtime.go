package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/aether/code_aether/pkg/protocol"
	libp2p "github.com/libp2p/go-libp2p"
	lphost "github.com/libp2p/go-libp2p/core/host"
	lpnetwork "github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
)

type Mode string

const (
	ModeClient    Mode = "client"
	ModeRelay     Mode = "relay"
	ModeBootstrap Mode = "bootstrap"
	ModeArchivist Mode = "archivist"
)

func (m Mode) Valid() bool {
	switch m {
	case ModeClient, ModeRelay, ModeBootstrap, ModeArchivist:
		return true
	default:
		return false
	}
}

type Config struct {
	Mode       Mode
	ListenAddr string
}

type Runtime interface {
	Start(ctx context.Context) error
	Close() error
	ListenAddress() string
}

type P2PRuntime struct {
	cfg Config

	handlerMu sync.RWMutex
	handler   Handler

	mu     sync.RWMutex
	host   lphost.Host
	once   sync.Once
	closed chan struct{}
}

func NewP2PRuntime(cfg Config) (*P2PRuntime, error) {
	if !cfg.Mode.Valid() {
		return nil, fmt.Errorf("invalid runtime mode %q", cfg.Mode)
	}
	if strings.TrimSpace(cfg.ListenAddr) == "" {
		cfg.ListenAddr = "127.0.0.1:0"
	}
	if _, err := listenMultiaddrString(cfg.ListenAddr); err != nil {
		return nil, err
	}
	return &P2PRuntime{cfg: cfg, closed: make(chan struct{})}, nil
}

func (r *P2PRuntime) SetHandler(handler Handler) {
	r.handlerMu.Lock()
	r.handler = handler
	r.handlerMu.Unlock()
}

func (r *P2PRuntime) Start(ctx context.Context) error {
	if r.handlerSnapshot() == nil {
		return errors.New("peer transport handler is required")
	}
	listenAddr, err := listenMultiaddrString(r.cfg.ListenAddr)
	if err != nil {
		return err
	}
	h, err := newPeerHost(libp2p.ListenAddrStrings(listenAddr))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	for _, protocolID := range peerTransportProtocols() {
		h.SetStreamHandler(protocolID, r.handlePeerTransport)
	}
	r.mu.Lock()
	r.host = h
	r.mu.Unlock()
	go func() {
		<-ctx.Done()
		_ = r.Close()
	}()
	return nil
}

func (r *P2PRuntime) Close() error {
	var retErr error
	r.once.Do(func() {
		close(r.closed)
		r.mu.RLock()
		h := r.host
		r.mu.RUnlock()
		if h != nil {
			retErr = h.Close()
		}
	})
	return retErr
}

func (r *P2PRuntime) ListenAddress() string {
	r.mu.RLock()
	h := r.host
	r.mu.RUnlock()
	if h != nil {
		return hostListenAddress(h)
	}
	return strings.TrimSpace(r.cfg.ListenAddr)
}

func (r *P2PRuntime) handlerSnapshot() Handler {
	r.handlerMu.RLock()
	defer r.handlerMu.RUnlock()
	return r.handler
}

func (r *P2PRuntime) handlePeerTransport(stream lpnetwork.Stream) {
	defer stream.Close()
	handler := r.handlerSnapshot()
	if handler == nil {
		r.writeTransportResponse(stream, responseEnvelope{TransportError: Error{Code: "transport_unavailable", Message: "peer transport handler is unavailable"}})
		return
	}
	requestBytes, err := readStreamPayload(stream, maxTransportFrameSize)
	if err != nil {
		r.writeTransportResponse(stream, responseEnvelope{TransportError: Error{Code: "invalid_request", Message: err.Error()}})
		return
	}
	transportReq, err := unmarshalRequest(requestBytes)
	if err != nil {
		r.writeTransportResponse(stream, responseEnvelope{TransportError: Error{Code: "invalid_request", Message: err.Error()}})
		return
	}
	negotiation, err := protocol.NegotiatePeerTransport(
		[]string{string(stream.Protocol())},
		transportReq.AdvertisedCapabilities,
		mergeCapabilities(transportReq.RequiredCapabilities, requiredCapabilities(transportReq.Operation)),
	)
	if err != nil {
		var negotiationErr *protocol.NegotiationError
		if errors.As(err, &negotiationErr) {
			r.writeTransportResponse(stream, responseEnvelope{TransportError: Error{
				Code:             string(negotiationErr.Code),
				Message:          negotiationErr.Message,
				OfferedProtocols: append([]string(nil), negotiationErr.OfferedProtocols...),
				MissingRequired:  append([]string(nil), negotiationErr.MissingRequired...),
			}})
			return
		}
		r.writeTransportResponse(stream, responseEnvelope{TransportError: Error{Code: "negotiation_failed", Message: err.Error()}})
		return
	}
	payload, transportErr := handler.HandlePeerOperation(context.Background(), transportReq.Operation, transportReq.Payload)
	if transportErr != nil {
		r.writeTransportResponse(stream, responseEnvelope{
			NegotiatedProtocol:   negotiation.Protocol.String(),
			AcceptedCapabilities: protocol.FeatureFlagStrings(negotiation.CapabilityResult.Accepted),
			IgnoredCapabilities:  append([]string(nil), negotiation.CapabilityResult.IgnoredRemote...),
			TransportError:       *transportErr,
		})
		return
	}
	r.writeTransportResponse(stream, responseEnvelope{
		NegotiatedProtocol:   negotiation.Protocol.String(),
		AcceptedCapabilities: protocol.FeatureFlagStrings(negotiation.CapabilityResult.Accepted),
		IgnoredCapabilities:  append([]string(nil), negotiation.CapabilityResult.IgnoredRemote...),
		Payload:              payload,
	})
}

func (r *P2PRuntime) writeTransportResponse(stream lpnetwork.Stream, resp responseEnvelope) {
	raw, err := marshalResponse(resp)
	if err != nil {
		_ = stream.Reset()
		return
	}
	if err := writeStreamPayload(stream, raw); err != nil {
		_ = stream.Reset()
		return
	}
	_ = stream.CloseWrite()
}

func listenMultiaddrString(address string) (string, error) {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return "", errors.New("listen address is required")
	}
	host, port, err := net.SplitHostPort(trimmed)
	if err != nil {
		return "", fmt.Errorf("listen address %q must be host:port: %w", trimmed, err)
	}
	if strings.TrimSpace(port) == "" {
		return "", fmt.Errorf("listen address %q is missing a port", trimmed)
	}
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("listen address %q must use an IP literal or localhost", trimmed)
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return fmt.Sprintf("/ip4/%s/tcp/%s", ipv4.String(), port), nil
	}
	return fmt.Sprintf("/ip6/%s/tcp/%s", ip.String(), port), nil
}

func hostListenAddress(h lphost.Host) string {
	addresses := make([]string, 0, len(h.Addrs()))
	for _, addr := range h.Addrs() {
		if _, err := addr.ValueForProtocol(ma.P_TCP); err != nil {
			continue
		}
		addresses = append(addresses, addr.String()+"/p2p/"+h.ID().String())
	}
	if len(addresses) == 0 {
		for _, addr := range h.Addrs() {
			addresses = append(addresses, addr.String()+"/p2p/"+h.ID().String())
		}
	}
	sort.Strings(addresses)
	if len(addresses) == 0 {
		return ""
	}
	return addresses[0]
}

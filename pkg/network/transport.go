package network

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/aether/code_aether/pkg/protocol"
	libp2p "github.com/libp2p/go-libp2p"
	lphost "github.com/libp2p/go-libp2p/core/host"
	lppeer "github.com/libp2p/go-libp2p/core/peer"
	lpprotocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type Operation string

const (
	OperationPeerInfo          Operation = "peer.info"
	OperationPeerExchange      Operation = "peer.exchange"
	OperationBootstrapRegister Operation = "peer.bootstrap.register"
	OperationBootstrapPeers    Operation = "peer.bootstrap.peers"
	OperationManifestPublish   Operation = "peer.manifest.publish"
	OperationManifestFetch     Operation = "peer.manifest.fetch"
	OperationPreviewFetch      Operation = "peer.preview.fetch"
	OperationJoin              Operation = "peer.join"
	OperationDeliver           Operation = "peer.deliver"
	OperationRelayStore        Operation = "peer.relay.store"
	OperationRelayDrain        Operation = "peer.relay.drain"

	maxTransportFrameSize = 8 << 20
)

func (o Operation) Valid() bool {
	switch o {
	case OperationPeerInfo,
		OperationPeerExchange,
		OperationBootstrapRegister,
		OperationBootstrapPeers,
		OperationManifestPublish,
		OperationManifestFetch,
		OperationPreviewFetch,
		OperationJoin,
		OperationDeliver,
		OperationRelayStore,
		OperationRelayDrain:
		return true
	default:
		return false
	}
}

type Request struct {
	AdvertisedCapabilities []string  `json:"advertised_capabilities,omitempty"`
	RequiredCapabilities   []string  `json:"required_capabilities,omitempty"`
	Operation              Operation `json:"operation"`
	Payload                []byte    `json:"payload,omitempty"`
}

type Response struct {
	NegotiatedProtocol   string   `json:"negotiated_protocol"`
	AcceptedCapabilities []string `json:"accepted_capabilities,omitempty"`
	IgnoredCapabilities  []string `json:"ignored_capabilities,omitempty"`
	Payload              []byte   `json:"payload,omitempty"`
}

type Error struct {
	Code             string   `json:"code"`
	Message          string   `json:"message"`
	OfferedProtocols []string `json:"offered_protocols,omitempty"`
	MissingRequired  []string `json:"missing_required,omitempty"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Code) == "" {
		return e.Message
	}
	if strings.TrimSpace(e.Message) == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type Handler interface {
	HandlePeerOperation(ctx context.Context, operation Operation, payload []byte) ([]byte, *Error)
}

type ResponseMeta struct {
	NegotiatedProtocol   string
	AcceptedCapabilities []string
	IgnoredCapabilities  []string
}

type Client struct {
	timeout time.Duration
}

func NewClient(timeout time.Duration) Client {
	if timeout <= 0 {
		timeout = 1200 * time.Millisecond
	}
	return Client{timeout: timeout}
}

func (c Client) Call(ctx context.Context, address string, operation Operation, payload any, out any) (ResponseMeta, error) {
	req, err := NewRequest(operation, payload)
	if err != nil {
		return ResponseMeta{}, err
	}
	return c.Do(ctx, address, req, out)
}

func (c Client) Do(ctx context.Context, address string, req Request, out any) (ResponseMeta, error) {
	if !req.Operation.Valid() {
		return ResponseMeta{}, fmt.Errorf("unsupported peer operation %q", req.Operation)
	}
	normalized, err := NormalizePeerAddress(address)
	if err != nil {
		return ResponseMeta{}, err
	}
	peerInfo, err := peerInfoFromAddress(normalized)
	if err != nil {
		return ResponseMeta{}, err
	}
	h, err := newPeerHost(libp2p.NoListenAddrs)
	if err != nil {
		return ResponseMeta{}, err
	}
	defer h.Close()
	if err := h.Connect(ctx, peerInfo); err != nil {
		return ResponseMeta{}, err
	}
	stream, err := h.NewStream(ctx, peerInfo.ID, peerTransportProtocols()...)
	if err != nil {
		return ResponseMeta{}, err
	}
	defer stream.Close()
	if c.timeout > 0 {
		_ = stream.SetDeadline(time.Now().Add(c.timeout))
	}
	raw, err := marshalRequest(req)
	if err != nil {
		return ResponseMeta{}, err
	}
	if err := writeStreamPayload(stream, raw); err != nil {
		_ = stream.Reset()
		return ResponseMeta{}, err
	}
	if err := stream.CloseWrite(); err != nil {
		_ = stream.Reset()
		return ResponseMeta{}, err
	}
	responseBytes, err := readStreamPayload(stream, maxTransportFrameSize)
	if err != nil {
		return ResponseMeta{}, err
	}
	transportResp, err := unmarshalResponse(responseBytes)
	if err != nil {
		return ResponseMeta{}, err
	}
	if strings.TrimSpace(transportResp.TransportError.Code) != "" {
		return ResponseMeta{}, &transportResp.TransportError
	}
	if out != nil && len(transportResp.Payload) > 0 {
		if err := UnmarshalPayload(transportResp.Payload, out); err != nil {
			return ResponseMeta{}, err
		}
	}
	return ResponseMeta{
		NegotiatedProtocol:   transportResp.NegotiatedProtocol,
		AcceptedCapabilities: append([]string(nil), transportResp.AcceptedCapabilities...),
		IgnoredCapabilities:  append([]string(nil), transportResp.IgnoredCapabilities...),
	}, nil
}

func NewRequest(operation Operation, payload any) (Request, error) {
	raw, err := MarshalPayload(payload)
	if err != nil {
		return Request{}, err
	}
	return Request{
		AdvertisedCapabilities: protocol.FeatureFlagStrings(protocol.DefaultPeerTransportFeatureFlags()),
		RequiredCapabilities:   requiredCapabilities(operation),
		Operation:              operation,
		Payload:                raw,
	}, nil
}

func MarshalPayload(payload any) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	value, err := structpb.NewValue(decoded)
	if err != nil {
		return nil, err
	}
	encoded, err := proto.Marshal(value)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func UnmarshalPayload(raw []byte, out any) error {
	if out == nil || len(raw) == 0 {
		return nil
	}
	var value structpb.Value
	if err := proto.Unmarshal(raw, &value); err != nil {
		return err
	}
	jsonRaw, err := protojson.Marshal(&value)
	if err != nil {
		return err
	}
	return decodeJSON(strings.NewReader(string(jsonRaw)), out)
}

func NormalizePeerAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", errors.New("peer address is required")
	}
	if strings.HasPrefix(address, "/") {
		return normalizeMultiaddrAddress(address)
	}
	return normalizeLegacyPeerAddress(address)
}

func normalizeLegacyPeerAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	address = strings.TrimPrefix(address, "http://")
	address = strings.TrimPrefix(address, "https://")
	if idx := strings.IndexAny(address, "/?#"); idx >= 0 {
		address = address[:idx]
	}
	address = strings.TrimSpace(address)
	if address == "" {
		return "", errors.New("peer address is required")
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("peer address %q must be host:port: %w", address, err)
	}
	if strings.TrimSpace(port) == "" {
		return "", fmt.Errorf("peer address %q is missing a port", address)
	}
	canonicalHost, networkProto, err := normalizeTransportHost(strings.TrimSpace(host), false)
	if err != nil {
		return "", fmt.Errorf("peer address %q: %w", address, err)
	}
	return fmt.Sprintf("/%s/%s/tcp/%s", networkProto, canonicalHost, port), nil
}

func normalizeMultiaddrAddress(address string) (string, error) {
	maddr, err := ma.NewMultiaddr(strings.TrimSpace(address))
	if err != nil {
		return "", err
	}
	port, err := maddr.ValueForProtocol(ma.P_TCP)
	if err != nil || strings.TrimSpace(port) == "" {
		return "", fmt.Errorf("peer address %q must include /tcp/<port>", address)
	}
	hostValue, networkProto, err := multiaddrHostComponent(maddr)
	if err != nil {
		return "", err
	}
	canonicalHost, networkProto, err := normalizeTransportHost(hostValue, networkProto == "ip6")
	if err != nil {
		return "", err
	}
	canonical := fmt.Sprintf("/%s/%s/tcp/%s", networkProto, canonicalHost, strings.TrimSpace(port))
	if peerID, err := maddr.ValueForProtocol(ma.P_P2P); err == nil && strings.TrimSpace(peerID) != "" {
		if _, err := lppeer.Decode(peerID); err != nil {
			return "", fmt.Errorf("invalid /p2p peer id: %w", err)
		}
		canonical += "/p2p/" + peerID
	}
	return canonical, nil
}

func multiaddrHostComponent(maddr ma.Multiaddr) (string, string, error) {
	if value, err := maddr.ValueForProtocol(ma.P_IP4); err == nil && strings.TrimSpace(value) != "" {
		return value, "ip4", nil
	}
	if value, err := maddr.ValueForProtocol(ma.P_IP6); err == nil && strings.TrimSpace(value) != "" {
		return value, "ip6", nil
	}
	if value, err := maddr.ValueForProtocol(ma.P_DNS4); err == nil && strings.TrimSpace(value) != "" {
		return value, "ip4", nil
	}
	if value, err := maddr.ValueForProtocol(ma.P_DNS6); err == nil && strings.TrimSpace(value) != "" {
		return value, "ip6", nil
	}
	if value, err := maddr.ValueForProtocol(ma.P_DNS); err == nil && strings.TrimSpace(value) != "" {
		return value, "ip4", nil
	}
	return "", "", fmt.Errorf("peer address %q must use /ip4, /ip6, or localhost DNS", maddr.String())
}

func normalizeTransportHost(host string, preferIPv6 bool) (string, string, error) {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		if preferIPv6 {
			return "::1", "ip6", nil
		}
		return "127.0.0.1", "ip4", nil
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		return "", "", errors.New("must use an IP literal or localhost")
	}
	if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		if !ip.IsLoopback() {
			return "", "", fmt.Errorf("peer address %q is not allowed", host)
		}
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4.String(), "ip4", nil
	}
	return ip.String(), "ip6", nil
}

func peerInfoFromAddress(address string) (lppeer.AddrInfo, error) {
	peerInfo, err := lppeer.AddrInfoFromString(address)
	if err == nil {
		return *peerInfo, nil
	}
	if strings.Contains(address, "/p2p/") {
		return lppeer.AddrInfo{}, err
	}
	return lppeer.AddrInfo{}, fmt.Errorf("peer address %q must include /p2p/<peer-id>", address)
}

func decodeJSON(r io.Reader, out any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func requiredCapabilities(operation Operation) []string {
	capabilitySet := map[string]struct{}{
		string(protocol.FeaturePeerTransport): {},
	}
	switch operation {
	case OperationPeerInfo:
		capabilitySet[string(protocol.FeaturePeerMetadata)] = struct{}{}
	case OperationPeerExchange:
		capabilitySet[string(protocol.FeaturePeerMetadata)] = struct{}{}
	case OperationBootstrapRegister, OperationBootstrapPeers:
		capabilitySet[string(protocol.FeaturePeerBootstrap)] = struct{}{}
	case OperationManifestPublish, OperationManifestFetch, OperationPreviewFetch:
		capabilitySet[string(protocol.FeaturePeerManifest)] = struct{}{}
	case OperationJoin:
		capabilitySet[string(protocol.FeaturePeerJoin)] = struct{}{}
	case OperationDeliver:
		capabilitySet[string(protocol.FeaturePeerDelivery)] = struct{}{}
	case OperationRelayStore, OperationRelayDrain:
		capabilitySet[string(protocol.FeaturePeerRelay)] = struct{}{}
	}
	out := make([]string, 0, len(capabilitySet))
	for capability := range capabilitySet {
		out = append(out, capability)
	}
	sort.Strings(out)
	return out
}

func mergeCapabilities(left, right []string) []string {
	set := make(map[string]struct{}, len(left)+len(right))
	for _, values := range [][]string{left, right} {
		for _, value := range values {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			set[trimmed] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

type responseEnvelope struct {
	NegotiatedProtocol   string
	AcceptedCapabilities []string
	IgnoredCapabilities  []string
	Payload              []byte
	TransportError       Error
}

func marshalRequest(req Request) ([]byte, error) {
	if !req.Operation.Valid() {
		return nil, fmt.Errorf("unsupported peer operation %q", req.Operation)
	}
	msg, err := structpb.NewStruct(map[string]any{
		"advertised_capabilities": stringSliceToAny(cloneTrimmedStrings(req.AdvertisedCapabilities)),
		"required_capabilities":   stringSliceToAny(cloneTrimmedStrings(req.RequiredCapabilities)),
		"operation":               string(req.Operation),
		"payload":                 base64.StdEncoding.EncodeToString(req.Payload),
	})
	if err != nil {
		return nil, err
	}
	return proto.MarshalOptions{Deterministic: true}.Marshal(msg)
}

func unmarshalRequest(raw []byte) (Request, error) {
	var envelope structpb.Struct
	if err := proto.Unmarshal(raw, &envelope); err != nil {
		return Request{}, err
	}
	fields := envelope.GetFields()
	req := Request{
		AdvertisedCapabilities: stringSliceFromValue(fields["advertised_capabilities"]),
		RequiredCapabilities:   stringSliceFromValue(fields["required_capabilities"]),
		Operation:              Operation(strings.TrimSpace(stringValue(fields["operation"]))),
	}
	if payload := stringValue(fields["payload"]); strings.TrimSpace(payload) != "" {
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return Request{}, fmt.Errorf("decode payload: %w", err)
		}
		req.Payload = decoded
	}
	if !req.Operation.Valid() {
		return Request{}, fmt.Errorf("unsupported peer operation %q", req.Operation)
	}
	return req, nil
}

func marshalResponse(resp responseEnvelope) ([]byte, error) {
	fields := map[string]any{
		"negotiated_protocol":   strings.TrimSpace(resp.NegotiatedProtocol),
		"accepted_capabilities": stringSliceToAny(cloneTrimmedStrings(resp.AcceptedCapabilities)),
		"ignored_capabilities":  stringSliceToAny(cloneTrimmedStrings(resp.IgnoredCapabilities)),
		"payload":               base64.StdEncoding.EncodeToString(resp.Payload),
	}
	if trimmed := strings.TrimSpace(resp.TransportError.Code); trimmed != "" || strings.TrimSpace(resp.TransportError.Message) != "" || len(resp.TransportError.OfferedProtocols) > 0 || len(resp.TransportError.MissingRequired) > 0 {
		fields["transport_error"] = map[string]any{
			"code":              strings.TrimSpace(resp.TransportError.Code),
			"message":           strings.TrimSpace(resp.TransportError.Message),
			"offered_protocols": stringSliceToAny(cloneTrimmedStrings(resp.TransportError.OfferedProtocols)),
			"missing_required":  stringSliceToAny(cloneTrimmedStrings(resp.TransportError.MissingRequired)),
		}
	}
	msg, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, err
	}
	return proto.MarshalOptions{Deterministic: true}.Marshal(msg)
}

func unmarshalResponse(raw []byte) (responseEnvelope, error) {
	var msg structpb.Struct
	if err := proto.Unmarshal(raw, &msg); err != nil {
		return responseEnvelope{}, err
	}
	fields := msg.GetFields()
	resp := responseEnvelope{
		NegotiatedProtocol:   strings.TrimSpace(stringValue(fields["negotiated_protocol"])),
		AcceptedCapabilities: stringSliceFromValue(fields["accepted_capabilities"]),
		IgnoredCapabilities:  stringSliceFromValue(fields["ignored_capabilities"]),
	}
	if payload := stringValue(fields["payload"]); strings.TrimSpace(payload) != "" {
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return responseEnvelope{}, fmt.Errorf("decode payload: %w", err)
		}
		resp.Payload = decoded
	}
	if transportError := fields["transport_error"]; transportError != nil {
		errFields := transportError.GetStructValue().GetFields()
		resp.TransportError = Error{
			Code:             strings.TrimSpace(stringValue(errFields["code"])),
			Message:          strings.TrimSpace(stringValue(errFields["message"])),
			OfferedProtocols: stringSliceFromValue(errFields["offered_protocols"]),
			MissingRequired:  stringSliceFromValue(errFields["missing_required"]),
		}
	}
	return resp, nil
}

func stringSliceToAny(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func stringValue(value *structpb.Value) string {
	if value == nil {
		return ""
	}
	return value.GetStringValue()
}

func stringSliceFromValue(value *structpb.Value) []string {
	if value == nil {
		return nil
	}
	list := value.GetListValue()
	if list == nil {
		return nil
	}
	out := make([]string, 0, len(list.GetValues()))
	for _, item := range list.GetValues() {
		trimmed := strings.TrimSpace(item.GetStringValue())
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func cloneTrimmedStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func readStreamPayload(r io.Reader, maxSize int) ([]byte, error) {
	limited := io.LimitReader(r, int64(maxSize)+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(payload) > maxSize {
		return nil, fmt.Errorf("transport frame too large: %d", len(payload))
	}
	return payload, nil
}

func writeStreamPayload(w io.Writer, payload []byte) error {
	if len(payload) == 0 {
		return nil
	}
	_, err := w.Write(payload)
	return err
}

func peerTransportProtocols() []lpprotocol.ID {
	protocols := protocol.CanonicalProtocolStrings(protocol.FamilyPeer)
	out := make([]lpprotocol.ID, 0, len(protocols))
	for _, value := range protocols {
		out = append(out, lpprotocol.ID(value))
	}
	return out
}

func newPeerHost(options ...libp2p.Option) (lphost.Host, error) {
	base := []libp2p.Option{
		libp2p.Security(noise.ID, noise.New),
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
	}
	base = append(base, options...)
	return libp2p.New(base...)
}

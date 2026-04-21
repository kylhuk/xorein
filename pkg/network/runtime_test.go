package network

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	lppeer "github.com/libp2p/go-libp2p/core/peer"
	lpprotocol "github.com/libp2p/go-libp2p/core/protocol"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type testHandler struct{}

func (testHandler) HandlePeerOperation(_ context.Context, operation Operation, payload []byte) ([]byte, *Error) {
	if operation != OperationPeerInfo {
		return nil, &Error{Code: "unsupported_operation", Message: "unsupported peer operation"}
	}
	return append([]byte(nil), payload...), nil
}

func TestNewP2PRuntimeRejectsInvalidMode(t *testing.T) {
	if _, err := NewP2PRuntime(Config{Mode: Mode("invalid")}); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestP2PRuntimeStartPublishesLibp2pListenAddress(t *testing.T) {
	runtime, err := NewP2PRuntime(Config{Mode: ModeRelay, ListenAddr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("NewP2PRuntime() error = %v", err)
	}
	runtime.SetHandler(testHandler{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = runtime.Close() }()

	address := runtime.ListenAddress()
	if !strings.Contains(address, "/tcp/") || !strings.Contains(address, "/p2p/") {
		t.Fatalf("ListenAddress() = %q, want libp2p multiaddr with peer id", address)
	}
}

func TestP2PRuntimeNegotiatesPeerTransport(t *testing.T) {
	runtime, err := NewP2PRuntime(Config{Mode: ModeClient, ListenAddr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("NewP2PRuntime() error = %v", err)
	}
	runtime.SetHandler(testHandler{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = runtime.Close() }()

	client := NewClient(time.Second)
	var payload map[string]string
	meta, err := client.Call(ctx, runtime.ListenAddress(), OperationPeerInfo, map[string]string{"hello": "world"}, &payload)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if meta.NegotiatedProtocol != "/aether/peer/0.1.0" {
		t.Fatalf("negotiated protocol = %q want %q", meta.NegotiatedProtocol, "/aether/peer/0.1.0")
	}
	if payload["hello"] != "world" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestP2PRuntimeRejectsUnsupportedCapability(t *testing.T) {
	runtime, err := NewP2PRuntime(Config{Mode: ModeClient, ListenAddr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("NewP2PRuntime() error = %v", err)
	}
	runtime.SetHandler(testHandler{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = runtime.Close() }()

	request, err := NewRequest(OperationPeerInfo, map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	request.RequiredCapabilities = append(request.RequiredCapabilities, "cap.peer.experimental")

	client := NewClient(time.Second)
	_, err = client.Do(ctx, runtime.ListenAddress(), request, nil)
	if err == nil {
		t.Fatal("expected unsupported capability error")
	}
	var transportErr *Error
	if !errors.As(err, &transportErr) {
		t.Fatalf("error type = %T want *Error", err)
	}
	if transportErr.Code != "unsupported-capability" {
		t.Fatalf("code = %q want unsupported-capability", transportErr.Code)
	}
	if len(transportErr.MissingRequired) != 1 || transportErr.MissingRequired[0] != "cap.peer.experimental" {
		t.Fatalf("missing required = %#v", transportErr.MissingRequired)
	}
}

func TestP2PRuntimeRejectsUnsupportedProtocolVersion(t *testing.T) {
	runtime, err := NewP2PRuntime(Config{Mode: ModeClient, ListenAddr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("NewP2PRuntime() error = %v", err)
	}
	runtime.SetHandler(testHandler{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = runtime.Close() }()

	peerInfo, err := lppeer.AddrInfoFromString(runtime.ListenAddress())
	if err != nil {
		t.Fatalf("AddrInfoFromString() error = %v", err)
	}
	h, err := newPeerHost(libp2p.NoListenAddrs)
	if err != nil {
		t.Fatalf("newPeerHost() error = %v", err)
	}
	defer h.Close()
	if err := h.Connect(ctx, *peerInfo); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if _, err := h.NewStream(ctx, peerInfo.ID, lpprotocol.ID("/aether/peer/1.0.0")); err == nil {
		t.Fatal("expected unsupported protocol stream error")
	}
}

func TestMarshalPayloadUsesProtobufBinary(t *testing.T) {
	raw, err := MarshalPayload(map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("MarshalPayload() error = %v", err)
	}
	var out map[string]string
	if err := UnmarshalPayload(raw, &out); err != nil {
		t.Fatalf("UnmarshalPayload() error = %v", err)
	}
	if out["hello"] != "world" {
		t.Fatalf("payload = %#v", out)
	}
}

func TestTransportRequestUsesProtobufStructEnvelope(t *testing.T) {
	req := Request{
		AdvertisedCapabilities: []string{"cap.peer.transport", " cap.peer.metadata "},
		RequiredCapabilities:   []string{"cap.peer.transport"},
		Operation:              OperationPeerInfo,
		Payload:                []byte{0x01, 0x02, 0x03},
	}
	raw, err := marshalRequest(req)
	if err != nil {
		t.Fatalf("marshalRequest() error = %v", err)
	}
	var envelope structpb.Struct
	if err := proto.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("proto.Unmarshal() error = %v", err)
	}
	fields := envelope.GetFields()
	if got, want := stringValue(fields["operation"]), string(OperationPeerInfo); got != want {
		t.Fatalf("operation = %q want %q", got, want)
	}
	if got, want := stringSliceFromValue(fields["advertised_capabilities"]), []string{"cap.peer.transport", "cap.peer.metadata"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("advertised capabilities = %#v want %#v", got, want)
	}
	if got, want := stringValue(fields["payload"]), base64.StdEncoding.EncodeToString(req.Payload); got != want {
		t.Fatalf("payload = %q want %q", got, want)
	}
	wantMsg, err := structpb.NewStruct(map[string]any{
		"advertised_capabilities": []any{"cap.peer.transport", "cap.peer.metadata"},
		"required_capabilities":   []any{"cap.peer.transport"},
		"operation":               string(OperationPeerInfo),
		"payload":                 base64.StdEncoding.EncodeToString([]byte{0x01, 0x02, 0x03}),
	})
	if err != nil {
		t.Fatalf("structpb.NewStruct() error = %v", err)
	}
	want, err := proto.MarshalOptions{Deterministic: true}.Marshal(wantMsg)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	if !bytes.Equal(raw, want) {
		t.Fatalf("transport wire = %x want %x", raw, want)
	}
}

func TestTransportResponseUsesProtobufStructEnvelope(t *testing.T) {
	resp := responseEnvelope{
		NegotiatedProtocol:   "/aether/peer/0.1.0",
		AcceptedCapabilities: []string{" cap.peer.transport ", "cap.peer.metadata"},
		IgnoredCapabilities:  []string{" cap.peer.future "},
		Payload:              []byte{0x09, 0x08},
		TransportError: Error{
			Code:             "unsupported-capability",
			Message:          "peer transport requires unsupported capabilities [cap.peer.experimental]",
			OfferedProtocols: []string{" /aether/peer/0.1.0 "},
			MissingRequired:  []string{" cap.peer.experimental "},
		},
	}
	raw, err := marshalResponse(resp)
	if err != nil {
		t.Fatalf("marshalResponse() error = %v", err)
	}
	var envelope structpb.Struct
	if err := proto.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("proto.Unmarshal() error = %v", err)
	}
	fields := envelope.GetFields()
	if got, want := stringValue(fields["negotiated_protocol"]), resp.NegotiatedProtocol; got != want {
		t.Fatalf("negotiated protocol = %q want %q", got, want)
	}
	if got, want := stringSliceFromValue(fields["accepted_capabilities"]), []string{"cap.peer.transport", "cap.peer.metadata"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("accepted capabilities = %#v want %#v", got, want)
	}
	if got, want := stringValue(fields["payload"]), base64.StdEncoding.EncodeToString(resp.Payload); got != want {
		t.Fatalf("payload = %q want %q", got, want)
	}
	if transportError := fields["transport_error"]; transportError == nil {
		t.Fatal("transport_error missing")
	} else if got, want := stringValue(transportError.GetStructValue().GetFields()["code"]), resp.TransportError.Code; got != want {
		t.Fatalf("transport error code = %q want %q", got, want)
	}
}

func TestNormalizePeerAddressCanonicalizesLibp2pMultiaddr(t *testing.T) {
	address, err := NormalizePeerAddress("/dns4/localhost/tcp/4242/p2p/12D3KooWQ6a8YcLU32y3fJNKC7xMbxD9xAfW7sY9qsKf8GugHgdR")
	if err != nil {
		t.Fatalf("NormalizePeerAddress() error = %v", err)
	}
	if got, want := address, "/ip4/127.0.0.1/tcp/4242/p2p/12D3KooWQ6a8YcLU32y3fJNKC7xMbxD9xAfW7sY9qsKf8GugHgdR"; got != want {
		t.Fatalf("address = %q want %q", got, want)
	}
}

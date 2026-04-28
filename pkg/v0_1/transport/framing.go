// Package transport implements the 4-byte-length PeerStream framing layer
// per docs/spec/v0.1/30-transport-and-noise.md §3.1–3.2.
//
// Wire format: [4-byte big-endian uint32 length][proto.Marshal(PeerStreamRequest)]
// Inner payload bytes within PeerStreamRequest/Response remain JSON for v0.1.
package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protowire"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
)

const (
	// MaxFrameSize is the maximum allowed PeerStream payload (8 MiB, spec 30 §3.2).
	MaxFrameSize = 8 << 20 // 8 388 608 bytes

	frameLenSize = 4 // uint32 big-endian
)

// ErrFrameTooLarge is returned when the declared frame length exceeds MaxFrameSize.
var ErrFrameTooLarge = errors.New("peerstream: frame exceeds 8 MiB limit")

// Proto field numbers — PeerStreamRequest (spec 30 §3.1, proto/aether.proto:1103)
const (
	reqFieldOperation    protowire.Number = 1
	reqFieldPayload      protowire.Number = 2
	reqFieldAdvCaps      protowire.Number = 3
	reqFieldReqCaps      protowire.Number = 4
	reqFieldProtocolID   protowire.Number = 5
	reqFieldSecurityMode protowire.Number = 6
	reqFieldRequestID    protowire.Number = 7
)

// Proto field numbers — PeerStreamResponse (proto/aether.proto:1119)
const (
	respFieldNegotiatedProtocol protowire.Number = 1
	respFieldAcceptedCaps       protowire.Number = 2
	respFieldIgnoredCaps        protowire.Number = 3
	respFieldPayload            protowire.Number = 4
	respFieldError              protowire.Number = 5
	respFieldRequestID          protowire.Number = 6
)

// Proto field numbers — PeerStreamError (proto/aether.proto:1090)
const (
	errFieldCode                protowire.Number = 1
	errFieldMessage             protowire.Number = 2
	errFieldMissingCapabilities protowire.Number = 3
	errFieldUnsupportedVersion  protowire.Number = 4
)

// SecurityMode enum values (proto/aether.proto:656, spec 04).
// Values 1–6 are allocated per spec; 7 is reserved.
func securityModeToProto(mode string) int32 {
	switch mode {
	case "seal":
		return 1 // SECURITY_MODE_SEAL
	case "tree":
		return 2 // SECURITY_MODE_TREE
	case "clear":
		return 3 // SECURITY_MODE_CLEAR
	case "crowd":
		return 4 // SECURITY_MODE_CROWD
	case "channel":
		return 5 // SECURITY_MODE_CHANNEL
	case "mediashield":
		return 6 // SECURITY_MODE_MEDIA_SHIELD
	default:
		return 0 // SECURITY_MODE_UNSPECIFIED
	}
}

func securityModeFromProto(v int32) string {
	switch v {
	case 1:
		return "seal"
	case 2:
		return "tree"
	case 3:
		return "clear"
	case 4:
		return "crowd"
	case 5:
		return "channel"
	case 6:
		return "mediashield"
	default:
		return ""
	}
}

// WriteRequest writes a PeerStreamRequest as a length-prefixed proto binary frame.
func WriteRequest(w io.Writer, req *proto.PeerStreamRequest) error {
	return writeFrameBytes(w, marshalRequest(req))
}

// ReadRequest reads a length-prefixed proto binary frame and decodes a PeerStreamRequest.
func ReadRequest(r io.Reader) (*proto.PeerStreamRequest, error) {
	buf, err := readFrameBytes(r)
	if err != nil {
		return nil, err
	}
	return unmarshalRequest(buf)
}

// WriteResponse writes a PeerStreamResponse as a length-prefixed proto binary frame.
func WriteResponse(w io.Writer, resp *proto.PeerStreamResponse) error {
	return writeFrameBytes(w, marshalResponse(resp))
}

// ReadResponse reads a length-prefixed proto binary frame and decodes a PeerStreamResponse.
func ReadResponse(r io.Reader) (*proto.PeerStreamResponse, error) {
	buf, err := readFrameBytes(r)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse(buf)
}

func writeFrameBytes(w io.Writer, payload []byte) error {
	if len(payload) > MaxFrameSize {
		return ErrFrameTooLarge
	}
	var lenBuf [frameLenSize]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	if _, err := w.Write(lenBuf[:]); err != nil {
		return fmt.Errorf("peerstream write: length prefix: %w", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("peerstream write: payload: %w", err)
	}
	return nil
}

func readFrameBytes(r io.Reader) ([]byte, error) {
	var lenBuf [frameLenSize]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, fmt.Errorf("peerstream read: length prefix: %w", err)
	}
	frameLen := binary.BigEndian.Uint32(lenBuf[:])
	if frameLen > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}
	payload := make([]byte, frameLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("peerstream read: payload: %w", err)
	}
	return payload, nil
}

// marshalRequest encodes a PeerStreamRequest into proto binary per field numbers above.
func marshalRequest(req *proto.PeerStreamRequest) []byte {
	var b []byte
	if req.Operation != "" {
		b = protowire.AppendTag(b, reqFieldOperation, protowire.BytesType)
		b = protowire.AppendString(b, req.Operation)
	}
	if len(req.Payload) > 0 {
		b = protowire.AppendTag(b, reqFieldPayload, protowire.BytesType)
		b = protowire.AppendBytes(b, req.Payload)
	}
	for _, cap := range req.AdvertisedCaps {
		b = protowire.AppendTag(b, reqFieldAdvCaps, protowire.BytesType)
		b = protowire.AppendString(b, cap)
	}
	for _, cap := range req.RequiredCaps {
		b = protowire.AppendTag(b, reqFieldReqCaps, protowire.BytesType)
		b = protowire.AppendString(b, cap)
	}
	if req.ProtocolID != "" {
		b = protowire.AppendTag(b, reqFieldProtocolID, protowire.BytesType)
		b = protowire.AppendString(b, req.ProtocolID)
	}
	if mode := securityModeToProto(req.SecurityMode); mode != 0 {
		b = protowire.AppendTag(b, reqFieldSecurityMode, protowire.VarintType)
		b = protowire.AppendVarint(b, uint64(mode))
	}
	if req.RequestID != "" {
		b = protowire.AppendTag(b, reqFieldRequestID, protowire.BytesType)
		b = protowire.AppendString(b, req.RequestID)
	}
	return b
}

func unmarshalRequest(b []byte) (*proto.PeerStreamRequest, error) {
	req := &proto.PeerStreamRequest{}
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("peerstream read: bad tag: %w", protowire.ParseError(n))
		}
		b = b[n:]
		switch {
		case num == reqFieldOperation && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: operation: %w", protowire.ParseError(n))
			}
			req.Operation = v
			b = b[n:]
		case num == reqFieldPayload && typ == protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: payload: %w", protowire.ParseError(n))
			}
			req.Payload = append([]byte(nil), v...)
			b = b[n:]
		case num == reqFieldAdvCaps && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: advertised_caps: %w", protowire.ParseError(n))
			}
			req.AdvertisedCaps = append(req.AdvertisedCaps, v)
			b = b[n:]
		case num == reqFieldReqCaps && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: required_caps: %w", protowire.ParseError(n))
			}
			req.RequiredCaps = append(req.RequiredCaps, v)
			b = b[n:]
		case num == reqFieldProtocolID && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: protocol_id: %w", protowire.ParseError(n))
			}
			req.ProtocolID = v
			b = b[n:]
		case num == reqFieldSecurityMode && typ == protowire.VarintType:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: security_mode: %w", protowire.ParseError(n))
			}
			req.SecurityMode = securityModeFromProto(int32(v))
			b = b[n:]
		case num == reqFieldRequestID && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: request_id: %w", protowire.ParseError(n))
			}
			req.RequestID = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: unknown field %d: %w", num, protowire.ParseError(n))
			}
			b = b[n:]
		}
	}
	return req, nil
}

// marshalResponse encodes a PeerStreamResponse into proto binary.
// Enforces spec 02 §1.2: when error is set, payload MUST be empty.
func marshalResponse(resp *proto.PeerStreamResponse) []byte {
	if resp.Error != nil && len(resp.Payload) > 0 {
		// Spec violation: clear payload when returning an error.
		resp = &proto.PeerStreamResponse{
			NegotiatedProtocol: resp.NegotiatedProtocol,
			AcceptedCaps:       resp.AcceptedCaps,
			IgnoredCaps:        resp.IgnoredCaps,
			Error:              resp.Error,
			RequestID:          resp.RequestID,
		}
	}
	var b []byte
	if resp.NegotiatedProtocol != "" {
		b = protowire.AppendTag(b, respFieldNegotiatedProtocol, protowire.BytesType)
		b = protowire.AppendString(b, resp.NegotiatedProtocol)
	}
	for _, cap := range resp.AcceptedCaps {
		b = protowire.AppendTag(b, respFieldAcceptedCaps, protowire.BytesType)
		b = protowire.AppendString(b, cap)
	}
	for _, cap := range resp.IgnoredCaps {
		b = protowire.AppendTag(b, respFieldIgnoredCaps, protowire.BytesType)
		b = protowire.AppendString(b, cap)
	}
	if len(resp.Payload) > 0 {
		b = protowire.AppendTag(b, respFieldPayload, protowire.BytesType)
		b = protowire.AppendBytes(b, resp.Payload)
	}
	if resp.Error != nil {
		errBytes := marshalError(resp.Error)
		b = protowire.AppendTag(b, respFieldError, protowire.BytesType)
		b = protowire.AppendBytes(b, errBytes)
	}
	if resp.RequestID != "" {
		b = protowire.AppendTag(b, respFieldRequestID, protowire.BytesType)
		b = protowire.AppendString(b, resp.RequestID)
	}
	return b
}

func unmarshalResponse(b []byte) (*proto.PeerStreamResponse, error) {
	resp := &proto.PeerStreamResponse{}
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("peerstream read: bad tag: %w", protowire.ParseError(n))
		}
		b = b[n:]
		switch {
		case num == respFieldNegotiatedProtocol && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: negotiated_protocol: %w", protowire.ParseError(n))
			}
			resp.NegotiatedProtocol = v
			b = b[n:]
		case num == respFieldAcceptedCaps && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: accepted_caps: %w", protowire.ParseError(n))
			}
			resp.AcceptedCaps = append(resp.AcceptedCaps, v)
			b = b[n:]
		case num == respFieldIgnoredCaps && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: ignored_caps: %w", protowire.ParseError(n))
			}
			resp.IgnoredCaps = append(resp.IgnoredCaps, v)
			b = b[n:]
		case num == respFieldPayload && typ == protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: payload: %w", protowire.ParseError(n))
			}
			resp.Payload = append([]byte(nil), v...)
			b = b[n:]
		case num == respFieldError && typ == protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: error: %w", protowire.ParseError(n))
			}
			e, err := unmarshalError(v)
			if err != nil {
				return nil, fmt.Errorf("peerstream read: error field: %w", err)
			}
			resp.Error = e
			b = b[n:]
		case num == respFieldRequestID && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: request_id: %w", protowire.ParseError(n))
			}
			resp.RequestID = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("peerstream read: unknown field %d: %w", num, protowire.ParseError(n))
			}
			b = b[n:]
		}
	}
	return resp, nil
}

func marshalError(e *proto.PeerStreamError) []byte {
	var b []byte
	if e.Code != "" {
		b = protowire.AppendTag(b, errFieldCode, protowire.BytesType)
		b = protowire.AppendString(b, e.Code)
	}
	if e.Message != "" {
		b = protowire.AppendTag(b, errFieldMessage, protowire.BytesType)
		b = protowire.AppendString(b, e.Message)
	}
	for _, cap := range e.MissingCapabilities {
		b = protowire.AppendTag(b, errFieldMissingCapabilities, protowire.BytesType)
		b = protowire.AppendString(b, cap)
	}
	if e.UnsupportedVersion != "" {
		b = protowire.AppendTag(b, errFieldUnsupportedVersion, protowire.BytesType)
		b = protowire.AppendString(b, e.UnsupportedVersion)
	}
	return b
}

func unmarshalError(b []byte) (*proto.PeerStreamError, error) {
	e := &proto.PeerStreamError{}
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("bad tag: %w", protowire.ParseError(n))
		}
		b = b[n:]
		switch {
		case num == errFieldCode && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("code: %w", protowire.ParseError(n))
			}
			e.Code = v
			b = b[n:]
		case num == errFieldMessage && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("message: %w", protowire.ParseError(n))
			}
			e.Message = v
			b = b[n:]
		case num == errFieldMissingCapabilities && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("missing_capabilities: %w", protowire.ParseError(n))
			}
			e.MissingCapabilities = append(e.MissingCapabilities, v)
			b = b[n:]
		case num == errFieldUnsupportedVersion && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return nil, fmt.Errorf("unsupported_version: %w", protowire.ParseError(n))
			}
			e.UnsupportedVersion = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("unknown field %d: %w", num, protowire.ParseError(n))
			}
			b = b[n:]
		}
	}
	return e, nil
}

package protocol

// PeerStreamError is a structured failure from any stream operation or negotiation.
// Source: docs/spec/v0.1/02-canonical-envelope.md §3, proto PeerStreamError.
type PeerStreamError struct {
	Code                string   `json:"code"`
	Message             string   `json:"message,omitempty"`
	MissingCapabilities []string `json:"missing_capabilities,omitempty"`
	UnsupportedVersion  string   `json:"unsupported_version,omitempty"`
}

func (e *PeerStreamError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Code + ": " + e.Message
	}
	return e.Code
}

// PeerStreamRequest is the native request envelope over a libp2p stream.
// Wire format: [4-byte BE uint32 length][JSON bytes], max 8 MiB.
// Source: docs/spec/v0.1/02-canonical-envelope.md §1 (proto semantics preserved in JSON bridge).
type PeerStreamRequest struct {
	Operation      string   `json:"operation"`
	Payload        []byte   `json:"payload,omitempty"`
	AdvertisedCaps []string `json:"advertised_caps,omitempty"`
	RequiredCaps   []string `json:"required_caps,omitempty"`
	ProtocolID     string   `json:"protocol_id,omitempty"`
	SecurityMode   string   `json:"security_mode,omitempty"`
	RequestID      string   `json:"request_id,omitempty"`
}

// PeerStreamResponse is the native reply envelope over a libp2p stream.
type PeerStreamResponse struct {
	NegotiatedProtocol string           `json:"negotiated_protocol,omitempty"`
	AcceptedCaps       []string         `json:"accepted_caps,omitempty"`
	IgnoredCaps        []string         `json:"ignored_caps,omitempty"`
	Payload            []byte           `json:"payload,omitempty"`
	Error              *PeerStreamError `json:"error,omitempty"`
	RequestID          string           `json:"request_id,omitempty"`
}

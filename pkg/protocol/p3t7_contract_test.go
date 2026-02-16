package protocol

import (
	"bytes"
	"testing"

	pb "github.com/aether/code_aether/gen/go/proto"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

func TestP3T7PositiveNegotiation(t *testing.T) {
	protocolID, ok := NegotiateProtocol(FamilyChat, []ProtocolID{chatV01}, nil)
	if !ok || protocolID != chatV01 {
		t.Fatalf("unexpected negotiation result: got %v(%t) want %v(true)", protocolID, ok, chatV01)
	}
}

func TestP3T7MalformedAndDowngradeOffers(t *testing.T) {
	cases := []struct {
		name   string
		offers []ProtocolID
		want   bool
	}{
		{"mismatchedFamily", []ProtocolID{{Family: FamilyVoice, Version: ProtocolVersion{Major: 0, Minor: 1}}}, false},
		{"belowMinimumMinor", []ProtocolID{chatV01}, false},
		{"majorDowngradeAllowedByPolicy", []ProtocolID{{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 5}}}, true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var policy CompatibilityPolicy = VersionCompatibilityPolicy{allowMinorDowngrade: true, allowMajorDowngrade: true, minimumMinor: 1}
			if tt.name == "mismatchedFamily" {
				policy = nil
			}
			if tt.name == "belowMinimumMinor" {
				policy = VersionCompatibilityPolicy{minimumMinor: 2}
			}
			_, ok := NegotiateProtocol(FamilyChat, tt.offers, policy)
			if ok != tt.want {
				t.Fatalf("negotiation mismatch: got %v want %v", ok, tt.want)
			}
		})
	}
}

func TestP3T7SignatureFailureClassification(t *testing.T) {
	if pb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_SIGNATURE_MISMATCH.String() != "ENVELOPE_VERIFICATION_ERROR_SIGNATURE_MISMATCH" {
		t.Fatalf("unexpected signature error string: %s", pb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_SIGNATURE_MISMATCH)
	}
}

func TestP3T7SignedEnvelopeUnknownFieldForwardCompat(t *testing.T) {
	base := &pb.SignedEnvelope{
		EnvelopeId:  "forward",
		PayloadType: pb.PayloadType_PAYLOAD_TYPE_IDENTITY,
	}
	data, err := proto.Marshal(base)
	if err != nil {
		t.Fatalf("marshal base envelope: %v", err)
	}
	unknown := protowire.AppendVarint(nil, protowire.EncodeTag(99, protowire.VarintType))
	unknown = protowire.AppendVarint(unknown, 0x42)
	withUnknown := append(data, unknown...)
	var decoded pb.SignedEnvelope
	if unmarshalErr := proto.Unmarshal(withUnknown, &decoded); unmarshalErr != nil {
		t.Fatalf("unexpected error parsing unknown field: %v", unmarshalErr)
	}
	if decoded.EnvelopeId != base.EnvelopeId {
		t.Fatalf("envelope id changed: got %q want %q", decoded.EnvelopeId, base.EnvelopeId)
	}
	remarshaled, err := proto.Marshal(&decoded)
	if err != nil {
		t.Fatalf("remarshal decoded envelope: %v", err)
	}
	if !bytes.Contains(remarshaled, unknown) {
		t.Fatalf("unknown field was not preserved across unmarshal/remarshal")
	}
}

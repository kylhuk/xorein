package v24

import (
	"encoding/binary"
	"testing"

	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestLocalAPIFrameMalformedRefusal(t *testing.T) {
	if err := localapi.ValidateFrame([]byte{0, 1, 2}, 1024); err == nil {
		t.Fatalf("expected malformed frame refusal")
	} else if refusal, ok := err.(localapi.RefusalError); !ok || refusal.Reason != localapi.RefusalReasonMalformedFrame {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func TestLocalAPIFrameOversizeRefusal(t *testing.T) {
	frame := buildFrame(2048, []byte("payload"))
	if err := localapi.ValidateFrame(frame, 1024); err == nil {
		t.Fatalf("expected oversize frame refusal")
	} else if refusal, ok := err.(localapi.RefusalError); !ok || refusal.Reason != localapi.RefusalReasonOversizeFrame {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func buildFrame(payloadLen int, payload []byte) []byte {
	buf := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(buf[:4], uint32(payloadLen))
	copy(buf[4:], payload)
	return buf
}

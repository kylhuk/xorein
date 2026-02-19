package localapi

import (
	"encoding/binary"
	"fmt"
)

const frameHeaderSize = 4

func ValidateFrame(frame []byte, maxPayload int) error {
	if len(frame) < frameHeaderSize {
		return RefusalError{
			Reason: RefusalReasonMalformedFrame,
			Detail: fmt.Sprintf("frame length %d shorter than header %d", len(frame), frameHeaderSize),
		}
	}

	if maxPayload < 0 {
		maxPayload = 0
	}

	payloadLen := int(binary.BigEndian.Uint32(frame[:frameHeaderSize]))
	if payloadLen > maxPayload {
		return RefusalError{
			Reason: RefusalReasonOversizeFrame,
			Detail: fmt.Sprintf("payload %d exceeds max %d", payloadLen, maxPayload),
		}
	}

	available := len(frame) - frameHeaderSize
	if available < payloadLen {
		return RefusalError{
			Reason: RefusalReasonMalformedFrame,
			Detail: fmt.Sprintf("payload length %d but frame carries %d bytes", payloadLen, available),
		}
	}

	return nil
}

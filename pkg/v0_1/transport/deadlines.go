package transport

import (
	"context"
	"fmt"
	"time"

	libp2pnet "github.com/libp2p/go-libp2p/core/network"
	libp2phost "github.com/libp2p/go-libp2p/core/host"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
)

const (
	// readDeadline is the max time to wait for the full request body (spec 30 §3.2).
	readDeadline = 60 * time.Second
	// writeDeadline is the max time to write the response (spec 30 §3.2).
	writeDeadline = 30 * time.Second
	// halfOpenTimeout closes half-open streams with no activity (spec 30 §3.2).
	halfOpenTimeout = 120 * time.Second
)

// RegisterFamilyWithDeadlines wires a FamilyHandler with spec-mandated stream deadlines.
// Use this instead of RegisterFamily for all production family registrations.
func RegisterFamilyWithDeadlines(h libp2phost.Host, handler FamilyHandler) {
	h.SetStreamHandler(handler.ProtocolID(), func(s libp2pnet.Stream) {
		now := time.Now()
		_ = s.SetReadDeadline(now.Add(readDeadline))
		_ = s.SetWriteDeadline(now.Add(writeDeadline))

		// Half-open GC: close stream if caller never sends data.
		halfOpenTimer := time.AfterFunc(halfOpenTimeout, func() {
			s.Close() //nolint:errcheck
		})
		defer halfOpenTimer.Stop()
		defer s.Close()

		ctx := context.Background()

		req, err := ReadRequest(s)
		if err != nil {
			WriteResponse(s, &proto.PeerStreamResponse{ //nolint:errcheck
				Error: &proto.PeerStreamError{
					Code:    proto.CodeOperationFailed,
					Message: fmt.Sprintf("frame read: %v", err),
				},
			})
			return
		}
		halfOpenTimer.Stop() // request received; disable half-open GC

		// Clear read deadline to allow handler execution time.
		_ = s.SetReadDeadline(time.Time{})

		resp := handler.HandleStream(ctx, req)
		if resp == nil {
			resp = &proto.PeerStreamResponse{
				RequestID: req.RequestID,
				Error: &proto.PeerStreamError{
					Code:    proto.CodeOperationFailed,
					Message: "handler returned nil response",
				},
			}
		}
		resp.RequestID = req.RequestID

		WriteResponse(s, resp) //nolint:errcheck
	})
}

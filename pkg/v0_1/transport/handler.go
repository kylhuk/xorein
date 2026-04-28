package transport

import (
	"context"
	"fmt"

	libp2pnet "github.com/libp2p/go-libp2p/core/network"
	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"
	libp2phost "github.com/libp2p/go-libp2p/core/host"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
)

// FamilyHandler is implemented by each protocol family (peer, dm, identity, chat, etc.).
// RegisterFamily wires it into the libp2p host for the given protocol ID.
type FamilyHandler interface {
	// ProtocolID returns the canonical multistream protocol ID, e.g. "/aether/peer/0.1.0".
	ProtocolID() libp2pprotocol.ID

	// HandleStream processes an incoming PeerStreamRequest and returns a response.
	// The handler is responsible for capability negotiation via OperationRequiredCaps
	// and for populating NegotiatedProtocol, AcceptedCaps, etc. in the response.
	HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse
}

// RegisterFamily wires a FamilyHandler into the libp2p host under its ProtocolID.
// Incoming streams are framed, dispatched, and the response is framed back.
func RegisterFamily(h libp2phost.Host, handler FamilyHandler) {
	h.SetStreamHandler(handler.ProtocolID(), func(s libp2pnet.Stream) {
		defer s.Close()
		ctx := context.Background()

		req, err := ReadRequest(s)
		if err != nil {
			// Malformed frame — write a structured error and close.
			WriteResponse(s, errorResponse("", &proto.PeerStreamError{ //nolint:errcheck
				Code:    proto.CodeOperationFailed,
				Message: fmt.Sprintf("frame read: %v", err),
			}))
			return
		}

		resp := handler.HandleStream(ctx, req)
		if resp == nil {
			resp = errorResponse(req.RequestID, &proto.PeerStreamError{
				Code:    proto.CodeOperationFailed,
				Message: "handler returned nil response",
			})
		}
		resp.RequestID = req.RequestID

		WriteResponse(s, resp) //nolint:errcheck
	})
}

func errorResponse(requestID string, e *proto.PeerStreamError) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{
		RequestID: requestID,
		Error:     e,
	}
}

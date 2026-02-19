package journeys

import (
	"context"
	"fmt"

	"github.com/aether/code_aether/pkg/v24/harmolyn/attach"
	"github.com/aether/code_aether/pkg/v24/localapi"
)

// HistoryRequest captures the inputs for a history fetch.
type HistoryRequest struct {
	ChannelID string
	Limit     int
}

// LocalAPIClient defines the minimal local API surface consumed by harmolyn
// journeys. Every journey action accepts a valid session token to enforce
// token gating.
type LocalAPIClient interface {
	Send(ctx context.Context, token localapi.SessionToken, payload string) error
	Read(ctx context.Context, token localapi.SessionToken, cursor string) (string, error)
	Search(ctx context.Context, token localapi.SessionToken, query string) ([]string, error)
	FetchHistory(ctx context.Context, token localapi.SessionToken, req HistoryRequest) ([]string, error)
	MediaControl(ctx context.Context, token localapi.SessionToken, action string) error
}

// JourneyError exposes deterministic reasons and next-action guidance for
// degraded journey paths.
type JourneyError struct {
	Reason     attach.FailureReason
	NextAction attach.NextAction
	Detail     string
}

// Error implements the error interface for JourneyError.
func (j JourneyError) Error() string {
	return fmt.Sprintf("journey degraded: %s (%s) - %s", j.Reason, j.NextAction, j.Detail)
}

type attachProvider interface {
	Attach(ctx context.Context) attach.AttachResult
}

// Coordinator gates every journey through attach state and the local API
// client, satisfying ST1 and ST3 by disallowing direct relay/database access.
type Coordinator struct {
	Attach attachProvider
	Client LocalAPIClient
}

// NewCoordinator builds a journeys coordinator that depends on an attach
// coordinator and the local API client implementation.
func NewCoordinator(att attachProvider, client LocalAPIClient) *Coordinator {
	return &Coordinator{Attach: att, Client: client}
}

// Send routes a payload through the local API after ensuring attach state.
func (c *Coordinator) Send(ctx context.Context, payload string) error {
	token, err := c.ensureReady(ctx)
	if err != nil {
		return err
	}
	return c.Client.Send(ctx, token, payload)
}

// Read routes a cursor read through the local API after ensuring attach state.
func (c *Coordinator) Read(ctx context.Context, cursor string) (string, error) {
	token, err := c.ensureReady(ctx)
	if err != nil {
		return "", err
	}
	return c.Client.Read(ctx, token, cursor)
}

// Search routes query execution through the local API after ensuring attach state.
func (c *Coordinator) Search(ctx context.Context, query string) ([]string, error) {
	token, err := c.ensureReady(ctx)
	if err != nil {
		return nil, err
	}
	return c.Client.Search(ctx, token, query)
}

// FetchHistory routes history fetches through the local API after ensuring attach state.
func (c *Coordinator) FetchHistory(ctx context.Context, req HistoryRequest) ([]string, error) {
	token, err := c.ensureReady(ctx)
	if err != nil {
		return nil, err
	}
	return c.Client.FetchHistory(ctx, token, req)
}

// MediaControl routes media control requests through the local API after ensuring attach state.
func (c *Coordinator) MediaControl(ctx context.Context, action string) error {
	token, err := c.ensureReady(ctx)
	if err != nil {
		return err
	}
	return c.Client.MediaControl(ctx, token, action)
}

func (c *Coordinator) ensureReady(ctx context.Context) (localapi.SessionToken, *JourneyError) {
	if c.Client == nil {
		return localapi.SessionToken{}, c.buildEarlyError(attach.FailureReasonDaemonStartFailed, attach.NextActionRepair, "local API client missing")
	}
	token, err := c.ensureAttached(ctx)
	if err != nil {
		return localapi.SessionToken{}, err
	}
	return token, nil
}

func (c *Coordinator) ensureAttached(ctx context.Context) (localapi.SessionToken, *JourneyError) {
	if c.Attach == nil {
		return localapi.SessionToken{}, c.buildEarlyError(attach.FailureReasonDaemonStartFailed, attach.NextActionRepair, "attach coordinator missing")
	}
	result := c.Attach.Attach(ctx)
	if result.Attached && result.Token.Valid() {
		return result.Token, nil
	}
	if !result.Attached {
		if result.Failure != nil {
			return localapi.SessionToken{}, journeyErrorFromFailure(result.Failure)
		}
		return localapi.SessionToken{}, c.buildEarlyError(attach.FailureReasonDaemonStartFailed, attach.NextActionRetry, "attach failed without diagnostic")
	}
	return localapi.SessionToken{}, c.buildEarlyError(attach.FailureReasonAuthFailed, attach.NextActionReset, "attached token invalid")
}

func journeyErrorFromFailure(f *attach.Failure) *JourneyError {
	return &JourneyError{Reason: f.Reason, NextAction: f.NextAction, Detail: f.Detail}
}

func (c *Coordinator) buildEarlyError(reason attach.FailureReason, action attach.NextAction, detail string) *JourneyError {
	return &JourneyError{Reason: reason, NextAction: action, Detail: detail}
}

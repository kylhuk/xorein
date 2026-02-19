package v24

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v24/harmolyn/attach"
	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestAttachFailureUXStates(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		probe      attach.DaemonProbe
		starter    attach.DaemonStarter
		handshaker attach.Handshaker
		reason     attach.FailureReason
		nextAction attach.NextAction
	}{
		{
			name:       "start failure",
			probe:      &stubProbe{status: attach.DaemonStatus{Running: false}},
			starter:    &stubStarter{err: errExample},
			handshaker: &stubHandshaker{},
			reason:     attach.FailureReasonDaemonStartFailed,
			nextAction: attach.NextActionRetry,
		},
		{
			name:       "incompatible daemon",
			probe:      &stubProbe{status: attach.DaemonStatus{Running: true}},
			handshaker: &stubHandshaker{err: attach.ErrDaemonIncompatible},
			reason:     attach.FailureReasonDaemonIncompatible,
			nextAction: attach.NextActionRepair,
		},
		{
			name:       "auth failure",
			probe:      &stubProbe{status: attach.DaemonStatus{Running: true}},
			handshaker: &stubHandshaker{err: attach.ErrAuthFailed},
			reason:     attach.FailureReasonAuthFailed,
			nextAction: attach.NextActionReset,
		},
		{
			name:       "socket permission",
			probe:      &stubProbe{error: attach.ErrSocketPermissionDenied},
			handshaker: &stubHandshaker{},
			reason:     attach.FailureReasonSocketPermissionDenied,
			nextAction: attach.NextActionOpenLogs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord := &attach.Coordinator{Probe: tt.probe, Starter: tt.starter, Handshaker: tt.handshaker, Store: attach.NewMemoryTokenStore()}
			res := coord.Attach(ctx)
			if res.Attached {
				t.Fatalf("expected failure for %s", tt.name)
			}
			if res.Failure == nil {
				t.Fatalf("missing failure for %s", tt.name)
			}
			if res.Failure.Reason != tt.reason {
				t.Fatalf("unexpected reason for %s: got %v", tt.name, res.Failure.Reason)
			}
			if res.Failure.NextAction != tt.nextAction {
				t.Fatalf("unexpected action for %s: got %v", tt.name, res.Failure.NextAction)
			}
		})
	}
}

func TestAttachGracefulReattach(t *testing.T) {
	ctx := context.Background()
	handshake := &sequentialHandshaker{tokens: []localapi.SessionToken{
		{Value: "alpha", ExpiresAt: time.Now().Add(time.Minute)},
		{Value: "beta", ExpiresAt: time.Now().Add(2 * time.Minute)},
	}}
	probe := &stubProbe{status: attach.DaemonStatus{Running: false}}
	starter := &stubStarter{}
	coord := &attach.Coordinator{Probe: probe, Starter: starter, Handshaker: handshake, Store: attach.NewMemoryTokenStore()}

	first := coord.Attach(ctx)
	if !first.Attached || first.Failure != nil {
		t.Fatalf("initial attach failed: %v", first.Failure)
	}
	second := coord.Reattach(ctx)
	if !second.Attached || second.Failure != nil {
		t.Fatalf("reattach failed: %v", second.Failure)
	}
	if first.Token.Value == second.Token.Value {
		t.Fatalf("reattach did not refresh token: %s", first.Token.Value)
	}
	if handshake.calls < 2 {
		t.Fatalf("expected handshake twice, got %d", handshake.calls)
	}
}

var errExample = errors.New("example failure")

type stubProbe struct {
	status attach.DaemonStatus
	error  error
}

func (s *stubProbe) Probe(ctx context.Context) (attach.DaemonStatus, error) {
	return s.status, s.error
}

func (s *stubProbe) WaitReady(ctx context.Context) error {
	s.status.Running = true
	return nil
}

type stubStarter struct {
	err     error
	started bool
}

func (s *stubStarter) Start(ctx context.Context) error {
	s.started = true
	return s.err
}

type stubHandshaker struct {
	token localapi.SessionToken
	err   error
}

func (s *stubHandshaker) Handshake(ctx context.Context) (localapi.SessionToken, error) {
	return s.token, s.err
}

type sequentialHandshaker struct {
	tokens []localapi.SessionToken
	calls  int
}

func (s *sequentialHandshaker) Handshake(ctx context.Context) (localapi.SessionToken, error) {
	s.calls++
	idx := s.calls - 1
	if idx >= len(s.tokens) {
		idx = len(s.tokens) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return s.tokens[idx], nil
}

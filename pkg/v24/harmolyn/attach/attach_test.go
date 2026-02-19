package attach

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestAttachBootstrapsAndRecordsToken(t *testing.T) {
	ctx := context.Background()
	probe := &fakeProbe{status: DaemonStatus{Running: false}}
	starter := &fakeStarter{}
	handshake := &fakeHandshaker{tokens: []localapi.SessionToken{{Value: "token", ExpiresAt: time.Now().Add(time.Hour)}}}
	store := NewMemoryTokenStore()
	coord := &Coordinator{Probe: probe, Starter: starter, Handshaker: handshake, Store: store}

	res := coord.Attach(ctx)
	if !res.Attached {
		t.Fatalf("expected attached in success path")
	}
	if res.Failure != nil {
		t.Fatalf("unexpected failure: %v", res.Failure)
	}
	if !starter.called {
		t.Fatalf("expected starter to be invoked")
	}
	if !probe.waitCalled {
		t.Fatalf("expected wait ready after start")
	}
	saved, ok := store.Load()
	if !ok || saved.Value != "token" {
		t.Fatalf("token not recorded, got %v", saved)
	}
}

func TestAttachUsesCachedToken(t *testing.T) {
	ctx := context.Background()
	token := localapi.SessionToken{Value: "cached", ExpiresAt: time.Now().Add(time.Hour)}
	store := NewMemoryTokenStore()
	store.Save(token)
	probe := &fakeProbe{status: DaemonStatus{Running: false}}
	coord := &Coordinator{Probe: probe, Starter: &fakeStarter{}, Handshaker: &fakeHandshaker{}, Store: store}

	res := coord.Attach(ctx)
	if !res.Attached {
		t.Fatalf("expected attach success from cache")
	}
	if res.Failure != nil {
		t.Fatalf("unexpected failure: %v", res.Failure)
	}
	if probe.probeCalls != 0 {
		t.Fatalf("probe should not run when token cached")
	}
}

func TestAttachStartFailure(t *testing.T) {
	ctx := context.Background()
	probe := &fakeProbe{status: DaemonStatus{Running: false}}
	starter := &fakeStarter{err: errors.New("boom")}
	coord := &Coordinator{Probe: probe, Starter: starter, Handshaker: &fakeHandshaker{}, Store: NewMemoryTokenStore()}

	res := coord.Attach(ctx)
	if res.Attached {
		t.Fatalf("expected attach to fail")
	}
	if res.Failure == nil || res.Failure.Reason != FailureReasonDaemonStartFailed {
		t.Fatalf("wrong failure reason: %v", res.Failure)
	}
}

func TestAttachIncompatibleDaemon(t *testing.T) {
	ctx := context.Background()
	probe := &fakeProbe{status: DaemonStatus{Running: true}}
	handshake := &fakeHandshaker{err: ErrDaemonIncompatible}
	coord := &Coordinator{Probe: probe, Handshaker: handshake, Store: NewMemoryTokenStore()}

	res := coord.Attach(ctx)
	if res.Attached {
		t.Fatalf("expected attach to fail")
	}
	if res.Failure == nil || res.Failure.Reason != FailureReasonDaemonIncompatible {
		t.Fatalf("unexpected failure: %v", res.Failure)
	}
}

func TestAttachAuthFailure(t *testing.T) {
	ctx := context.Background()
	probe := &fakeProbe{status: DaemonStatus{Running: true}}
	handshake := &fakeHandshaker{err: ErrAuthFailed}
	coord := &Coordinator{Probe: probe, Handshaker: handshake, Store: NewMemoryTokenStore()}

	res := coord.Attach(ctx)
	if res.Attached {
		t.Fatalf("expected attach to fail")
	}
	if res.Failure == nil || res.Failure.Reason != FailureReasonAuthFailed {
		t.Fatalf("unexpected failure: %v", res.Failure)
	}
}

func TestAttachPermissionDenied(t *testing.T) {
	ctx := context.Background()
	probe := &fakeProbe{error: ErrSocketPermissionDenied}
	coord := &Coordinator{Probe: probe, Handshaker: &fakeHandshaker{}, Store: NewMemoryTokenStore()}

	res := coord.Attach(ctx)
	if res.Attached {
		t.Fatalf("expected attach to fail")
	}
	if res.Failure == nil || res.Failure.Reason != FailureReasonSocketPermissionDenied {
		t.Fatalf("unexpected failure: %v", res.Failure)
	}
}

func TestReattachForcesFreshHandshake(t *testing.T) {
	ctx := context.Background()
	handshake := &fakeHandshaker{tokens: []localapi.SessionToken{
		{Value: "first", ExpiresAt: time.Now().Add(time.Minute)},
		{Value: "second", ExpiresAt: time.Now().Add(time.Hour)},
	}}
	probe := &fakeProbe{status: DaemonStatus{Running: false}}
	starter := &fakeStarter{}
	coord := &Coordinator{Probe: probe, Starter: starter, Handshaker: handshake, Store: NewMemoryTokenStore()}

	first := coord.Attach(ctx)
	if !first.Attached || first.Failure != nil {
		t.Fatalf("initial attach failed: %v", first.Failure)
	}
	second := coord.Reattach(ctx)
	if !second.Attached || second.Failure != nil {
		t.Fatalf("reattach failed: %v", second.Failure)
	}
	if handshake.calls < 2 {
		t.Fatalf("expected handshake to run twice, got %d", handshake.calls)
	}
}

type fakeProbe struct {
	status     DaemonStatus
	error      error
	waitErr    error
	waitCalled bool
	probeCalls int
}

func (f *fakeProbe) Probe(ctx context.Context) (DaemonStatus, error) {
	f.probeCalls++
	if f.error != nil {
		return DaemonStatus{}, f.error
	}
	return f.status, nil
}

func (f *fakeProbe) WaitReady(ctx context.Context) error {
	f.waitCalled = true
	if f.waitErr == nil {
		f.status.Running = true
	}
	return f.waitErr
}

type fakeStarter struct {
	called bool
	err    error
}

func (f *fakeStarter) Start(ctx context.Context) error {
	f.called = true
	return f.err
}

type fakeHandshaker struct {
	tokens []localapi.SessionToken
	err    error
	calls  int
}

func (f *fakeHandshaker) Handshake(ctx context.Context) (localapi.SessionToken, error) {
	f.calls++
	if f.err != nil {
		return localapi.SessionToken{}, f.err
	}
	if len(f.tokens) == 0 {
		return localapi.SessionToken{Value: fmt.Sprintf("auto-%d", f.calls), ExpiresAt: time.Now().Add(time.Minute)}, nil
	}
	idx := f.calls - 1
	if idx >= len(f.tokens) {
		idx = len(f.tokens) - 1
	}
	return f.tokens[idx], nil
}

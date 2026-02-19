package attach

import (
	"context"
	"errors"
	"sync"

	"github.com/aether/code_aether/pkg/v24/localapi"
)

var (
	// ErrDaemonIncompatible surfaces when the daemon advertises an API
	// version that does not overlap with the UI's supported range.
	ErrDaemonIncompatible = errors.New("daemon incompatible")

	// ErrAuthFailed is raised when the daemon rejects a valid client
	// handshake or session token for deterministic reasons.
	ErrAuthFailed = errors.New("auth failed")

	// ErrSocketPermissionDenied hints at filesystem/acl rules preventing the
	// client from probing the daemon socket.
	ErrSocketPermissionDenied = errors.New("socket permission denied")
)

// FailureReason enumerates deterministic attach failure codes (phase0
// error-taxonomy aligned).
type FailureReason string

const (
	FailureReasonDaemonStartFailed      FailureReason = "DAEMON_START_FAILED"
	FailureReasonDaemonIncompatible     FailureReason = "DAEMON_INCOMPATIBLE"
	FailureReasonAuthFailed             FailureReason = "AUTH_FAILED"
	FailureReasonSocketPermissionDenied FailureReason = "SOCKET_PERMISSION_DENIED"
)

// NextAction is the composer-defined guidance that pairs with a failure
// reason.
type NextAction string

const (
	NextActionRetry    NextAction = "RETRY"
	NextActionOpenLogs NextAction = "OPEN_LOGS"
	NextActionRepair   NextAction = "REPAIR"
	NextActionReset    NextAction = "RESET"
)

// Failure captures a deterministic error message plus the next user action.
type Failure struct {
	Reason     FailureReason
	NextAction NextAction
	Detail     string
}

// AttachResult summarizes the attach attempt.
type AttachResult struct {
	Token    localapi.SessionToken
	Failure  *Failure
	Attached bool
}

// DaemonStatus is the lightweight probe result from the daemon.
type DaemonStatus struct {
	Running bool
	Version string
}

// DaemonProbe describes how harmolyn inspects the daemon socket without
// assuming how the daemon actually listens.
type DaemonProbe interface {
	Probe(ctx context.Context) (DaemonStatus, error)
	WaitReady(ctx context.Context) error
}

// DaemonStarter models the ability to start a daemon when it is missing.
type DaemonStarter interface {
	Start(ctx context.Context) error
}

// Handshaker performs the deterministic protocol negotiation with the daemon.
type Handshaker interface {
	Handshake(ctx context.Context) (localapi.SessionToken, error)
}

// TokenStore persists the session token between attach attempts.
type TokenStore interface {
	Save(localapi.SessionToken)
	Load() (localapi.SessionToken, bool)
	Clear()
}

// MemoryTokenStore is the default TokenStore used in tests.
type MemoryTokenStore struct {
	mu    sync.Mutex
	token localapi.SessionToken
	ok    bool
}

// NewMemoryTokenStore yields an empty store suited for deterministic tests.
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{}
}

// Save overwrites the stored token.
func (s *MemoryTokenStore) Save(token localapi.SessionToken) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = token
	s.ok = true
}

// Load returns the saved token if it exists.
func (s *MemoryTokenStore) Load() (localapi.SessionToken, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ok {
		return localapi.SessionToken{}, false
	}
	return s.token, true
}

// Clear removes the stored session state.
func (s *MemoryTokenStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = localapi.SessionToken{}
	s.ok = false
}

// Coordinator orchestrates the attach lifecycle.
type Coordinator struct {
	Probe      DaemonProbe
	Starter    DaemonStarter
	Handshaker Handshaker
	Store      TokenStore
}

// Attach tries to ensure harmolyn has a valid session token with the daemon.
func (c *Coordinator) Attach(ctx context.Context) AttachResult {
	if c.Store == nil {
		c.Store = NewMemoryTokenStore()
	}

	if token, ok := c.Store.Load(); ok && token.Valid() {
		return AttachResult{Token: token, Attached: true}
	}

	status, err := c.safeProbe(ctx)
	if err != nil {
		return AttachResult{Failure: buildFailure(FailureReasonSocketPermissionDenied, NextActionOpenLogs, err)}
	}

	if !status.Running {
		if starterErr := c.safeStart(ctx); starterErr != nil {
			return AttachResult{Failure: buildFailure(FailureReasonDaemonStartFailed, NextActionRetry, starterErr)}
		}

		if waitErr := c.safeWait(ctx); waitErr != nil {
			return AttachResult{Failure: buildFailure(FailureReasonDaemonStartFailed, NextActionRetry, waitErr)}
		}

		status, err = c.safeProbe(ctx)
		if err != nil {
			return AttachResult{Failure: buildFailure(FailureReasonSocketPermissionDenied, NextActionOpenLogs, err)}
		}
		if !status.Running {
			return AttachResult{Failure: buildFailure(FailureReasonDaemonStartFailed, NextActionRetry, errors.New("daemon failed to reach running state"))}
		}
	}

	token, err := c.safeHandshake(ctx)
	if err != nil {
		switch {
		case errors.Is(err, ErrDaemonIncompatible):
			return AttachResult{Failure: buildFailure(FailureReasonDaemonIncompatible, NextActionRepair, err)}
		case errors.Is(err, ErrAuthFailed):
			return AttachResult{Failure: buildFailure(FailureReasonAuthFailed, NextActionReset, err)}
		default:
			return AttachResult{Failure: buildFailure(FailureReasonAuthFailed, NextActionReset, err)}
		}
	}

	c.Store.Save(token)
	return AttachResult{Token: token, Attached: true}
}

// Detach removes any cached session state so attach can re-run.
func (c *Coordinator) Detach() {
	if c.Store != nil {
		c.Store.Clear()
	}
}

// Reattach performs a graceful detach then runs Attach again (covers daemon restart).
func (c *Coordinator) Reattach(ctx context.Context) AttachResult {
	c.Detach()
	return c.Attach(ctx)
}

func (c *Coordinator) safeProbe(ctx context.Context) (DaemonStatus, error) {
	if c.Probe == nil {
		return DaemonStatus{}, errors.New("probe unavailable")
	}
	return c.Probe.Probe(ctx)
}

func (c *Coordinator) safeStart(ctx context.Context) error {
	if c.Starter == nil {
		return errors.New("starter unavailable")
	}
	return c.Starter.Start(ctx)
}

func (c *Coordinator) safeWait(ctx context.Context) error {
	if c.Probe == nil {
		return errors.New("probe unavailable")
	}
	return c.Probe.WaitReady(ctx)
}

func (c *Coordinator) safeHandshake(ctx context.Context) (localapi.SessionToken, error) {
	if c.Handshaker == nil {
		return localapi.SessionToken{}, errors.New("handshaker unavailable")
	}
	return c.Handshaker.Handshake(ctx)
}

func buildFailure(reason FailureReason, action NextAction, err error) *Failure {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return &Failure{Reason: reason, NextAction: action, Detail: msg}
}

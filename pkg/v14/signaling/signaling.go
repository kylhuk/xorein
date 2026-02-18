package signaling

import (
	"fmt"
	"sync"
	"time"
)

// State describes the current position in the signaling lifecycle.
type State string

const (
	StateIdle    State = "idle"
	StateCreated State = "created"
	StateJoined  State = "joined"
	StateLeft    State = "left"
)

// Operation describes a lifecycle action for error reporting.
type Operation string

const (
	OpCreate Operation = "create"
	OpJoin   Operation = "join"
	OpLeave  Operation = "leave"
	OpRetry  Operation = "retry"
)

// SignalingError is the deterministic, contract-level taxonomy for signaling failures.
type SignalingError struct {
	Code    string
	Op      Operation
	Message string
}

// Error implements error.
func (e SignalingError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Op, e.Message)
}

// Lifecycle manages a deterministic signaling session lifecycle.
type Lifecycle struct {
	mu         sync.Mutex
	state      State
	retries    int
	maxRetries int
	lastOp     Operation
}

// NewLifecycle returns a Lifecycle with deterministic retry defaults.
func NewLifecycle(maxRetries int) *Lifecycle {
	if maxRetries < 1 {
		maxRetries = 3
	}
	return &Lifecycle{state: StateIdle, maxRetries: maxRetries}
}

// CurrentState returns the lifecycle state.
func (l *Lifecycle) CurrentState() State {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.state
}

// CreateRoom marks the beginning of the lifecycle.
func (l *Lifecycle) CreateRoom() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state != StateIdle && l.state != StateLeft {
		return SignalingError{Code: "invalid-transition", Op: OpCreate, Message: "room cannot be created from current state"}
	}
	l.state = StateCreated
	l.retries = 0
	l.lastOp = OpCreate
	return nil
}

// JoinRoom moves to the joined state deterministically.
func (l *Lifecycle) JoinRoom() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state != StateCreated && l.state != StateLeft {
		return SignalingError{Code: "invalid-transition", Op: OpJoin, Message: "join requires an existing room"}
	}
	l.state = StateJoined
	l.lastOp = OpJoin
	return nil
}

// LeaveRoom returns to the idle state and records the leave operation.
func (l *Lifecycle) LeaveRoom() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state != StateJoined {
		return SignalingError{Code: "not-joined", Op: OpLeave, Message: "no active room to leave"}
	}
	l.state = StateLeft
	l.lastOp = OpLeave
	return nil
}

// Retry attempts to rejoin the last room.
func (l *Lifecycle) Retry() (time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state != StateCreated && l.state != StateJoined {
		return 0, SignalingError{Code: "not-ready", Op: OpRetry, Message: "retry requires an active room"}
	}
	if l.retries >= l.maxRetries {
		return 0, SignalingError{Code: "retry-limit", Op: OpRetry, Message: "retry budget exhausted"}
	}
	l.retries++
	l.lastOp = OpRetry
	backoff := time.Duration(l.retries) * 250 * time.Millisecond
	return backoff, nil
}

// LastOperation returns the last operation executed.
func (l *Lifecycle) LastOperation() Operation {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lastOp
}

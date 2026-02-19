package daemon

import (
	"fmt"
	"sync"
)

// LockReason names refusal causes from the singleton guard.
type LockReason string

const (
	LockUnavailable       LockReason = "lock unavailable"
	LockInvalidTransition LockReason = "invalid transition"
)

// RefusalError is returned when the daemon refuses to run because of
// deterministic conditions (busy lock, invalid state change, etc.).
type RefusalError struct {
	Reason  LockReason
	Details string
}

func (e *RefusalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Reason, e.Details)
}

// LockManager owns the lifecycle lock path and enforces deterministic states.
type LockManager struct {
	mu    sync.Mutex
	path  string
	state State
	owner string
	lock  bool
}

// AcquiredLock describes the held lock and current lifecycle state.
type AcquiredLock struct {
	manager *LockManager
	owner   string
	state   State
}

// NewLockManager returns a manager anchored to the provided path.
func NewLockManager(path string) *LockManager {
	return &LockManager{path: path, state: StateStopped}
}

// Acquire tries to take the lifecycle lock and move to `next` state.
func (l *LockManager) Acquire(owner string, next State) (*AcquiredLock, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lock && l.owner != owner {
		return nil, &RefusalError{
			Reason:  LockUnavailable,
			Details: fmt.Sprintf("path=%s owned by %s", l.path, l.owner),
		}
	}

	if err := ValidateTransition(l.state, next); err != nil {
		return nil, &RefusalError{
			Reason:  LockInvalidTransition,
			Details: err.Error(),
		}
	}

	l.lock = true
	l.owner = owner
	l.state = next

	return &AcquiredLock{manager: l, owner: owner, state: next}, nil
}

// Release relinquishes the lock and advances to `next` state.
func (a *AcquiredLock) Release(next State) error {
	mgr := a.manager
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if mgr.owner != a.owner {
		return fmt.Errorf("lock release mismatch: expected %s got %s", a.owner, mgr.owner)
	}

	if err := ValidateTransition(a.state, next); err != nil {
		return err
	}

	mgr.state = next
	mgr.owner = ""
	mgr.lock = false

	return nil
}

// State returns the agent's current lifecycle phase.
func (l *LockManager) State() State {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.state
}

// Owner returns the current lock owner (empty if unlocked).
func (l *LockManager) Owner() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.owner
}

// Path returns the locked filesystem path configured for this manager.
func (l *LockManager) Path() string {
	return l.path
}

// IsLocked reports whether the guard is currently held.
func (l *LockManager) IsLocked() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lock
}

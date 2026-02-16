package presence

import (
	"fmt"
	"strings"
	"time"
)

type State string

const (
	StateOnline    State = "online"
	StateIdle      State = "idle"
	StateDND       State = "dnd"
	StateInvisible State = "invisible"
	StateOffline   State = "offline"
)

type Event string

const (
	EventIdleTimeout Event = "idle-timeout"
	EventActivity    Event = "activity"
	EventDisconnect  Event = "disconnect"
)

type TransitionReason string

const (
	ReasonTransitionApplied   TransitionReason = "transition-applied"
	ReasonIgnoredByPrecedence TransitionReason = "ignored-by-precedence"
)

func NextState(current State, event Event) (State, TransitionReason) {
	if current == StateDND && event == EventIdleTimeout {
		return current, ReasonIgnoredByPrecedence
	}
	switch event {
	case EventIdleTimeout:
		if current == StateOnline {
			return StateIdle, ReasonTransitionApplied
		}
	case EventActivity:
		if current == StateIdle || current == StateOnline {
			return StateOnline, ReasonTransitionApplied
		}
	case EventDisconnect:
		return StateOffline, ReasonTransitionApplied
	}
	return current, ReasonIgnoredByPrecedence
}

func ValidateCustomStatus(status string, maxLen int) (string, error) {
	normalized := strings.TrimSpace(status)
	if maxLen < 1 {
		return "", fmt.Errorf("maxLen must be >= 1")
	}
	if len(normalized) > maxLen {
		return "", fmt.Errorf("status length exceeds max")
	}
	return normalized, nil
}

func IsStale(lastUpdated, now time.Time, ttl time.Duration) bool {
	if ttl <= 0 {
		return true
	}
	return now.Sub(lastUpdated) > ttl
}

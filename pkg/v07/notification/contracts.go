package notification

import "time"

type TriggerType string

const (
	TriggerDesktop TriggerType = "notification.desktop"
	TriggerPush    TriggerType = "notification.push"
)

type DedupeWindow struct {
	LastFired time.Time
	Interval  time.Duration
}

func (d DedupeWindow) ShouldSuppress(now time.Time) bool {
	if d.Interval <= 0 {
		return false
	}
	return now.Sub(d.LastFired) < d.Interval
}

func ResolveAction(trigger TriggerType, fallback string) string {
	if trigger == TriggerDesktop {
		if fallback != "" {
			return fallback
		}
		return "desktop:show"
	}
	return "push:show"
}

func SuppressionReason(active bool, mode TriggerType) string {
	if active {
		return string(mode) + ".suppressed"
	}
	return string(mode) + ".ready"
}

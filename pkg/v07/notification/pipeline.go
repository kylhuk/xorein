package notification

import (
	"fmt"
	"strings"
	"time"
)

type NotificationKind string

const (
	KindDM         NotificationKind = "notification.kind.dm"
	KindMention    NotificationKind = "notification.kind.mention"
	KindCallInvite NotificationKind = "notification.kind.call"
)

type Notification struct {
	Kind      NotificationKind
	ServerID  string
	ChannelID string
	MessageID string
	Actor     string
	Trigger   TriggerType
	Payload   string
	Timestamp time.Time
}

type NotificationAction struct {
	Trigger TriggerType
	Target  ActionTarget
	Label   string
}

type Pipeline struct {
	dedupe   map[string]DedupeWindow
	interval time.Duration
}

func NewPipeline(interval time.Duration) *Pipeline {
	return &Pipeline{dedupe: make(map[string]DedupeWindow), interval: interval}
}

type ActionTarget struct {
	ServerID  string
	ChannelID string
	MessageID string
	Kind      NotificationKind
}

func ParseActionTarget(raw string) ActionTarget {
	parts := strings.Split(raw, ":")
	target := ActionTarget{Kind: KindDM}
	if len(parts) >= 1 {
		target.ServerID = parts[0]
	}
	if len(parts) >= 2 {
		target.ChannelID = parts[1]
	}
	if len(parts) >= 3 {
		target.MessageID = parts[2]
	}
	return target
}

func (p *Pipeline) Fire(notif Notification, now time.Time) (NotificationAction, string) {
	key := fmt.Sprintf("%s-%s-%s", notif.ServerID, notif.ChannelID, notif.MessageID)
	window, ok := p.dedupe[key]
	if !ok {
		window = DedupeWindow{LastFired: time.Time{}, Interval: p.interval}
	}
	suppressed := window.ShouldSuppress(now)
	if !suppressed {
		window.LastFired = now
		p.dedupe[key] = window
	}
	target := ParseActionTarget(fmt.Sprintf("%s:%s:%s", notif.ServerID, notif.ChannelID, notif.MessageID))
	actionLabel := ResolveAction(notif.Trigger, fmt.Sprintf("server:%s", notif.ServerID))
	status := SuppressionReason(suppressed, notif.Trigger)
	return NotificationAction{Trigger: notif.Trigger, Target: target, Label: actionLabel}, status
}

func BuildPayloadTarget(server, channel, message string, kind NotificationKind) string {
	return fmt.Sprintf("%s:%s:%s", server, channel, message)
}

func DetermineTrigger(kind NotificationKind) TriggerType {
	switch kind {
	case KindDM, KindMention:
		return TriggerDesktop
	case KindCallInvite:
		return TriggerPush
	}
	return TriggerDesktop
}

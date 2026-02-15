package phase4

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalidServerID    = errors.New("phase4: invalid server id")
	ErrInvalidChannelID   = errors.New("phase4: invalid channel id")
	ErrUnknownTopic       = errors.New("phase4: unknown topic")
	ErrAdmissionRejected  = errors.New("phase4: message rejected by admission policy")
	ErrRetryLimitExceeded = errors.New("phase4: retry limit exceeded")
)

// TopicFormat captures deterministic v0.1 topic conventions for pubsub channels.
type TopicFormat struct {
	Prefix string
}

// TopicFor returns the deterministic GossipSub topic name for a server/channel pair.
func (f TopicFormat) TopicFor(serverID, channelID string) (string, error) {
	normalizedServerID, err := normalizeTopicIdentifier(serverID)
	if err != nil {
		return "", ErrInvalidServerID
	}
	normalizedChannelID, err := normalizeTopicIdentifier(channelID)
	if err != nil {
		return "", ErrInvalidChannelID
	}
	prefix := strings.TrimSpace(f.Prefix)
	if prefix == "" {
		prefix = "aether/v0.1"
	}
	prefix = strings.TrimSuffix(prefix, "/")
	return fmt.Sprintf("%s/server/%s/channel/%s", prefix, normalizedServerID, normalizedChannelID), nil
}

// AdmissionCheck allows callers to inject validation for inbound/outbound payloads.
// Returning nil admits the payload; returning an error rejects it.
type AdmissionCheck func(topic string, payload []byte) error

// RetryPolicy bounds publish retries for deterministic failure behavior.
type RetryPolicy struct {
	MaxAttempts int
}

func (p RetryPolicy) maxAttempts() int {
	if p.MaxAttempts <= 0 {
		return 1
	}
	return p.MaxAttempts
}

// PublishFunc abstracts the underlying pubsub publish transport.
type PublishFunc func(topic string, payload []byte) error

// SubscriptionEventType captures lifecycle transitions for topic subscriptions.
type SubscriptionEventType string

const (
	SubscriptionEventJoin  SubscriptionEventType = "join"
	SubscriptionEventLeave SubscriptionEventType = "leave"
)

// SubscriptionEvent records deterministic lifecycle events for observability/tests.
type SubscriptionEvent struct {
	Type      SubscriptionEventType
	Topic     string
	MemberID  string
	Remaining int
}

// PubSubService is a local integration facade for deterministic topic mapping,
// pluggable admission checks, lifecycle events, and bounded publish retries.
type PubSubService struct {
	mu sync.Mutex

	format        TopicFormat
	admission     AdmissionCheck
	retryPolicy   RetryPolicy
	subscriptions map[string]map[string]struct{}
	events        []SubscriptionEvent
}

// NewPubSubService constructs a baseline pubsub service for P4-T3.
func NewPubSubService(format TopicFormat, admission AdmissionCheck, retry RetryPolicy) *PubSubService {
	return &PubSubService{
		format:        format,
		admission:     admission,
		retryPolicy:   retry,
		subscriptions: make(map[string]map[string]struct{}),
	}
}

// TopicFor proxies deterministic topic naming through the configured format.
func (s *PubSubService) TopicFor(serverID, channelID string) (string, error) {
	return s.format.TopicFor(serverID, channelID)
}

// Join registers a member against a server/channel topic and records lifecycle events.
func (s *PubSubService) Join(memberID, serverID, channelID string) (string, error) {
	topic, err := s.TopicFor(serverID, channelID)
	if err != nil {
		return "", err
	}
	normalizedMember, err := normalizeTopicIdentifier(memberID)
	if err != nil {
		return "", fmt.Errorf("phase4: invalid member id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	members, ok := s.subscriptions[topic]
	if !ok {
		members = make(map[string]struct{})
		s.subscriptions[topic] = members
	}
	members[normalizedMember] = struct{}{}
	s.events = append(s.events, SubscriptionEvent{
		Type:      SubscriptionEventJoin,
		Topic:     topic,
		MemberID:  normalizedMember,
		Remaining: len(members),
	})
	return topic, nil
}

// Leave unregisters a member and records a lifecycle event.
func (s *PubSubService) Leave(memberID, serverID, channelID string) (string, error) {
	topic, err := s.TopicFor(serverID, channelID)
	if err != nil {
		return "", err
	}
	normalizedMember, err := normalizeTopicIdentifier(memberID)
	if err != nil {
		return "", fmt.Errorf("phase4: invalid member id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	members, ok := s.subscriptions[topic]
	if !ok {
		return "", ErrUnknownTopic
	}
	delete(members, normalizedMember)
	remaining := len(members)
	if remaining == 0 {
		delete(s.subscriptions, topic)
	}
	s.events = append(s.events, SubscriptionEvent{
		Type:      SubscriptionEventLeave,
		Topic:     topic,
		MemberID:  normalizedMember,
		Remaining: remaining,
	})
	return topic, nil
}

// SubscribedTopics returns deterministic sorted topic membership for a member.
func (s *PubSubService) SubscribedTopics(memberID string) ([]string, error) {
	normalizedMember, err := normalizeTopicIdentifier(memberID)
	if err != nil {
		return nil, fmt.Errorf("phase4: invalid member id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	topics := make([]string, 0)
	for topic, members := range s.subscriptions {
		if _, ok := members[normalizedMember]; ok {
			topics = append(topics, topic)
		}
	}
	sort.Strings(topics)
	return topics, nil
}

// Events returns a copy of subscription lifecycle events in append order.
func (s *PubSubService) Events() []SubscriptionEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make([]SubscriptionEvent, len(s.events))
	copy(copied, s.events)
	return copied
}

// Publish applies admission checks then calls publish with bounded retry behavior.
// The returned value is the number of publish attempts performed.
func (s *PubSubService) Publish(topic string, payload []byte, publish PublishFunc) (int, error) {
	if strings.TrimSpace(topic) == "" {
		return 0, ErrUnknownTopic
	}
	if publish == nil {
		return 0, fmt.Errorf("phase4: publish function is required")
	}
	if s.admission != nil {
		if err := s.admission(topic, payload); err != nil {
			return 0, fmt.Errorf("%w: %v", ErrAdmissionRejected, err)
		}
	}

	maxAttempts := s.retryPolicy.maxAttempts()
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := publish(topic, payload); err != nil {
			lastErr = err
			continue
		}
		return attempt, nil
	}
	if lastErr == nil {
		lastErr = ErrRetryLimitExceeded
	}
	return maxAttempts, fmt.Errorf("%w: %v", ErrRetryLimitExceeded, lastErr)
}

func normalizeTopicIdentifier(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", errors.New("empty")
	}
	for _, r := range normalized {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return "", fmt.Errorf("unsupported rune %q", r)
		}
	}
	return normalized, nil
}

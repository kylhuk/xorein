package chat

import (
	"errors"
	"strings"
	"time"
)

// DeliveryState represents deterministic messaging state.
type DeliveryState string

const (
	DeliveryStatePending   DeliveryState = "pending"
	DeliveryStateDelivered DeliveryState = "delivered"
	DeliveryStateFailed    DeliveryState = "failed"
)

var (
	ErrMessageIDRequired = errors.New("message id is required")
	ErrChannelIDRequired = errors.New("channel id is required")
	ErrSenderRequired    = errors.New("sender id is required")
	ErrBodyRequired      = errors.New("message body required")
	ErrInvalidTransition = errors.New("invalid delivery transition")
)

// Message captures stateless chat deliveries.
type Message struct {
	ID        string
	ChannelID string
	Body      string
	Sender    string
	State     DeliveryState
	CreatedAt time.Time
}

// NewMessage builds a baseline message in pending state.
func NewMessage(id, channelID, sender, body string, now time.Time) (Message, error) {
	if strings.TrimSpace(id) == "" {
		return Message{}, ErrMessageIDRequired
	}
	if strings.TrimSpace(channelID) == "" {
		return Message{}, ErrChannelIDRequired
	}
	if strings.TrimSpace(sender) == "" {
		return Message{}, ErrSenderRequired
	}
	if strings.TrimSpace(body) == "" {
		return Message{}, ErrBodyRequired
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return Message{ID: id, ChannelID: channelID, Sender: sender, Body: body, State: DeliveryStatePending, CreatedAt: now.UTC()}, nil
}

// Acknowledge transitions a pending message to delivered.
func (m *Message) Acknowledge() error {
	if m.State != DeliveryStatePending {
		return ErrInvalidTransition
	}
	m.State = DeliveryStateDelivered
	return nil
}

// Fail marks the message as failed if it was pending.
func (m *Message) Fail(reason string) error {
	if m.State != DeliveryStatePending {
		return ErrInvalidTransition
	}
	m.State = DeliveryStateFailed
	if reason != "" {
		m.Body = m.Body + " (fail: " + reason + ")"
	}
	return nil
}

// ReadMarker tracks progression of read receipts.
type ReadMarker struct {
	ChannelID     string
	Reader        string
	LatestMessage string
	Timestamp     time.Time
}

// Touch records a new read position.
func (r *ReadMarker) Touch(messageID, reader string, now time.Time) error {
	if strings.TrimSpace(messageID) == "" {
		return errors.New("message id required")
	}
	if strings.TrimSpace(reader) == "" {
		return errors.New("reader required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	r.LatestMessage = strings.TrimSpace(messageID)
	r.Reader = strings.TrimSpace(reader)
	r.Timestamp = now.UTC()
	return nil
}

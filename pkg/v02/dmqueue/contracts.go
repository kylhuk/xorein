package dmqueue

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// QueueConfig defines deterministic queue bounds for offline DM delivery.
type QueueConfig struct {
	MaxItems          int
	TTL               time.Duration
	ReplicationFactor int
	RetryWindow       time.Duration
	GCInterval        time.Duration
}

// Message is a minimal offline queue message contract.
type Message struct {
	ID             string
	RecipientID    string
	Ciphertext     []byte
	MailboxAddress string
}

// MailboxAddress is a deterministic mailbox routing descriptor.
type MailboxAddress struct {
	RecipientID string
	Scope       string
	Digest      string
}

// DeriveMailboxAddress computes a deterministic mailbox address for recipient scope.
func DeriveMailboxAddress(recipientID string, scope string) (MailboxAddress, error) {
	if recipientID == "" || scope == "" {
		return MailboxAddress{}, fmt.Errorf("recipient and scope required")
	}
	sum := sha256.Sum256([]byte(recipientID + "|" + scope))
	return MailboxAddress{
		RecipientID: recipientID,
		Scope:       scope,
		Digest:      hex.EncodeToString(sum[:16]),
	}, nil
}

// ReplicationTargets returns deterministic replica targets for a mailbox address.
func ReplicationTargets(address MailboxAddress, replicationFactor int) []string {
	if replicationFactor < 1 {
		replicationFactor = 1
	}
	targets := make([]string, 0, replicationFactor)
	for idx := 0; idx < replicationFactor; idx++ {
		targets = append(targets, fmt.Sprintf("%s#%d", address.Digest, idx))
	}
	return targets
}

type ReceiptStatus string

const (
	ReceiptStatusPending        ReceiptStatus = "pending"
	ReceiptStatusRetryScheduled ReceiptStatus = "retry-scheduled"
	ReceiptStatusDelivered      ReceiptStatus = "delivered"
	ReceiptStatusExpired        ReceiptStatus = "expired"
)

type RetryReason string

const (
	RetryReasonWindowOpen     RetryReason = "retry-window-open"
	RetryReasonWindowClosed   RetryReason = "retry-window-closed"
	RetryReasonAcknowledged   RetryReason = "retry-acknowledged"
	RetryReasonUnknownMessage RetryReason = "retry-unknown-message"
)

type Receipt struct {
	MessageID  string
	Status     ReceiptStatus
	Attempts   uint32
	RetryAfter time.Time
}

type queuedMessage struct {
	Message
	enqueuedAt time.Time
	seq        uint64
}

// Queue is an in-memory deterministic mailbox queue.
type Queue struct {
	cfg      QueueConfig
	items    []queuedMessage
	next     uint64
	acked    map[string]struct{}
	seenID   map[string]struct{}
	receipts map[string]Receipt
	lastGC   time.Time
}

func normalizeConfig(cfg QueueConfig) (QueueConfig, error) {
	if cfg.MaxItems <= 0 {
		return QueueConfig{}, fmt.Errorf("max items must be > 0")
	}
	if cfg.TTL <= 0 {
		return QueueConfig{}, fmt.Errorf("ttl must be > 0")
	}
	if cfg.ReplicationFactor <= 0 {
		cfg.ReplicationFactor = 1
	}
	if cfg.RetryWindow <= 0 {
		cfg.RetryWindow = cfg.TTL / 2
		if cfg.RetryWindow <= 0 {
			cfg.RetryWindow = time.Second
		}
	}
	if cfg.GCInterval <= 0 {
		cfg.GCInterval = cfg.TTL
	}
	return cfg, nil
}

// NewQueue creates a bounded offline queue.
func NewQueue(cfg QueueConfig) (*Queue, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Queue{cfg: normalized, acked: make(map[string]struct{}), seenID: make(map[string]struct{}), receipts: make(map[string]Receipt)}, nil
}

// Enqueue appends a message if valid and not duplicated.
func (q *Queue) Enqueue(now time.Time, msg Message) (bool, error) {
	if msg.ID == "" || msg.RecipientID == "" || len(msg.Ciphertext) == 0 {
		return false, fmt.Errorf("invalid message")
	}
	if msg.MailboxAddress == "" {
		address, err := DeriveMailboxAddress(msg.RecipientID, "offline")
		if err != nil {
			return false, err
		}
		msg.MailboxAddress = address.Digest
	}
	if _, exists := q.seenID[msg.ID]; exists {
		return false, nil
	}
	q.seenID[msg.ID] = struct{}{}
	q.items = append(q.items, queuedMessage{Message: msg, enqueuedAt: now.UTC(), seq: q.next})
	q.receipts[msg.ID] = Receipt{MessageID: msg.ID, Status: ReceiptStatusPending}
	q.next++
	for len(q.items) > q.cfg.MaxItems {
		dropped := q.items[0]
		q.items = q.items[1:]
		delete(q.seenID, dropped.ID)
		q.receipts[dropped.ID] = Receipt{MessageID: dropped.ID, Status: ReceiptStatusExpired}
	}
	return true, nil
}

// DequeueRecipient returns unexpired recipient messages ordered by enqueue sequence.
func (q *Queue) DequeueRecipient(now time.Time, recipientID string, limit int) []Message {
	if limit <= 0 {
		return nil
	}
	q.compact(now)
	out := make([]Message, 0)
	for _, item := range q.items {
		if item.RecipientID != recipientID {
			continue
		}
		if _, ok := q.acked[item.ID]; ok {
			continue
		}
		out = append(out, Message{ID: item.ID, RecipientID: item.RecipientID, Ciphertext: append([]byte(nil), item.Ciphertext...)})
		if len(out) == limit {
			break
		}
	}
	return out
}

// Ack marks a message as acknowledged once.
func (q *Queue) Ack(messageID string) bool {
	if messageID == "" {
		return false
	}
	if _, exists := q.acked[messageID]; exists {
		return false
	}
	if _, exists := q.seenID[messageID]; !exists {
		return false
	}
	q.acked[messageID] = struct{}{}
	receipt := q.receipts[messageID]
	receipt.MessageID = messageID
	receipt.Status = ReceiptStatusDelivered
	receipt.RetryAfter = time.Time{}
	q.receipts[messageID] = receipt
	return true
}

// MarkDeliveryAttempt records delivery attempt result and schedules retry windows.
func (q *Queue) MarkDeliveryAttempt(now time.Time, messageID string, delivered bool) (Receipt, bool) {
	receipt, ok := q.receipts[messageID]
	if !ok {
		return Receipt{}, false
	}
	receipt.MessageID = messageID
	receipt.Attempts++
	if delivered {
		q.acked[messageID] = struct{}{}
		receipt.Status = ReceiptStatusDelivered
		receipt.RetryAfter = time.Time{}
		q.receipts[messageID] = receipt
		return receipt, true
	}
	receipt.Status = ReceiptStatusRetryScheduled
	receipt.RetryAfter = now.UTC().Add(q.cfg.RetryWindow)
	q.receipts[messageID] = receipt
	return receipt, true
}

// ReceiptForMessage returns current receipt information.
func (q *Queue) ReceiptForMessage(messageID string) (Receipt, bool) {
	receipt, ok := q.receipts[messageID]
	if !ok {
		return Receipt{}, false
	}
	return receipt, true
}

// ShouldRetry reports whether retry window is open for a message.
func (q *Queue) ShouldRetry(now time.Time, messageID string) (bool, RetryReason) {
	receipt, ok := q.receipts[messageID]
	if !ok {
		return false, RetryReasonUnknownMessage
	}
	if receipt.Status == ReceiptStatusDelivered {
		return false, RetryReasonAcknowledged
	}
	if receipt.Status == ReceiptStatusExpired {
		return false, RetryReasonWindowClosed
	}
	if receipt.RetryAfter.IsZero() || !now.Before(receipt.RetryAfter) {
		return true, RetryReasonWindowOpen
	}
	return false, RetryReasonWindowClosed
}

// ShouldRunGC indicates whether GC should execute at current time.
func (q *Queue) ShouldRunGC(now time.Time) bool {
	if q.lastGC.IsZero() {
		return true
	}
	return now.Sub(q.lastGC) >= q.cfg.GCInterval
}

func (q *Queue) compact(now time.Time) {
	if len(q.items) == 0 {
		if q.ShouldRunGC(now) {
			q.lastGC = now.UTC()
		}
		return
	}
	sort.SliceStable(q.items, func(i, j int) bool {
		if q.items[i].enqueuedAt.Equal(q.items[j].enqueuedAt) {
			return q.items[i].seq < q.items[j].seq
		}
		return q.items[i].enqueuedAt.Before(q.items[j].enqueuedAt)
	})
	cut := 0
	for cut < len(q.items) {
		if now.Sub(q.items[cut].enqueuedAt) < q.cfg.TTL {
			break
		}
		delete(q.seenID, q.items[cut].ID)
		delete(q.acked, q.items[cut].ID)
		receipt := q.receipts[q.items[cut].ID]
		receipt.MessageID = q.items[cut].ID
		receipt.Status = ReceiptStatusExpired
		receipt.RetryAfter = time.Time{}
		q.receipts[q.items[cut].ID] = receipt
		cut++
	}
	if cut > 0 {
		q.items = append([]queuedMessage(nil), q.items[cut:]...)
	}
	if q.ShouldRunGC(now) {
		q.lastGC = now.UTC()
	}
}

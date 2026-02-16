package dmqueue

import (
	"testing"
	"time"
)

func TestQueueEnqueueDequeueAck(t *testing.T) {
	q, err := NewQueue(QueueConfig{MaxItems: 10, TTL: time.Minute})
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}
	now := time.Unix(100, 0).UTC()
	if _, err := q.Enqueue(now, Message{ID: "b", RecipientID: "alice", Ciphertext: []byte{2}}); err != nil {
		t.Fatalf("enqueue b: %v", err)
	}
	if _, err := q.Enqueue(now, Message{ID: "a", RecipientID: "alice", Ciphertext: []byte{1}}); err != nil {
		t.Fatalf("enqueue a: %v", err)
	}

	out := q.DequeueRecipient(now, "alice", 10)
	if len(out) != 2 {
		t.Fatalf("len(out)=%d want 2", len(out))
	}
	if out[0].ID != "b" || out[1].ID != "a" {
		t.Fatalf("order=%q,%q want b,a", out[0].ID, out[1].ID)
	}
	if !q.Ack("a") {
		t.Fatal("expected ack for a")
	}
	if q.Ack("a") {
		t.Fatal("expected idempotent ack false on second call")
	}
}

func TestQueueRejectsExpiredAndInvalidMessage(t *testing.T) {
	q, err := NewQueue(QueueConfig{MaxItems: 1, TTL: time.Second})
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}
	now := time.Unix(100, 0).UTC()
	if _, err := q.Enqueue(now, Message{}); err == nil {
		t.Fatal("expected invalid message error")
	}
	if _, err := q.Enqueue(now, Message{ID: "x", RecipientID: "bob", Ciphertext: []byte{1}}); err != nil {
		t.Fatalf("enqueue x: %v", err)
	}
	out := q.DequeueRecipient(now.Add(2*time.Second), "bob", 10)
	if len(out) != 0 {
		t.Fatalf("len(out)=%d want 0", len(out))
	}
	receipt, ok := q.ReceiptForMessage("x")
	if !ok || receipt.Status != ReceiptStatusExpired {
		t.Fatalf("expected expired receipt, got ok=%v receipt=%+v", ok, receipt)
	}
}

func TestMailboxAddressAndReplicationTargets(t *testing.T) {
	addrA, err := DeriveMailboxAddress("alice", "offline")
	if err != nil {
		t.Fatalf("derive mailbox address: %v", err)
	}
	addrB, err := DeriveMailboxAddress("alice", "offline")
	if err != nil {
		t.Fatalf("derive mailbox address second: %v", err)
	}
	if addrA.Digest != addrB.Digest {
		t.Fatalf("expected deterministic digest got %q and %q", addrA.Digest, addrB.Digest)
	}
	targets := ReplicationTargets(addrA, 3)
	if len(targets) != 3 {
		t.Fatalf("len(targets)=%d want 3", len(targets))
	}
	if targets[0] == targets[1] {
		t.Fatalf("expected unique target suffixes: %+v", targets)
	}
}

func TestRetryWindowAndDeliveryReceipts(t *testing.T) {
	q, err := NewQueue(QueueConfig{MaxItems: 10, TTL: time.Minute, RetryWindow: 5 * time.Second})
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}
	now := time.Unix(100, 0).UTC()
	if _, err := q.Enqueue(now, Message{ID: "r1", RecipientID: "alice", Ciphertext: []byte{1}}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	receipt, ok := q.MarkDeliveryAttempt(now, "r1", false)
	if !ok || receipt.Status != ReceiptStatusRetryScheduled || receipt.Attempts == 0 {
		t.Fatalf("expected retry-scheduled receipt, got ok=%v receipt=%+v", ok, receipt)
	}

	retry, reason := q.ShouldRetry(now.Add(2*time.Second), "r1")
	if retry || reason != RetryReasonWindowClosed {
		t.Fatalf("expected retry window closed, got retry=%v reason=%q", retry, reason)
	}

	retry, reason = q.ShouldRetry(now.Add(6*time.Second), "r1")
	if !retry || reason != RetryReasonWindowOpen {
		t.Fatalf("expected retry window open, got retry=%v reason=%q", retry, reason)
	}

	receipt, ok = q.MarkDeliveryAttempt(now.Add(6*time.Second), "r1", true)
	if !ok || receipt.Status != ReceiptStatusDelivered {
		t.Fatalf("expected delivered receipt, got ok=%v receipt=%+v", ok, receipt)
	}

	retry, reason = q.ShouldRetry(now.Add(7*time.Second), "r1")
	if retry || reason != RetryReasonAcknowledged {
		t.Fatalf("expected acknowledged no-retry, got retry=%v reason=%q", retry, reason)
	}

	retry, reason = q.ShouldRetry(now, "unknown")
	if retry || reason != RetryReasonUnknownMessage {
		t.Fatalf("expected unknown message no-retry, got retry=%v reason=%q", retry, reason)
	}
}

func TestShouldRunGC(t *testing.T) {
	q, err := NewQueue(QueueConfig{MaxItems: 10, TTL: time.Minute, GCInterval: 10 * time.Second})
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}
	now := time.Unix(100, 0).UTC()
	if !q.ShouldRunGC(now) {
		t.Fatal("expected gc on first check")
	}
	if _, err := q.Enqueue(now, Message{ID: "g1", RecipientID: "bob", Ciphertext: []byte{1}}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	_ = q.DequeueRecipient(now, "bob", 1)
	if q.ShouldRunGC(now.Add(5 * time.Second)) {
		t.Fatal("expected gc window not reached")
	}
	if !q.ShouldRunGC(now.Add(11 * time.Second)) {
		t.Fatal("expected gc window reached")
	}
}

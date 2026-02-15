package phase4

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestTopicFormatTopicForDeterministicNaming(t *testing.T) {
	format := TopicFormat{}
	topic, err := format.TopicFor("Srv-01", "General_Main")
	if err != nil {
		t.Fatalf("TopicFor failed: %v", err)
	}
	want := "aether/v0.1/server/srv-01/channel/general_main"
	if topic != want {
		t.Fatalf("unexpected topic: want %q got %q", want, topic)
	}

	custom := TopicFormat{Prefix: "aether/custom"}
	topic, err = custom.TopicFor("Srv-02", "Voice")
	if err != nil {
		t.Fatalf("TopicFor with custom prefix failed: %v", err)
	}
	if topic != "aether/custom/server/srv-02/channel/voice" {
		t.Fatalf("unexpected custom topic: %s", topic)
	}

	if _, err := format.TopicFor("", "general"); !errors.Is(err, ErrInvalidServerID) {
		t.Fatalf("expected invalid server id error, got %v", err)
	}
	if _, err := format.TopicFor("srv", "bad/id"); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected invalid channel id error, got %v", err)
	}
}

func TestPubSubServiceLifecycleAndEvents(t *testing.T) {
	svc := NewPubSubService(TopicFormat{}, nil, RetryPolicy{MaxAttempts: 2})
	topic, err := svc.Join("Alice", "Srv-A", "General")
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	if topic != "aether/v0.1/server/srv-a/channel/general" {
		t.Fatalf("unexpected join topic: %s", topic)
	}
	if _, joinErr := svc.Join("Bob", "Srv-A", "General"); joinErr != nil {
		t.Fatalf("second join failed: %v", joinErr)
	}

	topics, err := svc.SubscribedTopics("alice")
	if err != nil {
		t.Fatalf("SubscribedTopics failed: %v", err)
	}
	if !reflect.DeepEqual(topics, []string{"aether/v0.1/server/srv-a/channel/general"}) {
		t.Fatalf("unexpected subscribed topics: %v", topics)
	}

	if _, err := svc.Leave("alice", "srv-a", "general"); err != nil {
		t.Fatalf("leave failed: %v", err)
	}
	events := svc.Events()
	if len(events) != 3 {
		t.Fatalf("expected 3 lifecycle events, got %d", len(events))
	}
	if events[0].Type != SubscriptionEventJoin || events[0].Remaining != 1 {
		t.Fatalf("unexpected first event: %+v", events[0])
	}
	if events[2].Type != SubscriptionEventLeave || events[2].Remaining != 1 {
		t.Fatalf("unexpected leave event: %+v", events[2])
	}
}

func TestPubSubServicePublishAdmissionAndRetry(t *testing.T) {
	svc := NewPubSubService(TopicFormat{}, func(topic string, payload []byte) error {
		if strings.Contains(topic, "blocked") {
			return errors.New("blocked topic")
		}
		if len(payload) == 0 {
			return errors.New("empty payload")
		}
		return nil
	}, RetryPolicy{MaxAttempts: 3})

	attempts := 0
	count, err := svc.Publish("aether/v0.1/server/srv/channel/general", []byte("hello"), func(_ string, _ []byte) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary publish failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected publish success after retries, got %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 attempts, got %d", count)
	}

	_, err = svc.Publish("aether/v0.1/server/srv/channel/blocked", []byte("hello"), func(_ string, _ []byte) error {
		return nil
	})
	if !errors.Is(err, ErrAdmissionRejected) {
		t.Fatalf("expected admission rejection, got %v", err)
	}

	count, err = svc.Publish("aether/v0.1/server/srv/channel/general", []byte("hello"), func(_ string, _ []byte) error {
		return errors.New("always fail")
	})
	if !errors.Is(err, ErrRetryLimitExceeded) {
		t.Fatalf("expected retry limit exceeded, got %v", err)
	}
	if count != 3 {
		t.Fatalf("expected retry attempts to equal policy limit, got %d", count)
	}
}

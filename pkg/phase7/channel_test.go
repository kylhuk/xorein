package phase7

import (
	"errors"
	"reflect"
	"testing"
)

func TestTopicForDeterministicFormatting(t *testing.T) {
	topic, err := TopicFor("Srv-01", ChannelID("General_Main"))
	if err != nil {
		t.Fatalf("topic generation failed: %v", err)
	}
	want := "aether/v0.1/server/srv-01/channel/general_main"
	if topic != want {
		t.Fatalf("unexpected topic, want %q got %q", want, topic)
	}

	if _, err := TopicFor("", ChannelID("general")); !errors.Is(err, ErrInvalidServerID) {
		t.Fatalf("expected invalid server id error, got %v", err)
	}
	if _, err := TopicFor("server", ChannelID("bad/id")); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected invalid channel id error, got %v", err)
	}
}

func TestChannelModelJoinLeaveLifecycle(t *testing.T) {
	m, err := NewChannelModel("srvA")
	if err != nil {
		t.Fatalf("new model failed: %v", err)
	}

	binding, err := m.RegisterChannel(ChannelMetadata{
		ID:          ChannelID("General"),
		DisplayName: "General",
		CreatedBy:   ParticipantID("alice"),
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if binding.Topic != "aether/v0.1/server/srva/channel/general" {
		t.Fatalf("unexpected binding topic: %s", binding.Topic)
	}

	topics, err := m.Join(ParticipantID("alice"), ChannelID("General"))
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	if !reflect.DeepEqual(topics, []string{"aether/v0.1/server/srva/channel/general"}) {
		t.Fatalf("unexpected topics after join: %v", topics)
	}

	members, err := m.Members(ChannelID("general"))
	if err != nil {
		t.Fatalf("members failed: %v", err)
	}
	if !reflect.DeepEqual(members, []ParticipantID{"alice"}) {
		t.Fatalf("unexpected members: %v", members)
	}

	topics, err = m.Leave(ParticipantID("alice"), ChannelID("general"))
	if err != nil {
		t.Fatalf("leave failed: %v", err)
	}
	if len(topics) != 0 {
		t.Fatalf("expected empty topics after leave, got %v", topics)
	}
}

func TestChannelModelErrors(t *testing.T) {
	m, err := NewChannelModel("server")
	if err != nil {
		t.Fatalf("new model failed: %v", err)
	}

	if _, err := m.Join(ParticipantID("a"), ChannelID("missing")); !errors.Is(err, ErrUnknownChannel) {
		t.Fatalf("expected unknown channel on join, got %v", err)
	}
	if _, err := m.Leave(ParticipantID("a"), ChannelID("missing")); !errors.Is(err, ErrUnknownChannel) {
		t.Fatalf("expected unknown channel on leave, got %v", err)
	}
	if _, err := NewChannelModel("bad/server"); !errors.Is(err, ErrInvalidServerID) {
		t.Fatalf("expected invalid server id, got %v", err)
	}
}

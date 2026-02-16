package phase7

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalidServerID  = errors.New("phase7: invalid server id")
	ErrInvalidChannelID = errors.New("phase7: invalid channel id")
	ErrUnknownChannel   = errors.New("phase7: unknown channel")
)

type ChannelID string

type ChannelMetadata struct {
	ID          ChannelID
	DisplayName string
	CreatedBy   ParticipantID
}

type TopicBinding struct {
	ServerID  string
	ChannelID ChannelID
	Topic     string
}

type channelState struct {
	metadata ChannelMetadata
	topic    string
	members  map[ParticipantID]struct{}
}

type ChannelModel struct {
	mu            sync.Mutex
	serverID      string
	channels      map[ChannelID]*channelState
	subscriptions map[ParticipantID]map[ChannelID]struct{}
}

func NewChannelModel(serverID string) (*ChannelModel, error) {
	normalized, err := normalizeIdentifier(serverID)
	if err != nil {
		return nil, ErrInvalidServerID
	}
	return &ChannelModel{
		serverID:      normalized,
		channels:      make(map[ChannelID]*channelState),
		subscriptions: make(map[ParticipantID]map[ChannelID]struct{}),
	}, nil
}

func TopicFor(serverID string, channelID ChannelID) (string, error) {
	normalizedServerID, err := normalizeIdentifier(serverID)
	if err != nil {
		return "", ErrInvalidServerID
	}
	normalizedChannelID, err := normalizeIdentifier(string(channelID))
	if err != nil {
		return "", ErrInvalidChannelID
	}
	return fmt.Sprintf("aether/v0.1/server/%s/channel/%s", normalizedServerID, normalizedChannelID), nil
}

func (m *ChannelModel) RegisterChannel(metadata ChannelMetadata) (TopicBinding, error) {
	normalizedID, err := normalizeIdentifier(string(metadata.ID))
	if err != nil {
		return TopicBinding{}, ErrInvalidChannelID
	}
	channelID := ChannelID(normalizedID)
	topic, err := TopicFor(m.serverID, channelID)
	if err != nil {
		return TopicBinding{}, err
	}
	state := &channelState{
		metadata: ChannelMetadata{
			ID:          channelID,
			DisplayName: strings.TrimSpace(metadata.DisplayName),
			CreatedBy:   metadata.CreatedBy,
		},
		topic:   topic,
		members: make(map[ParticipantID]struct{}),
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[channelID] = state

	return TopicBinding{ServerID: m.serverID, ChannelID: channelID, Topic: topic}, nil
}

func (m *ChannelModel) Join(participant ParticipantID, channelID ChannelID) ([]string, error) {
	normalizedID, err := normalizeIdentifier(string(channelID))
	if err != nil {
		return nil, ErrInvalidChannelID
	}
	id := ChannelID(normalizedID)

	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.channels[id]
	if !ok {
		return nil, ErrUnknownChannel
	}

	state.members[participant] = struct{}{}
	if _, ok := m.subscriptions[participant]; !ok {
		m.subscriptions[participant] = make(map[ChannelID]struct{})
	}
	m.subscriptions[participant][id] = struct{}{}

	return m.subscriptionTopicsLocked(participant), nil
}

func (m *ChannelModel) Leave(participant ParticipantID, channelID ChannelID) ([]string, error) {
	normalizedID, err := normalizeIdentifier(string(channelID))
	if err != nil {
		return nil, ErrInvalidChannelID
	}
	id := ChannelID(normalizedID)

	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.channels[id]
	if !ok {
		return nil, ErrUnknownChannel
	}

	delete(state.members, participant)
	if channels, ok := m.subscriptions[participant]; ok {
		delete(channels, id)
		if len(channels) == 0 {
			delete(m.subscriptions, participant)
		}
	}

	return m.subscriptionTopicsLocked(participant), nil
}

func (m *ChannelModel) Subscriptions(participant ParticipantID) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.subscriptionTopicsLocked(participant)
}

func (m *ChannelModel) Members(channelID ChannelID) ([]ParticipantID, error) {
	normalizedID, err := normalizeIdentifier(string(channelID))
	if err != nil {
		return nil, ErrInvalidChannelID
	}
	id := ChannelID(normalizedID)

	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.channels[id]
	if !ok {
		return nil, ErrUnknownChannel
	}

	members := make([]ParticipantID, 0, len(state.members))
	for member := range state.members {
		members = append(members, member)
	}
	sort.Slice(members, func(i, j int) bool { return members[i] < members[j] })
	return members, nil
}

func (m *ChannelModel) HasMember(participant ParticipantID, channelID ChannelID) (bool, error) {
	normalizedID, err := normalizeIdentifier(string(channelID))
	if err != nil {
		return false, ErrInvalidChannelID
	}
	id := ChannelID(normalizedID)

	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.channels[id]
	if !ok {
		return false, ErrUnknownChannel
	}
	_, exists := state.members[participant]
	return exists, nil
}

func (m *ChannelModel) subscriptionTopicsLocked(participant ParticipantID) []string {
	channels, ok := m.subscriptions[participant]
	if !ok || len(channels) == 0 {
		return nil
	}
	topics := make([]string, 0, len(channels))
	for channelID := range channels {
		if state, exists := m.channels[channelID]; exists {
			topics = append(topics, state.topic)
		}
	}
	sort.Strings(topics)
	return topics
}

func normalizeIdentifier(raw string) (string, error) {
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

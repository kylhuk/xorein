package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const currentSchemaVersion = 2

type persistedState struct {
	SchemaVersion int                       `json:"schema_version"`
	Identity      Identity                  `json:"identity"`
	ControlToken  string                    `json:"control_token"`
	KnownPeers    map[string]PeerRecord     `json:"known_peers"`
	Servers       map[string]ServerRecord   `json:"servers"`
	DMs           map[string]DMRecord       `json:"dms"`
	Messages      map[string]MessageRecord  `json:"messages"`
	Deliveries    map[string]struct{}       `json:"deliveries"`
	Voice         map[string]VoiceSession   `json:"voice"`
	RelayQueues   map[string][]Delivery     `json:"relay_queues"`
	Settings      map[string]string         `json:"settings"`
	Telemetry     []string                  `json:"telemetry"`
}

type legacyStateV1 struct {
	Identity     Identity                 `json:"identity"`
	KnownPeers   map[string]PeerRecord    `json:"known_peers"`
	Servers      map[string]ServerRecord  `json:"servers"`
	DMs          map[string]DMRecord      `json:"dms"`
	Messages     map[string]MessageRecord `json:"messages"`
	Settings     map[string]string        `json:"settings"`
	ControlToken string                   `json:"control_token"`
}

func defaultState() (persistedState, error) {
	identity, err := GenerateIdentity("xorein")
	if err != nil {
		return persistedState{}, err
	}
	return persistedState{
		SchemaVersion: currentSchemaVersion,
		Identity:      identity,
		ControlToken:  randomID("control"),
		KnownPeers:    map[string]PeerRecord{},
		Servers:       map[string]ServerRecord{},
		DMs:           map[string]DMRecord{},
		Messages:      map[string]MessageRecord{},
		Deliveries:    map[string]struct{}{},
		Voice:         map[string]VoiceSession{},
		RelayQueues:   map[string][]Delivery{},
		Settings:      map[string]string{},
		Telemetry:     []string{},
	}, nil
}

func loadState(dataDir string) (persistedState, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return persistedState{}, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(dataDir, "state.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			state, err := defaultState()
			if err != nil {
				return persistedState{}, err
			}
			if err := saveState(dataDir, state); err != nil {
				return persistedState{}, err
			}
			return state, nil
		}
		return persistedState{}, fmt.Errorf("read state: %w", err)
	}
	var probe struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		if renameErr := quarantineCorruptState(path, raw); renameErr != nil {
			return persistedState{}, renameErr
		}
		state, err := defaultState()
		if err != nil {
			return persistedState{}, err
		}
		if err := saveState(dataDir, state); err != nil {
			return persistedState{}, err
		}
		return state, nil
	}
	if probe.SchemaVersion == 0 || probe.SchemaVersion == 1 {
		var legacy legacyStateV1
		if err := json.Unmarshal(raw, &legacy); err != nil {
			return persistedState{}, fmt.Errorf("decode legacy state: %w", err)
		}
		state := persistedState{
			SchemaVersion: currentSchemaVersion,
			Identity:      legacy.Identity,
			ControlToken:  legacy.ControlToken,
			KnownPeers:    clonePeerMap(legacy.KnownPeers),
			Servers:       cloneServerMap(legacy.Servers),
			DMs:           cloneDMMap(legacy.DMs),
			Messages:      cloneMessageMap(legacy.Messages),
			Deliveries:    map[string]struct{}{},
			Voice:         map[string]VoiceSession{},
			RelayQueues:   map[string][]Delivery{},
			Settings:      cloneStringMap(legacy.Settings),
			Telemetry:     []string{"state-migrated:v1->v2"},
		}
		if state.ControlToken == "" {
			state.ControlToken = randomID("control")
		}
		if err := ensureState(&state); err != nil {
			return persistedState{}, err
		}
		if err := saveState(dataDir, state); err != nil {
			return persistedState{}, err
		}
		return state, nil
	}
	var state persistedState
	if err := json.Unmarshal(raw, &state); err != nil {
		return persistedState{}, fmt.Errorf("decode state: %w", err)
	}
	if err := ensureState(&state); err != nil {
		return persistedState{}, err
	}
	return state, nil
}

func ensureState(state *persistedState) error {
	if state.SchemaVersion == 0 {
		state.SchemaVersion = currentSchemaVersion
	}
	if state.KnownPeers == nil {
		state.KnownPeers = map[string]PeerRecord{}
	}
	if state.Servers == nil {
		state.Servers = map[string]ServerRecord{}
	}
	if state.DMs == nil {
		state.DMs = map[string]DMRecord{}
	}
	if state.Messages == nil {
		state.Messages = map[string]MessageRecord{}
	}
	if state.Deliveries == nil {
		state.Deliveries = map[string]struct{}{}
	}
	if state.Voice == nil {
		state.Voice = map[string]VoiceSession{}
	}
	if state.RelayQueues == nil {
		state.RelayQueues = map[string][]Delivery{}
	}
	if state.Settings == nil {
		state.Settings = map[string]string{}
	}
	if state.Telemetry == nil {
		state.Telemetry = []string{}
	}
	if state.Identity.PeerID == "" {
		identity, err := GenerateIdentity("xorein")
		if err != nil {
			return err
		}
		state.Identity = identity
	}
	if err := state.Identity.Validate(); err != nil {
		return fmt.Errorf("state identity: %w", err)
	}
	if state.ControlToken == "" {
		state.ControlToken = randomID("control")
	}
	return nil
}

func saveState(dataDir string, state persistedState) error {
	if err := ensureState(&state); err != nil {
		return err
	}
	path := filepath.Join(dataDir, "state.json")
	tmp := path + ".tmp"
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write temp state: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}
	return nil
}

func quarantineCorruptState(path string, raw []byte) error {
	stamp := time.Now().UTC().Format("20060102T150405.000000000")
	corrupt := filepath.Join(filepath.Dir(path), "state.corrupt-"+stamp+".json")
	if len(raw) > 0 {
		if err := os.WriteFile(corrupt, raw, 0o600); err != nil {
			return fmt.Errorf("write corrupt state: %w", err)
		}
	}
	_ = os.Remove(path)
	return nil
}

func clonePeerMap(in map[string]PeerRecord) map[string]PeerRecord {
	out := make(map[string]PeerRecord, len(in))
	for k, v := range in {
		copy := v
		copy.Addresses = append([]string(nil), v.Addresses...)
		out[k] = copy
	}
	return out
}

func cloneServerMap(in map[string]ServerRecord) map[string]ServerRecord {
	out := make(map[string]ServerRecord, len(in))
	for k, v := range in {
		copy := v
		copy.Members = append([]string(nil), v.Members...)
		copy.Channels = cloneChannelMap(v.Channels)
		out[k] = copy
	}
	return out
}

func cloneChannelMap(in map[string]ChannelRecord) map[string]ChannelRecord {
	out := make(map[string]ChannelRecord, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneDMMap(in map[string]DMRecord) map[string]DMRecord {
	out := make(map[string]DMRecord, len(in))
	for k, v := range in {
		copy := v
		copy.Participants = append([]string(nil), v.Participants...)
		out[k] = copy
	}
	return out
}

func cloneMessageMap(in map[string]MessageRecord) map[string]MessageRecord {
	out := make(map[string]MessageRecord, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func sortedPeers(in map[string]PeerRecord) []PeerRecord {
	out := make([]PeerRecord, 0, len(in))
	for _, peer := range in {
		out = append(out, peer)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PeerID < out[j].PeerID })
	return out
}

func sortedServers(in map[string]ServerRecord) []ServerRecord {
	out := make([]ServerRecord, 0, len(in))
	for _, server := range in {
		out = append(out, server)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func sortedDMs(in map[string]DMRecord) []DMRecord {
	out := make([]DMRecord, 0, len(in))
	for _, dm := range in {
		out = append(out, dm)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func sortedMessages(in map[string]MessageRecord) []MessageRecord {
	out := make([]MessageRecord, 0, len(in))
	for _, msg := range in {
		out = append(out, msg)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func sortedVoice(in map[string]VoiceSession) []VoiceSession {
	out := make([]VoiceSession, 0, len(in))
	for _, session := range in {
		out = append(out, session)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

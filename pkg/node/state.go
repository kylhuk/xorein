package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/aether/code_aether/pkg/storage"
)

var saveStateFn = saveState

const currentSchemaVersion = 2

const (
	legacyStateFile         = "state.json"
	legacyBackupTag         = "state.migrated-"
	stateBucketIdentity     = "identity"
	stateBucketControlToken = "control_token"
	stateBucketKnownPeers   = "known_peers"
	stateBucketServers      = "servers"
	stateBucketDMs          = "dms"
	stateBucketMessages     = "messages"
	stateBucketDeliveries   = "deliveries"
	stateBucketVoice        = "voice"
	stateBucketRelayQueues  = "relay_queues"
	stateBucketSettings     = "settings"
	stateBucketTelemetry    = "telemetry"
)

type persistedState struct {
	SchemaVersion int                          `json:"schema_version"`
	Identity      Identity                     `json:"identity"`
	ControlToken  string                       `json:"control_token"`
	KnownPeers    map[string]PeerRecord        `json:"known_peers"`
	Servers       map[string]ServerRecord      `json:"servers"`
	DMs           map[string]DMRecord          `json:"dms"`
	Messages      map[string]MessageRecord     `json:"messages"`
	Deliveries    map[string]struct{}          `json:"deliveries"`
	Voice         map[string]VoiceSession      `json:"voice"`
	RelayQueues   map[string][]RelayQueueEntry `json:"relay_queues"`
	Settings      map[string]string            `json:"settings"`
	Telemetry     []string                     `json:"telemetry"`
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
		RelayQueues:   map[string][]RelayQueueEntry{},
		Settings:      map[string]string{},
		Telemetry:     []string{},
	}, nil
}

func loadState(dataDir string) (persistedState, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return persistedState{}, fmt.Errorf("create data dir: %w", err)
	}
	snapshot, err := storage.Load(dataDir)
	if err == nil {
		state, err := stateFromSnapshot(snapshot)
		if err != nil {
			return persistedState{}, err
		}
		if err := ensureState(&state); err != nil {
			return persistedState{}, err
		}
		return state, nil
	}
	if errors.Is(err, storage.ErrWrongKey) {
		return persistedState{}, fmt.Errorf("open state store: %w", err)
	}
	if errors.Is(err, storage.ErrCorrupt) {
		if qerr := quarantineCorruptStore(dataDir); qerr != nil {
			return persistedState{}, qerr
		}
		return recoverStateAfterStoreCorruption(dataDir)
	}
	if !errors.Is(err, storage.ErrStoreNotFound) {
		return persistedState{}, fmt.Errorf("load state store: %w", err)
	}
	path := filepath.Join(dataDir, legacyStateFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if backupState, backupErr := recoverStateFromLegacyBackup(dataDir); backupErr == nil {
				return backupState, nil
			} else if !errors.Is(backupErr, os.ErrNotExist) {
				return persistedState{}, fmt.Errorf("recover legacy backup: %w", backupErr)
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
			RelayQueues:   map[string][]RelayQueueEntry{},
			Settings:      cloneStringMap(legacy.Settings),
			Telemetry:     []string{"state-migrated:v1->v2"},
		}
		if state.ControlToken == "" {
			state.ControlToken = randomID("control")
		}
		if err := ensureState(&state); err != nil {
			return persistedState{}, err
		}
		if err := backupLegacyState(path); err != nil {
			return persistedState{}, err
		}
		if err := saveState(dataDir, state); err != nil {
			return persistedState{}, err
		}
		_ = os.Remove(path)
		return state, nil
	}
	var state persistedState
	if err := json.Unmarshal(raw, &state); err != nil {
		return persistedState{}, fmt.Errorf("decode state: %w", err)
	}
	if err := ensureState(&state); err != nil {
		return persistedState{}, err
	}
	if err := backupLegacyState(path); err != nil {
		return persistedState{}, err
	}
	if err := saveState(dataDir, state); err != nil {
		return persistedState{}, err
	}
	_ = os.Remove(path)
	return state, nil
}

func ensureState(state *persistedState) error {
	if state.SchemaVersion == 0 {
		state.SchemaVersion = currentSchemaVersion
	}
	if state.KnownPeers == nil {
		state.KnownPeers = map[string]PeerRecord{}
	} else {
		for peerID, peer := range state.KnownPeers {
			peer.Addresses = normalizePeerAddresses(peer.Addresses)
			state.KnownPeers[peerID] = peer
		}
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
		state.RelayQueues = map[string][]RelayQueueEntry{}
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
	snapshot, err := snapshotFromState(state)
	if err != nil {
		return err
	}
	return storage.Save(dataDir, snapshot)
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

func recoverStateAfterStoreCorruption(dataDir string) (persistedState, error) {
	if state, err := recoverStateFromLegacyBackup(dataDir); err == nil {
		return state, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return persistedState{}, fmt.Errorf("recover legacy backup: %w", err)
	}
	if state, err := loadStateFromLegacyFile(filepath.Join(dataDir, legacyStateFile)); err == nil {
		if err := backupLegacyState(filepath.Join(dataDir, legacyStateFile)); err != nil {
			return persistedState{}, err
		}
		if err := saveState(dataDir, state); err != nil {
			return persistedState{}, err
		}
		_ = os.Remove(filepath.Join(dataDir, legacyStateFile))
		return state, nil
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

func recoverStateFromLegacyBackup(dataDir string) (persistedState, error) {
	entries, err := filepath.Glob(filepath.Join(dataDir, legacyBackupTag+"*.json"))
	if err != nil {
		return persistedState{}, err
	}
	if len(entries) == 0 {
		return persistedState{}, os.ErrNotExist
	}
	sort.Strings(entries)
	state, err := loadStateFromLegacyFile(entries[len(entries)-1])
	if err != nil {
		return persistedState{}, err
	}
	if err := saveState(dataDir, state); err != nil {
		return persistedState{}, err
	}
	return state, nil
}

func loadStateFromLegacyFile(path string) (persistedState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return persistedState{}, err
	}
	var probe struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return persistedState{}, fmt.Errorf("decode legacy probe: %w", err)
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
			RelayQueues:   map[string][]RelayQueueEntry{},
			Settings:      cloneStringMap(legacy.Settings),
			Telemetry:     []string{"state-migrated:v1->v2"},
		}
		if err := ensureState(&state); err != nil {
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

func backupLegacyState(path string) error {
	if filepath.Base(path) != legacyStateFile {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read legacy state for backup: %w", err)
	}
	stamp := time.Now().UTC().Format("20060102T150405.000000000")
	backup := filepath.Join(filepath.Dir(path), legacyBackupTag+stamp+".json")
	if err := os.WriteFile(backup, raw, 0o600); err != nil {
		return fmt.Errorf("write migrated legacy backup: %w", err)
	}
	return nil
}

func quarantineCorruptStore(dataDir string) error {
	stamp := time.Now().UTC().Format("20060102T150405.000000000")
	for _, candidate := range []struct {
		path string
		name string
	}{
		{path: filepath.Join(dataDir, storage.StoreFileName), name: storage.StoreFileName},
		{path: filepath.Join(dataDir, storage.StoreMetaFileName), name: storage.StoreMetaFileName},
		{path: filepath.Join(dataDir, storage.StoreDirName), name: storage.StoreDirName},
	} {
		if _, err := os.Stat(candidate.path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("stat corrupt store: %w", err)
		}
		corrupt := filepath.Join(dataDir, candidate.name+".corrupt-"+stamp)
		if err := os.Rename(candidate.path, corrupt); err != nil {
			return fmt.Errorf("quarantine corrupt store: %w", err)
		}
		return nil
	}
	return nil
}

func snapshotFromState(state persistedState) (storage.Snapshot, error) {
	encode := func(v any) ([]byte, error) {
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return raw, nil
	}
	buckets := map[string][]byte{}
	var err error
	if buckets[stateBucketIdentity], err = encode(state.Identity); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode identity: %w", err)
	}
	if buckets[stateBucketControlToken], err = encode(state.ControlToken); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode control token: %w", err)
	}
	if buckets[stateBucketKnownPeers], err = encode(state.KnownPeers); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode known peers: %w", err)
	}
	if buckets[stateBucketServers], err = encode(state.Servers); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode servers: %w", err)
	}
	if buckets[stateBucketDMs], err = encode(state.DMs); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode dms: %w", err)
	}
	if buckets[stateBucketMessages], err = encode(state.Messages); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode messages: %w", err)
	}
	if buckets[stateBucketDeliveries], err = encode(state.Deliveries); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode deliveries: %w", err)
	}
	if buckets[stateBucketVoice], err = encode(state.Voice); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode voice: %w", err)
	}
	if buckets[stateBucketRelayQueues], err = encode(state.RelayQueues); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode relay queues: %w", err)
	}
	if buckets[stateBucketSettings], err = encode(state.Settings); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode settings: %w", err)
	}
	if buckets[stateBucketTelemetry], err = encode(state.Telemetry); err != nil {
		return storage.Snapshot{}, fmt.Errorf("encode telemetry: %w", err)
	}
	return storage.Snapshot{SchemaVersion: state.SchemaVersion, Buckets: buckets}, nil
}

func stateFromSnapshot(snapshot storage.Snapshot) (persistedState, error) {
	decode := func(name string, target any) error {
		raw, ok := snapshot.Buckets[name]
		if !ok || len(raw) == 0 {
			return nil
		}
		if err := json.Unmarshal(raw, target); err != nil {
			return fmt.Errorf("decode %s: %w", name, err)
		}
		return nil
	}
	state := persistedState{SchemaVersion: snapshot.SchemaVersion}
	if err := decode(stateBucketIdentity, &state.Identity); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketControlToken, &state.ControlToken); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketKnownPeers, &state.KnownPeers); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketServers, &state.Servers); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketDMs, &state.DMs); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketMessages, &state.Messages); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketDeliveries, &state.Deliveries); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketVoice, &state.Voice); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketRelayQueues, &state.RelayQueues); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketSettings, &state.Settings); err != nil {
		return persistedState{}, err
	}
	if err := decode(stateBucketTelemetry, &state.Telemetry); err != nil {
		return persistedState{}, err
	}
	return state, nil
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

func clonePersistedState(in persistedState) persistedState {
	return persistedState{
		SchemaVersion: in.SchemaVersion,
		Identity:      in.Identity,
		ControlToken:  in.ControlToken,
		KnownPeers:    clonePeerMap(in.KnownPeers),
		Servers:       cloneServerMap(in.Servers),
		DMs:           cloneDMMap(in.DMs),
		Messages:      cloneMessageMap(in.Messages),
		Deliveries:    cloneDeliverySet(in.Deliveries),
		Voice:         cloneVoiceMap(in.Voice),
		RelayQueues:   cloneRelayQueueMap(in.RelayQueues),
		Settings:      cloneStringMap(in.Settings),
		Telemetry:     append([]string(nil), in.Telemetry...),
	}
}

func cloneDeliverySet(in map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for k := range in {
		out[k] = struct{}{}
	}
	return out
}

func cloneVoiceMap(in map[string]VoiceSession) map[string]VoiceSession {
	out := make(map[string]VoiceSession, len(in))
	for k, v := range in {
		copy := v
		copy.Participants = cloneVoiceParticipants(v.Participants)
		copy.LastFrameBy = cloneTimeMap(v.LastFrameBy)
		out[k] = copy
	}
	return out
}

func cloneVoiceParticipants(in map[string]VoiceParticipant) map[string]VoiceParticipant {
	out := make(map[string]VoiceParticipant, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneTimeMap(in map[string]time.Time) map[string]time.Time {
	out := make(map[string]time.Time, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneRelayQueueMap(in map[string][]RelayQueueEntry) map[string][]RelayQueueEntry {
	out := make(map[string][]RelayQueueEntry, len(in))
	for k, v := range in {
		out[k] = append([]RelayQueueEntry(nil), v...)
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

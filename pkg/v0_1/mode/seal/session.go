package seal

import (
	"encoding/json"
	"fmt"
)

// SessionStore is the minimal interface used by seal to persist ratchet state.
// The caller provides a store-specific implementation.
type SessionStore interface {
	// GetSession returns the raw session JSON for the given session ID, or nil if absent.
	GetSession(sessionID string) ([]byte, error)
	// PutSession atomically writes the raw session JSON for the given session ID.
	PutSession(sessionID string, data []byte) error
	// DeleteSession removes session state for the given session ID.
	DeleteSession(sessionID string) error
}

// ratchetStateJSON is the JSON-serializable form of RatchetState.
// SkipList keys are encoded as "hex(ratchetPub):counter".
type ratchetStateJSON struct {
	RootKey          []byte                 `json:"root_key"`
	SendChainKey     []byte                 `json:"send_chain_key"`
	RecvChainKey     []byte                 `json:"recv_chain_key"`
	SendCounter      uint32                 `json:"send_counter"`
	RecvCounter      uint32                 `json:"recv_counter"`
	PrevSendChainLen uint32                 `json:"prev_send_chain_len"`
	SendRatchetPriv  []byte                 `json:"send_ratchet_priv"`
	SendRatchetPub   []byte                 `json:"send_ratchet_pub"`
	RemoteRatchetPub []byte                 `json:"remote_ratchet_pub"`
	SkipList         []skipListEntry        `json:"skip_list"`
}

type skipListEntry struct {
	RatchetPub []byte `json:"rp"`
	Counter    uint32 `json:"c"`
	MessageKey []byte `json:"mk"`
}

// LoadSession loads and deserializes a RatchetState from the store.
// Returns nil (not error) if no session exists for sessionID.
func LoadSession(store SessionStore, sessionID string) (*RatchetState, error) {
	data, err := store.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("load session %q: %w", sessionID, err)
	}
	if data == nil {
		return nil, nil
	}
	return deserializeSession(data)
}

// SaveSession serializes and persists a RatchetState.
func SaveSession(store SessionStore, sessionID string, s *RatchetState) error {
	data, err := serializeSession(s)
	if err != nil {
		return fmt.Errorf("save session %q: %w", sessionID, err)
	}
	return store.PutSession(sessionID, data)
}

// DeleteSession zeroes all key material in the session and removes it from the store.
// Per spec §9.2, session secrets MUST be zeroed before deletion.
func DeleteSession(store SessionStore, sessionID string) error {
	// Load and zero key material before deleting.
	data, err := store.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("delete session %q: load for zeroing: %w", sessionID, err)
	}
	if data != nil {
		s, deserErr := deserializeSession(data)
		if deserErr == nil && s != nil {
			zeroRatchetState(s)
		}
		// Zero the raw serialized bytes too.
		for i := range data {
			data[i] = 0
		}
	}
	return store.DeleteSession(sessionID)
}

// zeroRatchetState zeroes all secret key material in a RatchetState per spec §9.2.
func zeroRatchetState(s *RatchetState) {
	for i := range s.RootKey {
		s.RootKey[i] = 0
	}
	for i := range s.SendChainKey {
		s.SendChainKey[i] = 0
	}
	for i := range s.RecvChainKey {
		s.RecvChainKey[i] = 0
	}
	for i := range s.SendRatchetPriv {
		s.SendRatchetPriv[i] = 0
	}
	for i := range s.SendRatchetPub {
		s.SendRatchetPub[i] = 0
	}
	for i := range s.RemoteRatchetPub {
		s.RemoteRatchetPub[i] = 0
	}
	// Zero all skip-list message keys.
	for k, mk := range s.SkipList {
		for i := range mk {
			mk[i] = 0
		}
		s.SkipList[k] = mk
	}
}

func serializeSession(s *RatchetState) ([]byte, error) {
	j := ratchetStateJSON{
		RootKey:          s.RootKey[:],
		SendChainKey:     s.SendChainKey[:],
		RecvChainKey:     s.RecvChainKey[:],
		SendCounter:      s.SendCounter,
		RecvCounter:      s.RecvCounter,
		PrevSendChainLen: s.PrevSendChainLen,
		SendRatchetPriv:  s.SendRatchetPriv[:],
		SendRatchetPub:   s.SendRatchetPub[:],
		RemoteRatchetPub: s.RemoteRatchetPub[:],
	}
	for k, mk := range s.SkipList {
		mkCopy := mk
		j.SkipList = append(j.SkipList, skipListEntry{
			RatchetPub: k.RatchetPub[:],
			Counter:    k.Counter,
			MessageKey: mkCopy[:],
		})
	}
	return json.Marshal(j)
}

func deserializeSession(data []byte) (*RatchetState, error) {
	var j ratchetStateJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}
	s := &RatchetState{
		SendCounter:      j.SendCounter,
		RecvCounter:      j.RecvCounter,
		PrevSendChainLen: j.PrevSendChainLen,
		SkipList:         make(map[SkipKey][32]byte, len(j.SkipList)),
	}
	copy32 := func(dst *[32]byte, src []byte) {
		copy(dst[:], src)
	}
	copy32(&s.RootKey, j.RootKey)
	copy32(&s.SendChainKey, j.SendChainKey)
	copy32(&s.RecvChainKey, j.RecvChainKey)
	copy32(&s.SendRatchetPriv, j.SendRatchetPriv)
	copy32(&s.SendRatchetPub, j.SendRatchetPub)
	copy32(&s.RemoteRatchetPub, j.RemoteRatchetPub)
	for _, e := range j.SkipList {
		var rp [32]byte
		var mk [32]byte
		copy(rp[:], e.RatchetPub)
		copy(mk[:], e.MessageKey)
		s.SkipList[SkipKey{RatchetPub: rp, Counter: e.Counter}] = mk
	}
	return s, nil
}

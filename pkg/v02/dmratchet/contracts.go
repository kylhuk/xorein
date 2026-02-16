package dmratchet

const (
	DefaultMaxSkippedKeysPerChain = 128
	DefaultMaxTotalSkippedKeys    = 512
	DefaultReplayWindow           = 1024
)

type ChainRole string

const (
	ChainSending   ChainRole = "sending"
	ChainReceiving ChainRole = "receiving"
)

type RatchetConfig struct {
	MaxSkippedKeysPerChain uint32
	MaxTotalSkippedKeys    uint32
	ReplayWindow           uint32
}

func DefaultConfig() RatchetConfig {
	return RatchetConfig{
		MaxSkippedKeysPerChain: DefaultMaxSkippedKeysPerChain,
		MaxTotalSkippedKeys:    DefaultMaxTotalSkippedKeys,
		ReplayWindow:           DefaultReplayWindow,
	}
}

type BindingExpectation struct {
	SessionID string
	BindingID string
	ModeEpoch string
}

type RatchetState struct {
	SessionID        string
	PeerID           string
	BindingID        string
	ModeEpoch        string
	RootKey          []byte
	SendingCounter   uint32
	ReceivingCounter uint32
	PreviousChainLen uint32
	SkippedPerChain  map[ChainRole]uint32
	TotalSkippedKeys uint32
	ReplayCache      map[string]struct{}
}

type PersistenceDecision struct {
	Valid          bool
	ResyncRequired bool
	Reason         ValidationReason
}

type ValidationReason string

const (
	ReasonStateValid                     ValidationReason = "state-valid"
	ReasonMissingRootKey                 ValidationReason = "missing-root-key"
	ReasonMissingSessionID               ValidationReason = "missing-session-id"
	ReasonSessionBindingMismatch         ValidationReason = "session-binding-mismatch"
	ReasonModeEpochMismatch              ValidationReason = "mode-epoch-mismatch"
	ReasonSkippedKeysExceeded            ValidationReason = "skipped-keys-exceeded"
	ReasonValidationReplayWindowExceeded ValidationReason = "replay-window-exceeded"
)

func ValidatePersistedState(state RatchetState, expectation BindingExpectation, config RatchetConfig) PersistenceDecision {
	config = normalizeConfig(config)
	if len(state.RootKey) == 0 {
		return PersistenceDecision{Reason: ReasonMissingRootKey, ResyncRequired: true}
	}
	if state.SessionID == "" {
		return PersistenceDecision{Reason: ReasonMissingSessionID, ResyncRequired: true}
	}
	if expectation.SessionID != "" && state.SessionID != expectation.SessionID {
		return PersistenceDecision{Reason: ReasonSessionBindingMismatch, ResyncRequired: true}
	}
	if expectation.BindingID != "" && state.BindingID != expectation.BindingID {
		return PersistenceDecision{Reason: ReasonSessionBindingMismatch, ResyncRequired: true}
	}
	if expectation.ModeEpoch != "" && state.ModeEpoch != expectation.ModeEpoch {
		return PersistenceDecision{Reason: ReasonModeEpochMismatch, ResyncRequired: true}
	}
	if config.MaxTotalSkippedKeys > 0 && state.TotalSkippedKeys > config.MaxTotalSkippedKeys {
		return PersistenceDecision{Reason: ReasonSkippedKeysExceeded, ResyncRequired: true}
	}
	for _, count := range state.SkippedPerChain {
		if config.MaxSkippedKeysPerChain > 0 && count > config.MaxSkippedKeysPerChain {
			return PersistenceDecision{Reason: ReasonSkippedKeysExceeded, ResyncRequired: true}
		}
	}
	if config.ReplayWindow > 0 && uint32(len(state.ReplayCache)) > config.ReplayWindow {
		return PersistenceDecision{Reason: ReasonValidationReplayWindowExceeded, ResyncRequired: true}
	}
	return PersistenceDecision{Valid: true, Reason: ReasonStateValid}
}

type IncomingMessage struct {
	MessageID string
	Chain     ChainRole
	Counter   uint32
}

type MessageDecision struct {
	Accept          bool
	Replay          bool
	Resync          bool
	Reason          DecisionReason
	Gap             uint32
	NewSkippedTotal uint32
	SkipChain       ChainRole
}

type DecisionReason string

const (
	ReasonInOrderMessage               DecisionReason = "message-in-order"
	ReasonOutOfOrderWithinBounds       DecisionReason = "message-out-of-order"
	ReasonDuplicateMessage             DecisionReason = "duplicate-ciphertext"
	ReasonReplayDetected               DecisionReason = "replay-detected"
	ReasonSkipWindowExceeded           DecisionReason = "skip-window-exceeded"
	ReasonTotalSkipBudgetExceeded      DecisionReason = "total-skip-budget-exceeded"
	ReasonDecisionReplayWindowExceeded DecisionReason = "replay-window-exceeded"
	ReasonMessageIDMissing             DecisionReason = "missing-message-id"
	ReasonCounterBehind                DecisionReason = "counter-behind"
)

func EvaluateIncomingMessage(state RatchetState, message IncomingMessage, config RatchetConfig) MessageDecision {
	config = normalizeConfig(config)
	if message.MessageID == "" {
		return MessageDecision{Reason: ReasonMessageIDMissing}
	}
	if state.ReplayCache != nil {
		if _, ok := state.ReplayCache[message.MessageID]; ok {
			return MessageDecision{Reason: ReasonReplayDetected, Replay: true}
		}
	}
	if config.ReplayWindow > 0 && uint32(len(state.ReplayCache)) >= config.ReplayWindow {
		return MessageDecision{Reason: ReasonDecisionReplayWindowExceeded, Resync: true}
	}
	if message.Counter < state.ReceivingCounter {
		return MessageDecision{Reason: ReasonCounterBehind}
	}
	if message.Counter == state.ReceivingCounter {
		return MessageDecision{Accept: true, Reason: ReasonInOrderMessage}
	}
	gap := message.Counter - state.ReceivingCounter
	if config.MaxSkippedKeysPerChain > 0 && gap > config.MaxSkippedKeysPerChain {
		return MessageDecision{Reason: ReasonSkipWindowExceeded, Resync: true}
	}
	total := uint64(state.TotalSkippedKeys) + uint64(gap)
	if config.MaxTotalSkippedKeys > 0 && total > uint64(config.MaxTotalSkippedKeys) {
		return MessageDecision{Reason: ReasonTotalSkipBudgetExceeded, Resync: true}
	}
	return MessageDecision{
		Accept:          true,
		Reason:          ReasonOutOfOrderWithinBounds,
		Gap:             gap,
		NewSkippedTotal: uint32(total),
		SkipChain:       message.Chain,
	}
}

func normalizeConfig(cfg RatchetConfig) RatchetConfig {
	if cfg.MaxSkippedKeysPerChain == 0 {
		cfg.MaxSkippedKeysPerChain = DefaultMaxSkippedKeysPerChain
	}
	if cfg.MaxTotalSkippedKeys == 0 {
		cfg.MaxTotalSkippedKeys = DefaultMaxTotalSkippedKeys
	}
	if cfg.ReplayWindow == 0 {
		cfg.ReplayWindow = DefaultReplayWindow
	}
	return cfg
}

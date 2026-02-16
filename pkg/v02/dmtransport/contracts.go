package dmtransport

// SecurityMode is the v0.2 conversation mode label.
type SecurityMode string

const (
	SecurityModeSeal  SecurityMode = "seal"
	SecurityModeTree  SecurityMode = "tree"
	SecurityModeClear SecurityMode = "clear"
)

// DeliveryPath is the selected DM transport path.
type DeliveryPath string

const (
	DeliveryPathDirect  DeliveryPath = "direct"
	DeliveryPathOffline DeliveryPath = "offline"
	DeliveryPathReject  DeliveryPath = "reject"
)

// ReasonCode is the deterministic path-selection reason taxonomy.
type ReasonCode string

const (
	ReasonDirectAvailable            ReasonCode = "direct-available"
	ReasonPeerOffline                ReasonCode = "peer-offline"
	ReasonDirectUnsupported          ReasonCode = "direct-unsupported"
	ReasonUnsupportedSecurityMode    ReasonCode = "unsupported-security-mode"
	ReasonSessionOpenRequested       ReasonCode = "direct-session-open-requested"
	ReasonSessionOpened              ReasonCode = "direct-session-opened"
	ReasonSessionClosed              ReasonCode = "direct-session-closed"
	ReasonSessionTimedOut            ReasonCode = "direct-session-timeout"
	ReasonSessionRetryScheduled      ReasonCode = "direct-session-retry-scheduled"
	ReasonSessionNoTransition        ReasonCode = "direct-session-no-transition"
	ReasonInvalidSessionBinding      ReasonCode = "invalid-session-binding"
	ReasonMissingIntegrityTag        ReasonCode = "missing-integrity-tag"
	ReasonMetadataMinimizationFailed ReasonCode = "metadata-minimization-failed"
	ReasonIntegrityVerified          ReasonCode = "integrity-verified"
)

// TransportDecision reports selected path and reason.
type TransportDecision struct {
	Path   DeliveryPath
	Reason ReasonCode
}

// SelectTransportPath picks direct/offline/reject behavior deterministically.
func SelectTransportPath(peerOnline bool, directCapable bool, mode SecurityMode) TransportDecision {
	if mode != SecurityModeSeal && mode != SecurityModeTree {
		return TransportDecision{Path: DeliveryPathReject, Reason: ReasonUnsupportedSecurityMode}
	}
	if !peerOnline {
		return TransportDecision{Path: DeliveryPathOffline, Reason: ReasonPeerOffline}
	}
	if !directCapable {
		return TransportDecision{Path: DeliveryPathOffline, Reason: ReasonDirectUnsupported}
	}
	return TransportDecision{Path: DeliveryPathDirect, Reason: ReasonDirectAvailable}
}

// SessionState captures deterministic direct-session lifecycle states.
type SessionState string

const (
	SessionStateClosed   SessionState = "closed"
	SessionStateOpening  SessionState = "opening"
	SessionStateOpen     SessionState = "open"
	SessionStateRetrying SessionState = "retrying"
)

// SessionEvent drives deterministic direct-session state transitions.
type SessionEvent string

const (
	SessionEventOpenRequested SessionEvent = "open-requested"
	SessionEventOpenConfirmed SessionEvent = "open-confirmed"
	SessionEventClose         SessionEvent = "close"
	SessionEventTimeout       SessionEvent = "timeout"
	SessionEventRetryTick     SessionEvent = "retry-tick"
)

// SessionBinding binds a direct stream to conversation and peer identity.
type SessionBinding struct {
	SessionID      string
	ConversationID string
	PeerID         string
}

// SessionDecision reports next lifecycle state and reason.
type SessionDecision struct {
	State  SessionState
	Reason ReasonCode
}

// TransitionSession deterministically updates direct-session lifecycle state.
func TransitionSession(current SessionState, event SessionEvent, binding SessionBinding) SessionDecision {
	requiresBinding := event == SessionEventOpenRequested || event == SessionEventOpenConfirmed || event == SessionEventRetryTick
	if requiresBinding && !binding.Valid() {
		return SessionDecision{State: SessionStateClosed, Reason: ReasonInvalidSessionBinding}
	}

	switch current {
	case SessionStateClosed:
		if event == SessionEventOpenRequested {
			return SessionDecision{State: SessionStateOpening, Reason: ReasonSessionOpenRequested}
		}
	case SessionStateOpening:
		switch event {
		case SessionEventOpenConfirmed:
			return SessionDecision{State: SessionStateOpen, Reason: ReasonSessionOpened}
		case SessionEventTimeout:
			return SessionDecision{State: SessionStateRetrying, Reason: ReasonSessionTimedOut}
		case SessionEventClose:
			return SessionDecision{State: SessionStateClosed, Reason: ReasonSessionClosed}
		}
	case SessionStateOpen:
		switch event {
		case SessionEventTimeout:
			return SessionDecision{State: SessionStateRetrying, Reason: ReasonSessionTimedOut}
		case SessionEventClose:
			return SessionDecision{State: SessionStateClosed, Reason: ReasonSessionClosed}
		}
	case SessionStateRetrying:
		switch event {
		case SessionEventRetryTick:
			return SessionDecision{State: SessionStateOpening, Reason: ReasonSessionRetryScheduled}
		case SessionEventClose:
			return SessionDecision{State: SessionStateClosed, Reason: ReasonSessionClosed}
		}
	}

	if current == "" {
		current = SessionStateClosed
	}
	return SessionDecision{State: current, Reason: ReasonSessionNoTransition}
}

// Valid indicates whether session-binding requirements are satisfied.
func (b SessionBinding) Valid() bool {
	return b.SessionID != "" && b.ConversationID != "" && b.PeerID != ""
}

// EnvelopeMetadata is the minimal metadata contract required for integrity checks.
type EnvelopeMetadata struct {
	MessageID            string
	ConversationID       string
	SenderID             string
	RecipientID          string
	SessionID            string
	SecurityMode         SecurityMode
	ModeEpochID          string
	IntegrityTag         []byte
	AdditionalFieldCount int
}

// EnvelopeDecision reports whether envelope metadata passed integrity policies.
type EnvelopeDecision struct {
	Accepted bool
	Reason   ReasonCode
}

// ValidateEnvelopeMetadata enforces metadata minimization and integrity requirements.
func ValidateEnvelopeMetadata(metadata EnvelopeMetadata) EnvelopeDecision {
	if metadata.SecurityMode != SecurityModeSeal && metadata.SecurityMode != SecurityModeTree {
		return EnvelopeDecision{Accepted: false, Reason: ReasonUnsupportedSecurityMode}
	}
	if metadata.SessionID == "" || metadata.ConversationID == "" || metadata.SenderID == "" || metadata.RecipientID == "" {
		return EnvelopeDecision{Accepted: false, Reason: ReasonInvalidSessionBinding}
	}
	if metadata.MessageID == "" || metadata.ModeEpochID == "" {
		return EnvelopeDecision{Accepted: false, Reason: ReasonMetadataMinimizationFailed}
	}
	if metadata.AdditionalFieldCount > 0 {
		return EnvelopeDecision{Accepted: false, Reason: ReasonMetadataMinimizationFailed}
	}
	if len(metadata.IntegrityTag) == 0 {
		return EnvelopeDecision{Accepted: false, Reason: ReasonMissingIntegrityTag}
	}
	return EnvelopeDecision{Accepted: true, Reason: ReasonIntegrityVerified}
}

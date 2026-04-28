// Package protocol implements v0.1 protocol registry, capability negotiation,
// and error taxonomy per docs/spec/v0.1/03-protocol-registry-and-negotiation.md.
package protocol

// PeerStream error codes per spec 03 §9 + spec 02 §1.3.
const (
	CodeMissingRequiredCapability = "MISSING_REQUIRED_CAPABILITY"
	CodeUnsupportedVersion        = "UNSUPPORTED_VERSION"
	CodeUnsupportedOperation      = "UNSUPPORTED_OPERATION"
	CodeSignatureMismatch         = "SIGNATURE_MISMATCH"
	CodeModeIncompatible          = "MODE_INCOMPATIBLE"
	CodeRelayOpacityViolation     = "RELAY_OPACITY_VIOLATION"
	CodeOperationFailed           = "OPERATION_FAILED"
	CodeRateLimited               = "RATE_LIMITED"
	CodeReplayDetected            = "REPLAY_DETECTED"
	CodeExpiredSignature          = "EXPIRED_SIGNATURE"

	// Moderation family error codes (spec 50 §7).
	CodeModerationUnauthorized     = "MODERATION_UNAUTHORIZED"
	CodeModerationForbiddenTarget  = "MODERATION_FORBIDDEN_TARGET"
	CodeModerationMissingSignature = "MODERATION_MISSING_SIGNATURE"
	CodeModerationInvalidSigner    = "MODERATION_INVALID_SIGNER"
	CodeModerationInvalidSignature = "MODERATION_INVALID_SIGNATURE"
	CodeSlowModeInvalidDuration    = "SLOW_MODE_INVALID_DURATION"
	CodeBanListFull                = "BAN_LIST_FULL"
	CodeMessageNotFound            = "MESSAGE_NOT_FOUND"

	// Governance family error codes (spec 51 §7).
	CodeGovernanceUnauthorized      = "GOVERNANCE_UNAUTHORIZED"
	CodeGovernanceForbiddenTarget   = "GOVERNANCE_FORBIDDEN_TARGET"
	CodeGovernanceStaleVersion      = "GOVERNANCE_STALE_VERSION"
	CodeGovernanceRoleNotFound      = "GOVERNANCE_ROLE_NOT_FOUND"
	CodeGovernanceRoleConflict      = "GOVERNANCE_ROLE_CONFLICT"
	CodeGovernanceBaseRoleProtected = "GOVERNANCE_BASE_ROLE_PROTECTED"
	CodeGovernanceInvalidBitfield   = "GOVERNANCE_INVALID_BITFIELD"
	CodeGovernanceMissingSignature  = "GOVERNANCE_MISSING_SIGNATURE"
	CodeGovernanceInvalidSignature  = "GOVERNANCE_INVALID_SIGNATURE"
	CodeGovernanceOwnerImmutable    = "GOVERNANCE_OWNER_IMMUTABLE"

	// Sync family error codes (spec 44 §7).
	CodeSyncRangeTooLarge      = "SYNC_RANGE_TOO_LARGE"
	CodeSyncFetchLimitExceeded = "SYNC_FETCH_LIMIT_EXCEEDED"
	CodeSyncNotAMember         = "SYNC_NOT_A_MEMBER"
	CodeSyncServerNotFound     = "SYNC_SERVER_NOT_FOUND"
	CodeSyncArchivistRequired  = "SYNC_ARCHIVIST_REQUIRED"
	CodeSyncSignatureInvalid   = "SYNC_SIGNATURE_INVALID"

	// Voice family error codes (spec 52 §7).
	CodeVoiceCodecUnsupported     = "VOICE_CODEC_UNSUPPORTED"
	CodeVoiceSessionNotFound      = "VOICE_SESSION_NOT_FOUND"
	CodeVoiceNotAuthorized        = "VOICE_NOT_AUTHORIZED"
	CodeMediaShieldKeyUnavailable = "MEDIASHIELD_KEY_UNAVAILABLE"
	CodeVoiceSignalExpired        = "VOICE_SIGNAL_EXPIRED"
	CodeVoiceSignalReplay         = "VOICE_SIGNAL_REPLAY"
	CodeVoiceFrameTooLarge        = "VOICE_FRAME_TOO_LARGE"
	CodeVoiceSFUNotCoordinator    = "VOICE_SFU_NOT_COORDINATOR"

	// Chat family error codes (spec 41).
	CodeInviteExpired     = "INVITE_EXPIRED"
	CodeInviteInvalid     = "INVITE_INVALID"
	CodeNotAMember        = "NOT_A_MEMBER"
	CodeChannelNotFound   = "CHANNEL_NOT_FOUND"
	CodeDuplicateDelivery = "DUPLICATE_DELIVERY"

	// Manifest family error codes (spec 42).
	CodeManifestNotFound   = "MANIFEST_NOT_FOUND"
	CodeManifestNotNewer   = "NOT_NEWER"
	CodeManifestOwnerMismatch = "OWNER_MISMATCH"
	CodeSignatureInvalid   = "SIGNATURE_INVALID"

	// Identity family error codes (spec 43).
	CodeNoOPKAvailable = "NO_OPK_AVAILABLE"

	// DM family error codes (spec 45).
	CodeDMRateLimited = "DM_RATE_LIMITED"

	// Friends family additional error codes (spec 47).
	CodeSignatureMismatchFriends = "SIGNATURE_MISMATCH"
)

// NewPeerStreamError constructs a PeerStreamError for the canonical error code.
// missingCaps is only meaningful for CodeMissingRequiredCapability.
func NewPeerStreamError(code, message string, missingCaps []string) *PeerStreamError {
	return &PeerStreamError{
		Code:                code,
		Message:             message,
		MissingCapabilities: missingCaps,
	}
}

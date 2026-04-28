package envelope

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"time"

	apb "github.com/aether/code_aether/gen/go/proto"
	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

const (
	// LiveFreshnessWindow is the max age of a signed_at timestamp for live operations.
	LiveFreshnessWindow = 5 * time.Minute
	// StoredFreshnessWindow is the max age for stored objects (prekey bundles, manifests).
	StoredFreshnessWindow = 7 * 24 * time.Hour
)

// SignEnvelope builds a canonical payload and signs it with both Ed25519 and ML-DSA-65,
// returning a populated SignedEnvelope per spec 02 §2.2 + spec 01 §6.
//
// payloadType must be a PayloadType proto enum value.
// msg is the proto message carried in the envelope.
// edPriv is the Ed25519 private key (64 bytes).
// mldsaPriv is the ML-DSA-65 private key (4032 bytes).
// signer is the IdentityProfile of the signing node (must include both public keys).
func SignEnvelope(
	payloadType apb.PayloadType,
	msg proto.Message,
	edPriv ed25519.PrivateKey,
	mldsaPriv []byte,
	signer *apb.IdentityProfile,
) (*apb.SignedEnvelope, error) {
	now := time.Now()
	signedAtMS := now.UnixMilli()

	canonical, err := BuildCanonicalPayload(uint32(payloadType), msg, signedAtMS)
	if err != nil {
		return nil, err
	}

	edSig := v0crypto.SignEd25519(edPriv, canonical)
	mldsaSig, err := v0crypto.SignMLDSA65(mldsaPriv, canonical)
	if err != nil {
		return nil, fmt.Errorf("sign envelope: %w", err)
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("sign envelope: marshal payload: %w", err)
	}

	return &apb.SignedEnvelope{
		EnvelopeId:         uuid.New().String(),
		PayloadType:        payloadType,
		Payload:            payload,
		Signer:             signer,
		SignatureAlgorithm: apb.SignatureAlgorithm_SIGNATURE_ALGORITHM_HYBRID_ED25519_ML_DSA_65,
		Signature:          edSig,
		CanonicalPayload:   canonical,
		SignedAt:           uint64(signedAtMS),
		MlDsa65Signature:   mldsaSig,
	}, nil
}

// VerifyEnvelope verifies a SignedEnvelope per spec 02 §2.3.
// now is the caller's current time (allows tests to control the clock).
// stored=true uses the StoredFreshnessWindow; stored=false uses LiveFreshnessWindow.
//
// Returns one of the ErrXxx sentinel errors from this package, or nil on success.
func VerifyEnvelope(env *apb.SignedEnvelope, now time.Time, stored bool) error {
	// Step 1: shape check.
	if env == nil {
		return ErrUnsignedPayload
	}
	if len(env.Signature) == 0 {
		return ErrUnsignedPayload
	}

	// Step 2: algorithm.
	if env.SignatureAlgorithm != apb.SignatureAlgorithm_SIGNATURE_ALGORITHM_HYBRID_ED25519_ML_DSA_65 {
		return ErrUnsupportedPayloadType
	}

	// Step 3: recompute canonical payload and compare.
	msg, err := payloadForType(env)
	if err != nil {
		return ErrUnsupportedPayloadType
	}
	recomputed, err := BuildCanonicalPayload(uint32(env.PayloadType), msg, int64(env.SignedAt))
	if err != nil {
		return ErrCanonicalizationMismatch
	}
	if !bytes.Equal(recomputed, env.CanonicalPayload) {
		return ErrCanonicalizationMismatch
	}

	// Step 4: verify both signatures.
	signer := env.Signer
	if signer == nil || len(signer.SigningPublicKey) == 0 || len(signer.MlDsa65PublicKey) == 0 {
		return ErrSignatureMismatch
	}
	edPub := ed25519.PublicKey(signer.SigningPublicKey)
	if !v0crypto.VerifyEd25519(edPub, env.CanonicalPayload, env.Signature) {
		return ErrSignatureMismatch
	}
	if err := v0crypto.VerifyMLDSA65(signer.MlDsa65PublicKey, env.CanonicalPayload, env.MlDsa65Signature); err != nil {
		return ErrSignatureMismatch
	}

	// Step 5: freshness window.
	window := LiveFreshnessWindow
	if stored {
		window = StoredFreshnessWindow
	}
	signedAt := time.UnixMilli(int64(env.SignedAt))
	age := now.Sub(signedAt)
	if age > window || age < -window {
		return ErrExpiredSignature
	}

	return nil
}

// payloadForType unmarshals the Payload bytes into a proto.Message for re-serialization.
// The canonical payload already contains the deterministic serialization, so we only
// need to produce the same bytes. For types not yet fully registered, we use a
// placeholder empty message — the BuildCanonicalPayload call will produce identical
// output if the payload bytes are consistent.
func payloadForType(env *apb.SignedEnvelope) (proto.Message, error) {
	var msg proto.Message
	switch env.PayloadType {
	case apb.PayloadType_PAYLOAD_TYPE_UNSPECIFIED:
		msg = &apb.AetherPlaceholder{}
	case apb.PayloadType_PAYLOAD_TYPE_IDENTITY:
		msg = &apb.IdentityProfile{}
	case apb.PayloadType_PAYLOAD_TYPE_SERVER_MANIFEST:
		msg = &apb.ServerManifest{}
	case apb.PayloadType_PAYLOAD_TYPE_CHAT_MESSAGE:
		msg = &apb.ChatMessage{}
	case apb.PayloadType_PAYLOAD_TYPE_VOICE_STATE:
		msg = &apb.VoiceState{}
	case apb.PayloadType_PAYLOAD_TYPE_VOICE_PIPELINE_BASELINE:
		msg = &apb.VoicePipelineBaseline{}
	case apb.PayloadType_PAYLOAD_TYPE_VOICE_SIGNAL_FRAME:
		msg = &apb.VoiceSignalFrame{}
	case apb.PayloadType_PAYLOAD_TYPE_VOICE_SIGNAL_SESSION_STATE:
		msg = &apb.VoiceSignalSessionState{}
	case apb.PayloadType_PAYLOAD_TYPE_DM_PREKEY_BUNDLE:
		msg = &apb.DmPrekeyRecord{}
	case apb.PayloadType_PAYLOAD_TYPE_DM_MESSAGE:
		msg = &apb.DmMessageEnvelope{}
	case apb.PayloadType_PAYLOAD_TYPE_DM_TRANSPORT_DECISION:
		msg = &apb.DmTransportDecision{}
	case apb.PayloadType_PAYLOAD_TYPE_DM_DELIVERY_RECEIPT:
		msg = &apb.DmDeliveryReceipt{}
	case apb.PayloadType_PAYLOAD_TYPE_GROUP_DM_MEMBERSHIP_EVENT:
		msg = &apb.GroupDmMembershipEvent{}
	case apb.PayloadType_PAYLOAD_TYPE_FRIEND_REQUEST:
		msg = &apb.FriendRequest{}
	case apb.PayloadType_PAYLOAD_TYPE_PRESENCE_STATE:
		msg = &apb.PresenceStateEntry{}
	case apb.PayloadType_PAYLOAD_TYPE_PRESENCE_EVENT:
		msg = &apb.PresenceEventRecord{}
	case apb.PayloadType_PAYLOAD_TYPE_NOTIFICATION:
		msg = &apb.NotificationEvent{}
	case apb.PayloadType_PAYLOAD_TYPE_MENTION_TOKEN:
		msg = &apb.MentionToken{}
	case apb.PayloadType_PAYLOAD_TYPE_MENTION_DECISION:
		msg = &apb.PolicyDecision{}
	case apb.PayloadType_PAYLOAD_TYPE_ROLE_STATE:
		msg = &apb.RoleState{}
	case apb.PayloadType_PAYLOAD_TYPE_POLICIES:
		msg = &apb.PolicyDecision{}
	case apb.PayloadType_PAYLOAD_TYPE_MODERATION_EVENT:
		msg = &apb.ModerationEvent{}
	case apb.PayloadType_PAYLOAD_TYPE_SLOW_MODE_DECISION:
		msg = &apb.SlowModeDecision{}
	case apb.PayloadType_PAYLOAD_TYPE_GOVERNANCE_METADATA:
		msg = &apb.GovernanceMetadata{}
	case apb.PayloadType_PAYLOAD_TYPE_FRIEND_IDENTITY_SHARE:
		msg = &apb.FriendIdentityShare{}
	case apb.PayloadType_PAYLOAD_TYPE_FRIEND_INVITE_PAYLOAD:
		msg = &apb.FriendInvitePayload{}
	case apb.PayloadType_PAYLOAD_TYPE_PRESENCE_DISSEMINATION_POLICY:
		msg = &apb.PresenceDisseminationPolicy{}
	case apb.PayloadType_PAYLOAD_TYPE_DM_DIRECT_SESSION_LIFECYCLE:
		msg = &apb.DmDirectSessionLifecycle{}
	case apb.PayloadType_PAYLOAD_TYPE_GROUP_DM_SENDER_KEY_ENVELOPE:
		msg = &apb.GroupDmSenderKeyEnvelope{}
	case apb.PayloadType_PAYLOAD_TYPE_GROUP_DM_REKEY_DECISION:
		msg = &apb.GroupDmRekeyDecision{}
	case apb.PayloadType_PAYLOAD_TYPE_GROUP_DM_HISTORY_SYNC_DECISION:
		msg = &apb.GroupDmHistorySyncDecision{}
	case apb.PayloadType_PAYLOAD_TYPE_GROUP_DM_GROWTH_DECISION:
		msg = &apb.GroupDmGrowthDecision{}
	default:
		return nil, ErrUnsupportedPayloadType
	}
	if err := proto.Unmarshal(env.Payload, msg); err != nil {
		return nil, fmt.Errorf("payloadForType: unmarshal type %v: %w", env.PayloadType, err)
	}
	return msg, nil
}

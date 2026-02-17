package historysync

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aether/code_aether/pkg/v07/history"
	"github.com/aether/code_aether/pkg/v07/merkle"
)

type SyncMode string

const (
	SyncModeSnapshot SyncMode = "snapshot"
	SyncModeResume   SyncMode = "resume"
)

type Request struct {
	ServerID string
	ClientID string
	Epoch    int
	Mode     SyncMode
	Cursor   string
}

type Response struct {
	Payload   []byte
	Proof     string
	Token     ResumeToken
	Locked    LockedHistory
	Completed bool
}

type ResumeToken struct {
	Epoch       int
	Cursor      string
	HistoryRoot string
}

var (
	ErrResumeTokenInvalid = errors.New("historysync.token.invalid")
	tokenSeparator        = "|"
)

func EncodeResumeToken(token ResumeToken) string {
	return fmt.Sprintf("%d%s%s%s%s", token.Epoch, tokenSeparator, encodeTokenComponent(token.Cursor), tokenSeparator, encodeTokenComponent(token.HistoryRoot))
}

func DecodeResumeToken(raw string) (ResumeToken, error) {
	parts := strings.Split(raw, tokenSeparator)
	if len(parts) != 3 {
		return ResumeToken{}, ErrResumeTokenInvalid
	}
	epoch, err := strconv.Atoi(parts[0])
	if err != nil {
		return ResumeToken{}, ErrResumeTokenInvalid
	}
	cursor, err := decodeTokenComponent(parts[1])
	if err != nil {
		return ResumeToken{}, ErrResumeTokenInvalid
	}
	historyRoot, err := decodeTokenComponent(parts[2])
	if err != nil {
		return ResumeToken{}, ErrResumeTokenInvalid
	}
	return ResumeToken{Epoch: epoch, Cursor: cursor, HistoryRoot: historyRoot}, nil
}

func encodeTokenComponent(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeTokenComponent(encoded string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

type Checkpoint struct {
	Token     ResumeToken
	CreatedAt time.Time
	Summary   string
	ModeEpoch ModeEpochSegment
	Locked    LockedHistory
}

type ModeEpochSegment struct {
	Epoch  int
	Mode   history.HistoryMode
	Locked bool
}

type LockedHistory struct {
	Segment     ModeEpochSegment
	Capsule     history.CapsuleMetadata
	IsAvailable bool
	Reason      string
}

type handlerState struct {
	lastCheckpoint Checkpoint
}

func NewHandler() *handlerState {
	return &handlerState{}
}

func (h *handlerState) Handle(req Request) Response {
	capsule := history.NewCapsuleMetadata(req.ServerID, req.Epoch, history.ModeEpochHistory)
	payload := buildPayloadChunk(req, capsule)
	proof, proofErr := buildHistoryProof(payload, capsule)
	segment := ModeEpochSegment{Epoch: req.Epoch, Mode: history.ModeEpochHistory, Locked: true}
	locked := LockedHistory{Segment: segment, Capsule: capsule, IsAvailable: false, Reason: "keys_missing"}
	token := ResumeToken{Epoch: req.Epoch, Cursor: req.Cursor, HistoryRoot: capsule.Root}
	h.lastCheckpoint = Checkpoint{Token: token, CreatedAt: time.Now().UTC(), Summary: fmt.Sprintf("history sync epoch=%d mode=%s", req.Epoch, req.Mode), ModeEpoch: segment, Locked: locked}
	completed := proofErr == nil
	proofPayload := proof
	if proofErr != nil {
		proofPayload = fmt.Sprintf("proof.error:%v", proofErr)
	}
	return Response{Payload: payload, Proof: proofPayload, Token: token, Locked: locked, Completed: completed}
}

func (h *handlerState) Checkpoint() Checkpoint {
	return h.lastCheckpoint
}

func ResumeTokenFromCheckpoint(checkpoint Checkpoint) ResumeToken {
	return checkpoint.Token
}

func ReencryptCapsule(capsule history.CapsuleMetadata, recipient string) history.CapsuleMetadata {
	capsule.CapsuleID = fmt.Sprintf("%s-%s", capsule.CapsuleID, recipient)
	capsule.Root = fmt.Sprintf("reencrypted:%s", capsule.Root)
	return capsule
}

func buildPayloadChunk(req Request, capsule history.CapsuleMetadata) []byte {
	return []byte(fmt.Sprintf("%s:%d:%s:%s", req.ServerID, req.Epoch, req.Mode, req.Cursor))
}

func buildHistoryProof(payload []byte, capsule history.CapsuleMetadata) (string, error) {
	builder := merkle.NewBuilder()
	builder.AddLeaf([]byte(capsule.Root))
	builder.AddLeaf(payload)
	proof, err := builder.Proof(1)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d:%s", builder.Root(), 1, strings.Join(proof, ",")), nil
}

func ModeEpochDescription(segment ModeEpochSegment) string {
	status := "open"
	if segment.Locked {
		status = "locked"
	}
	return fmt.Sprintf("mode=%s epoch=%d status=%s", segment.Mode, segment.Epoch, status)
}

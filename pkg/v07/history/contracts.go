package history

import "fmt"

type HistoryMode string

const (
	ModeEpochHistory HistoryMode = "epoch"
	ModeRolling      HistoryMode = "rolling"
)

func CanonicalRoot(base string, epoch int, mode HistoryMode) string {
	return fmt.Sprintf("root:%s:%s:%d", base, mode, epoch)
}

type ProofClassification struct {
	Matched bool
	Reason  string
}

func ClassifyProof(expected, actual string) ProofClassification {
	if expected == actual {
		return ProofClassification{Matched: true, Reason: "proof.match"}
	}
	return ProofClassification{Matched: false, Reason: "proof.mismatch"}
}

type SyncState struct {
	Cursor string
	Epoch  int
	Ready  bool
}

func ResumeSync(state SyncState) SyncState {
	state.Ready = true
	return state
}

type CapsuleMetadata struct {
	CapsuleID string
	Root      string
	Epoch     int
}

func NewCapsuleMetadata(base string, epoch int, mode HistoryMode) CapsuleMetadata {
	return CapsuleMetadata{
		CapsuleID: fmt.Sprintf("capsule-%d", epoch),
		Root:      CanonicalRoot(base, epoch, mode),
		Epoch:     epoch,
	}
}

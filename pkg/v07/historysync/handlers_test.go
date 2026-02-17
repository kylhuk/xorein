package historysync

import (
	"errors"
	"strings"
	"testing"

	"github.com/aether/code_aether/pkg/v07/history"
	"github.com/aether/code_aether/pkg/v07/merkle"
)

func TestHandlerCheckpointAndToken(t *testing.T) {
	h := NewHandler()
	req := Request{ServerID: "server-1", Epoch: 2, Cursor: "cur"}
	resp := h.Handle(req)
	want := history.CanonicalRoot(req.ServerID, req.Epoch, history.ModeEpochHistory)
	if resp.Token.HistoryRoot != want {
		t.Fatalf("unexpected history root %s", resp.Token.HistoryRoot)
	}
	checkpoint := h.Checkpoint()
	if checkpoint.Token.HistoryRoot != resp.Token.HistoryRoot {
		t.Fatalf("checkpoint root mismatch")
	}
	if token := ResumeTokenFromCheckpoint(checkpoint); token != resp.Token {
		t.Fatalf("resume token mismatch")
	}
	if !resp.Completed {
		t.Fatalf("expected completed response")
	}
}

func TestModeEpochDescriptionLocked(t *testing.T) {
	segment := ModeEpochSegment{Epoch: 5, Mode: history.ModeEpochHistory, Locked: true}
	desc := ModeEpochDescription(segment)
	if !strings.Contains(desc, "status=locked") {
		t.Fatalf("expected locked status, got %s", desc)
	}
	segment.Locked = false
	desc = ModeEpochDescription(segment)
	if !strings.Contains(desc, "status=open") {
		t.Fatalf("expected open status, got %s", desc)
	}
}

func TestReencryptCapsuleUpdatesRoot(t *testing.T) {
	capsule := history.NewCapsuleMetadata("srv", 7, history.ModeEpochHistory)
	updated := ReencryptCapsule(capsule, "recipient-x")
	if !strings.HasPrefix(updated.Root, "reencrypted:") {
		t.Fatalf("expected reencrypted root, got %s", updated.Root)
	}
	if updated.CapsuleID == capsule.CapsuleID {
		t.Fatalf("expected capsule id mutated")
	}
}

func TestMalformedMerkleProofs(t *testing.T) {
	builder := merkle.NewBuilder()
	builder.AddLeaf([]byte("leaf"))
	builder.AddLeaf([]byte("mirror"))
	root := builder.Root()
	proof, err := builder.Proof(0)
	if err != nil {
		t.Fatalf("proof generation failed: %v", err)
	}
	cases := []struct {
		name  string
		root  string
		data  []byte
		proof []string
		index int
		total int
		want  error
	}{
		{name: "invalid total leaves", root: root, data: []byte("leaf"), proof: proof, index: 0, total: 0, want: merkle.ErrInvalidProof},
		{name: "invalid proof entry", root: root, data: []byte("leaf"), proof: []string{"zz"}, index: 0, total: 2, want: merkle.ErrInvalidProof},
		{name: "invalid root encoding", root: "root:server-1:epoch:3", data: []byte("leaf"), proof: proof, index: 0, total: 2, want: merkle.ErrInvalidProof},
		{name: "divergent root", root: root, data: []byte("tampered"), proof: proof, index: 0, total: 2, want: merkle.ErrDivergentRoot},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := merkle.VerifyProof(tc.root, tc.data, tc.proof, tc.index, tc.total)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, err)
			}
		})
	}
}

func TestResumeTokenEncodingRoundTrip(t *testing.T) {
	token := ResumeToken{Epoch: 4, Cursor: "cursor", HistoryRoot: "root-value"}
	encoded := EncodeResumeToken(token)
	decoded, err := DecodeResumeToken(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded != token {
		t.Fatalf("round trip mismatch %v vs %v", decoded, token)
	}
}

func TestResumeTokenDecodeErrors(t *testing.T) {
	if _, err := DecodeResumeToken("bad|input"); err != ErrResumeTokenInvalid {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func FuzzDecodeResumeToken(f *testing.F) {
	seed := ResumeToken{Epoch: 2, Cursor: "seed", HistoryRoot: "root"}
	f.Add(EncodeResumeToken(seed))
	f.Add("bad|input")
	f.Add("||||")

	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = DecodeResumeToken(raw)
	})
}

func FuzzVerifyProof(f *testing.F) {
	builder := merkle.NewBuilder()
	builder.AddLeaf([]byte("leaf"))
	builder.AddLeaf([]byte("mirror"))
	root := builder.Root()
	proof, err := builder.Proof(0)
	if err != nil {
		f.Fatalf("proof generation failed: %v", err)
	}

	f.Add(root, "leaf", strings.Join(proof, ","), 0, 2)
	f.Add("invalid-root", "leaf", "zz", 0, 2)

	f.Fuzz(func(t *testing.T, root string, payload string, proofCSV string, index int, total int) {
		var proof []string
		if strings.TrimSpace(proofCSV) != "" {
			proof = strings.Split(proofCSV, ",")
		}
		_ = merkle.VerifyProof(root, []byte(payload), proof, index, total)
	})
}

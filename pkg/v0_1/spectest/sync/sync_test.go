package sync_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	syncpkg "github.com/aether/code_aether/pkg/v0_1/family/sync"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/spectest"
)

func vectorDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
	return filepath.Join(root, "docs", "spec", "v0.1", "91-test-vectors")
}

func TestSyncPins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

// katMessage is the JSON shape for messages in sync KAT inputs.
type katMessage struct {
	ID        string `json:"id"`
	Sequence  int64  `json:"sequence"`
	CreatedAt string `json:"created_at"`
	Body      string `json:"body"`
	Signature string `json:"signature"`
}

type katCase struct {
	Label          string          `json:"label"`
	Operation      string          `json:"operation"`
	AdvertisedCaps []string        `json:"advertised_caps"`
	Payload        json.RawMessage `json:"payload"`
}

type katInputs struct {
	ServerID    string       `json:"server_id"`
	IsArchivist bool         `json:"is_archivist"`
	Messages    []katMessage `json:"messages"`
	Members     []string     `json:"members"`
	Cases       []katCase    `json:"cases"`
}

type katExpectedCase struct {
	// coverage fields
	ServerID      *string  `json:"server_id,omitempty"`
	AvailableFrom *int64   `json:"available_from,omitempty"`
	AvailableTo   *int64   `json:"available_to,omitempty"`
	SnapshotRoot  *string  `json:"snapshot_root,omitempty"`
	// fetch fields
	MessagesCount *int     `json:"messages_count,omitempty"`
	// push fields
	AcceptedCount  *int `json:"accepted_count,omitempty"`
	DuplicateCount *int `json:"duplicate_count,omitempty"`
	RejectedCount  *int `json:"rejected_count,omitempty"`
	// error
	ErrorCode string `json:"error_code,omitempty"`
}

type katExpected struct {
	Cases []katExpectedCase `json:"cases"`
}

type kat struct {
	Description    string      `json:"description"`
	Inputs         katInputs   `json:"inputs"`
	ExpectedOutput katExpected `json:"expected_output"`
}

func runSyncKAT(t *testing.T, filename string) {
	t.Helper()
	vdir := vectorDir(t)
	data, err := os.ReadFile(filepath.Join(vdir, filename))
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	var k kat
	if err := json.Unmarshal(data, &k); err != nil {
		t.Fatalf("decode %s: %v", filename, err)
	}

	h := syncpkg.NewHandler(k.Inputs.IsArchivist)
	for _, m := range k.Inputs.Messages {
		ts, err := time.Parse(time.RFC3339, m.CreatedAt)
		if err != nil {
			t.Fatalf("parse created_at %q: %v", m.CreatedAt, err)
		}
		h.SeedMessageForTest(&syncpkg.Message{
			ID:        m.ID,
			ServerID:  k.Inputs.ServerID,
			Sequence:  m.Sequence,
			CreatedAt: ts,
			Body:      []byte(m.Body),
			Signature: []byte(m.Signature),
		})
	}
	for _, peer := range k.Inputs.Members {
		h.AddMember(k.Inputs.ServerID, peer)
	}

	ctx := context.Background()
	for i, c := range k.Inputs.Cases {
		want := k.ExpectedOutput.Cases[i]
		t.Run(c.Label, func(t *testing.T) {
			req := &proto.PeerStreamRequest{
				Operation:      c.Operation,
				AdvertisedCaps: c.AdvertisedCaps,
				Payload:        []byte(c.Payload),
			}
			resp := h.HandleStream(ctx, req)

			if want.ErrorCode != "" {
				if resp.Error == nil {
					t.Fatalf("expected error %q, got success (payload=%s)", want.ErrorCode, resp.Payload)
				}
				if resp.Error.Code != want.ErrorCode {
					t.Fatalf("error code: want %q got %q", want.ErrorCode, resp.Error.Code)
				}
				return
			}
			if resp.Error != nil {
				t.Fatalf("unexpected error: %s", resp.Error)
			}

			switch c.Operation {
			case "sync.coverage":
				var out struct {
					AvailableFrom int64  `json:"available_from"`
					AvailableTo   int64  `json:"available_to"`
					SnapshotRoot  string `json:"snapshot_root"`
				}
				if err := json.Unmarshal(resp.Payload, &out); err != nil {
					t.Fatalf("decode coverage response: %v", err)
				}
				if want.AvailableFrom != nil && out.AvailableFrom != *want.AvailableFrom {
					t.Errorf("available_from: want %d got %d", *want.AvailableFrom, out.AvailableFrom)
				}
				if want.AvailableTo != nil && out.AvailableTo != *want.AvailableTo {
					t.Errorf("available_to: want %d got %d", *want.AvailableTo, out.AvailableTo)
				}
				if want.SnapshotRoot != nil && out.SnapshotRoot != *want.SnapshotRoot {
					t.Errorf("snapshot_root: want %q got %q", *want.SnapshotRoot, out.SnapshotRoot)
				}

			case "sync.fetch":
				var out struct {
					Messages    json.RawMessage `json:"messages"`
					NotFoundIDs []string        `json:"not_found_ids"`
				}
				if err := json.Unmarshal(resp.Payload, &out); err != nil {
					t.Fatalf("decode fetch response: %v", err)
				}
				if want.MessagesCount != nil {
					var msgs []json.RawMessage
					if err := json.Unmarshal(out.Messages, &msgs); err != nil {
						t.Fatalf("decode messages array: %v", err)
					}
					if len(msgs) != *want.MessagesCount {
						t.Errorf("messages_count: want %d got %d", *want.MessagesCount, len(msgs))
					}
				}

			case "sync.push":
				var out struct {
					AcceptedCount  int `json:"accepted_count"`
					DuplicateCount int `json:"duplicate_count"`
					RejectedCount  int `json:"rejected_count"`
				}
				if err := json.Unmarshal(resp.Payload, &out); err != nil {
					t.Fatalf("decode push response: %v", err)
				}
				if want.AcceptedCount != nil && out.AcceptedCount != *want.AcceptedCount {
					t.Errorf("accepted_count: want %d got %d", *want.AcceptedCount, out.AcceptedCount)
				}
				if want.DuplicateCount != nil && out.DuplicateCount != *want.DuplicateCount {
					t.Errorf("duplicate_count: want %d got %d", *want.DuplicateCount, out.DuplicateCount)
				}
				if want.RejectedCount != nil && out.RejectedCount != *want.RejectedCount {
					t.Errorf("rejected_count: want %d got %d", *want.RejectedCount, out.RejectedCount)
				}
			}
		})
	}
}

func TestSyncCoverageEmpty(t *testing.T)      { runSyncKAT(t, "sync_coverage_empty_kat.json") }
func TestSyncCoverageFull(t *testing.T)       { runSyncKAT(t, "sync_coverage_full_kat.json") }
func TestSyncCoverageGap(t *testing.T)        { runSyncKAT(t, "sync_coverage_gap_kat.json") }
func TestSyncFetch(t *testing.T)              { runSyncKAT(t, "sync_fetch_kat.json") }
func TestSyncFetchLimit(t *testing.T)         { runSyncKAT(t, "sync_fetch_limit_kat.json") }
func TestSyncFetchNonmember(t *testing.T)     { runSyncKAT(t, "sync_fetch_nonmember_kat.json") }
func TestSyncPush(t *testing.T)              { runSyncKAT(t, "sync_push_kat.json") }
func TestSyncPushInvalidSig(t *testing.T)    { runSyncKAT(t, "sync_push_invalid_sig_kat.json") }
func TestSyncArchivistIsolation(t *testing.T) { runSyncKAT(t, "sync_archivist_isolation_kat.json") }

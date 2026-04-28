package moderation_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	govpkg "github.com/aether/code_aether/pkg/v0_1/family/governance"
	modpkg "github.com/aether/code_aether/pkg/v0_1/family/moderation"
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

func TestModerationPins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

type katCase struct {
	Label          string          `json:"label"`
	Operation      string          `json:"operation"`
	AdvertisedCaps []string        `json:"advertised_caps"`
	Payload        json.RawMessage `json:"payload"`
}

type govSetupEntry struct {
	ServerID string `json:"server_id"`
	PeerID   string `json:"peer_id"`
	Role     string `json:"role"`
}

type katInputs struct {
	GovernanceSetup []govSetupEntry     `json:"governance_setup"`
	ServerMembers   map[string][]string `json:"server_members"`  // serverID → []peerID
	StoredMessages  []string            `json:"stored_messages"` // message IDs pre-seeded
	Cases           []katCase           `json:"cases"`
}

type katExpectedCase struct {
	Accepted  *bool  `json:"accepted,omitempty"`
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

func runModerationKAT(t *testing.T, filename string) {
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

	gov := govpkg.NewHandler()
	for _, gs := range k.Inputs.GovernanceSetup {
		gov.AssignForTest(gs.ServerID, gs.PeerID, parseRole(t, gs.Role))
	}
	mod := modpkg.New(gov)
	for serverID, peers := range k.Inputs.ServerMembers {
		for _, peerID := range peers {
			mod.AddMember(serverID, peerID)
		}
	}
	for _, msgID := range k.Inputs.StoredMessages {
		mod.StoreMessage(msgID)
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
			resp := mod.HandleStream(ctx, req)
			if want.ErrorCode != "" {
				if resp.Error == nil {
					t.Fatalf("expected error %q, got success", want.ErrorCode)
				}
				if resp.Error.Code != want.ErrorCode {
					t.Fatalf("error code: want %q got %q", want.ErrorCode, resp.Error.Code)
				}
				return
			}
			if resp.Error != nil {
				t.Fatalf("unexpected error: %s", resp.Error)
			}
			if want.Accepted != nil {
				var out struct {
					Accepted bool `json:"accepted"`
				}
				if err := json.Unmarshal(resp.Payload, &out); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if out.Accepted != *want.Accepted {
					t.Fatalf("accepted: want %v got %v", *want.Accepted, out.Accepted)
				}
			}
		})
	}
}

func parseRole(t *testing.T, s string) govpkg.Role {
	t.Helper()
	switch s {
	case "member":
		return govpkg.RoleMember
	case "moderator":
		return govpkg.RoleModerator
	case "admin":
		return govpkg.RoleAdmin
	case "owner":
		return govpkg.RoleOwner
	default:
		t.Fatalf("unknown role %q", s)
		return 0
	}
}

func TestModerationKick(t *testing.T)              { runModerationKAT(t, "moderation_kick_kat.json") }
func TestModerationBanUnban(t *testing.T)          { runModerationKAT(t, "moderation_ban_unban_kat.json") }
func TestModerationMute(t *testing.T)              { runModerationKAT(t, "moderation_mute_kat.json") }
func TestModerationSlowMode(t *testing.T)          { runModerationKAT(t, "moderation_slow_mode_kat.json") }
func TestModerationDeleteMessage(t *testing.T)     { runModerationKAT(t, "moderation_delete_message_kat.json") }
func TestModerationUnauthorized(t *testing.T)      { runModerationKAT(t, "moderation_unauthorized_kat.json") }
func TestModerationForbiddenTarget(t *testing.T)   { runModerationKAT(t, "moderation_forbidden_target_kat.json") }
func TestModerationSignatureMismatch(t *testing.T) { runModerationKAT(t, "moderation_signature_mismatch_kat.json") }

package governance_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	govpkg "github.com/aether/code_aether/pkg/v0_1/family/governance"
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

func TestGovernancePins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

// ------- standard KAT shape -------

type katCase struct {
	Label          string          `json:"label"`
	Operation      string          `json:"operation"`
	AdvertisedCaps []string        `json:"advertised_caps"`
	Payload        json.RawMessage `json:"payload"`
}

type katSetupEntry struct {
	ServerID string `json:"server_id"`
	PeerID   string `json:"peer_id"`
	Role     string `json:"role"`
}

type preexistingRole struct {
	ServerID            string `json:"server_id"`
	RoleName            string `json:"role_name"`
	PermissionsBitfield uint64 `json:"permissions_bitfield"`
}

type preAssign struct {
	ServerID           string `json:"server_id"`
	PeerID             string `json:"peer_id"`
	Role               string `json:"role"`
	PolicyVersionUsed  uint64 `json:"policy_version_used"`
}

type katInputs struct {
	GovernanceSetup        []katSetupEntry   `json:"governance_setup"`
	PreexistingCustomRoles []preexistingRole  `json:"preexisting_custom_roles"`
	PreAssign              *preAssign         `json:"pre_assign"`
	Cases                  []katCase         `json:"cases"`
}

type katExpectedCase struct {
	Accepted  *bool  `json:"accepted,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type katExpected struct {
	Cases []katExpectedCase `json:"cases"`
}

type kat struct {
	Description    string          `json:"description"`
	Inputs         katInputs       `json:"inputs"`
	ExpectedOutput json.RawMessage `json:"expected_output"`
}

func runGovernanceKAT(t *testing.T, filename string) {
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

	var expected katExpected
	if err := json.Unmarshal(k.ExpectedOutput, &expected); err != nil {
		t.Fatalf("decode expected_output: %v", err)
	}

	h := govpkg.NewHandler()
	for _, setup := range k.Inputs.GovernanceSetup {
		h.AssignForTest(setup.ServerID, setup.PeerID, parseTestRole(t, setup.Role))
	}
	for _, cr := range k.Inputs.PreexistingCustomRoles {
		h.CreateRoleForTest(cr.ServerID, cr.RoleName, cr.PermissionsBitfield)
	}

	ctx := context.Background()
	// Execute any pre_assign operation (advances policy_version before test cases run).
	if pa := k.Inputs.PreAssign; pa != nil {
		prePayload, _ := json.Marshal(map[string]any{
			"actor_peer_id":  pa.PeerID,
			"target_peer_id": pa.PeerID,
			"server_id":      pa.ServerID,
			"role":           pa.Role,
			"policy_version": pa.PolicyVersionUsed,
		})
		// Find the actor who can perform this assignment (use the first admin/owner in setup).
		for _, s := range k.Inputs.GovernanceSetup {
			if s.ServerID == pa.ServerID && (s.Role == "admin" || s.Role == "owner") {
				actorPayload, _ := json.Marshal(map[string]any{
					"actor_peer_id":  s.PeerID,
					"target_peer_id": pa.PeerID,
					"server_id":      pa.ServerID,
					"role":           pa.Role,
					"policy_version": pa.PolicyVersionUsed,
				})
				_ = prePayload
				req := &proto.PeerStreamRequest{
					Operation:      "governance.assign_role",
					AdvertisedCaps: []string{"cap.rbac"},
					Payload:        actorPayload,
				}
				if resp := h.HandleStream(ctx, req); resp.Error != nil {
					t.Fatalf("pre_assign failed: %s", resp.Error)
				}
				break
			}
		}
	}
	for i, c := range k.Inputs.Cases {
		want := expected.Cases[i]
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
			if want.Accepted != nil {
				var out struct {
					Accepted bool `json:"accepted"`
				}
				if err := json.Unmarshal(resp.Payload, &out); err != nil {
					t.Fatalf("decode response payload: %v", err)
				}
				if out.Accepted != *want.Accepted {
					t.Fatalf("accepted: want %v got %v", *want.Accepted, out.Accepted)
				}
			}
		})
	}
}

func parseTestRole(t *testing.T, s string) govpkg.Role {
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

// ------- convergence KAT shape -------

type convergenceCase struct {
	Label          string `json:"label"`
	A              struct {
		Role          string `json:"role"`
		Version       uint64 `json:"version"`
		UpdatedAtUnix uint64 `json:"updated_at_unix"`
	} `json:"a"`
	B struct {
		Role          string `json:"role"`
		Version       uint64 `json:"version"`
		UpdatedAtUnix uint64 `json:"updated_at_unix"`
	} `json:"b"`
	ExpectedWinner string `json:"expected_winner"`
}

type convergenceKAT struct {
	Inputs struct {
		Cases []convergenceCase `json:"cases"`
	} `json:"inputs"`
}

func TestGovernanceConvergence(t *testing.T) {
	vdir := vectorDir(t)
	data, err := os.ReadFile(filepath.Join(vdir, "governance_convergence_kat.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var k convergenceKAT
	if err := json.Unmarshal(data, &k); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, c := range k.Inputs.Cases {
		t.Run(c.Label, func(t *testing.T) {
			rsA := govpkg.RoleState{
				Role:      parseTestRole(t, c.A.Role),
				Version:   c.A.Version,
				UpdatedAt: time.Unix(int64(c.A.UpdatedAtUnix), 0),
			}
			rsB := govpkg.RoleState{
				Role:      parseTestRole(t, c.B.Role),
				Version:   c.B.Version,
				UpdatedAt: time.Unix(int64(c.B.UpdatedAtUnix), 0),
			}
			winner := govpkg.ResolveRoleState(rsA, rsB)
			var expectedRS govpkg.RoleState
			switch c.ExpectedWinner {
			case "a":
				expectedRS = rsA
			case "b":
				expectedRS = rsB
			default:
				t.Fatalf("unknown winner %q", c.ExpectedWinner)
			}
			if winner.Role != expectedRS.Role || winner.Version != expectedRS.Version {
				t.Errorf("wrong winner: want role=%v ver=%d, got role=%v ver=%d",
					expectedRS.Role, expectedRS.Version, winner.Role, winner.Version)
			}
		})
	}
}

func TestGovernanceAssignModerator(t *testing.T) {
	runGovernanceKAT(t, "governance_assign_moderator_kat.json")
}
func TestGovernanceAssignAdmin(t *testing.T) { runGovernanceKAT(t, "governance_assign_admin_kat.json") }
func TestGovernanceRevokeRole(t *testing.T)  { runGovernanceKAT(t, "governance_revoke_role_kat.json") }
func TestGovernanceCreateRole(t *testing.T)  { runGovernanceKAT(t, "governance_create_role_kat.json") }
func TestGovernanceDeleteRole(t *testing.T)  { runGovernanceKAT(t, "governance_delete_role_kat.json") }
func TestGovernanceSync(t *testing.T)        { runGovernanceKAT(t, "governance_sync_kat.json") }
func TestGovernanceUnauthorized(t *testing.T) {
	runGovernanceKAT(t, "governance_unauthorized_kat.json")
}
func TestGovernanceForbiddenTarget(t *testing.T) {
	runGovernanceKAT(t, "governance_forbidden_target_kat.json")
}
func TestGovernanceStaleVersion(t *testing.T) {
	runGovernanceKAT(t, "governance_stale_version_kat.json")
}
func TestGovernanceReplay(t *testing.T) { runGovernanceKAT(t, "governance_replay_kat.json") }

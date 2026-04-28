package voice_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	ms "github.com/aether/code_aether/pkg/v0_1/mode/mediashield"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/spectest"
	voicepkg "github.com/aether/code_aether/pkg/v0_1/family/voice"
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

func TestVoicePins(t *testing.T) {
	spectest.VerifyPin(t, vectorDir(t))
}

type katCase struct {
	Label          string          `json:"label"`
	Operation      string          `json:"operation"`
	AdvertisedCaps []string        `json:"advertised_caps"`
	Payload        json.RawMessage `json:"payload"`
	Synthetic      string          `json:"synthetic,omitempty"` // "large_frame" → generate oversized ciphertext
}

type sessionSetupEntry struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
	Codec     string `json:"codec"`
}

type katInputs struct {
	MediaShieldPeerKeys map[string]map[string]string `json:"mediashield_peer_keys"`
	SessionSetup        []sessionSetupEntry          `json:"session_setup"`
	RelayMode           bool                         `json:"relay_mode"`
	SFUMode             bool                         `json:"sfu_mode"`
	Cases               []katCase                    `json:"cases"`
}

type katExpectedCase struct {
	OK          *bool  `json:"ok,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	Coordinator string `json:"coordinator,omitempty"`
}

type katExpected struct {
	Cases []katExpectedCase `json:"cases"`
}

type kat struct {
	Description    string      `json:"description"`
	Inputs         katInputs   `json:"inputs"`
	ExpectedOutput katExpected `json:"expected_output"`
}

func buildKey(peerID, raw string) *ms.PeerKey {
	key := make([]byte, 32)
	copy(key, []byte(raw))
	return &ms.PeerKey{PeerID: peerID, Key: key}
}

// buildPayload returns the bytes to send to the handler. For synthetic cases it
// generates appropriate payloads (e.g. oversized ciphertext).
func buildPayload(c katCase) []byte {
	if c.Synthetic == "large_frame" {
		// Build a frame payload with ciphertext > 65535 bytes.
		type framePayload struct {
			SessionID  string `json:"session_id"`
			SenderID   string `json:"sender_id"`
			Counter    uint64 `json:"counter"`
			Ciphertext []byte `json:"ciphertext"`
		}
		var base struct {
			SessionID string `json:"session_id"`
			SenderID  string `json:"sender_id"`
			Counter   uint64 `json:"counter"`
		}
		_ = json.Unmarshal(c.Payload, &base)
		p := framePayload{
			SessionID:  base.SessionID,
			SenderID:   base.SenderID,
			Counter:    base.Counter,
			Ciphertext: bytes.Repeat([]byte{0xAB}, 65536),
		}
		b, _ := json.Marshal(p)
		return b
	}
	return []byte(c.Payload)
}

func runVoiceKAT(t *testing.T, filename string) {
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

	h := &voicepkg.Handler{
		LocalPeerID: "test-local",
		RelayMode:   k.Inputs.RelayMode,
	}

	for sessionID, peers := range k.Inputs.MediaShieldPeerKeys {
		keys := make(map[string]*ms.PeerKey, len(peers))
		for peerID, rawKey := range peers {
			keys[peerID] = buildKey(peerID, rawKey)
		}
		if err := h.CreateSession(sessionID, nil, keys); err != nil {
			t.Fatalf("CreateSession %q: %v", sessionID, err)
		}
	}

	ctx := context.Background()
	for i, c := range k.Inputs.Cases {
		want := k.ExpectedOutput.Cases[i]
		t.Run(c.Label, func(t *testing.T) {
			req := &proto.PeerStreamRequest{
				Operation:      c.Operation,
				AdvertisedCaps: c.AdvertisedCaps,
				Payload:        buildPayload(c),
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
		})
	}
}

func TestVoiceJoinLeave(t *testing.T)                 { runVoiceKAT(t, "voice_join_leave_kat.json") }
func TestVoiceMute(t *testing.T)                      { runVoiceKAT(t, "voice_mute_kat.json") }
func TestVoiceOfferAnswer(t *testing.T)               { runVoiceKAT(t, "voice_offer_answer_kat.json") }
func TestVoiceICE(t *testing.T)                       { runVoiceKAT(t, "voice_ice_kat.json") }
func TestVoiceFrame(t *testing.T)                     { runVoiceKAT(t, "voice_frame_kat.json") }
func TestVoiceRestart(t *testing.T)                   { runVoiceKAT(t, "voice_restart_kat.json") }
func TestVoiceTerminate(t *testing.T)                 { runVoiceKAT(t, "voice_terminate_kat.json") }
func TestVoiceSFUElection(t *testing.T)               { runVoiceKAT(t, "voice_sfu_election_kat.json") }
func TestVoiceSFUOpacity(t *testing.T)                { runVoiceKAT(t, "voice_sfu_opacity_kat.json") }
func TestVoiceRelayOpacity(t *testing.T)              { runVoiceKAT(t, "voice_relay_opacity_kat.json") }
func TestVoiceCodecUnsupported(t *testing.T)          { runVoiceKAT(t, "voice_codec_unsupported_kat.json") }
func TestVoiceSignalReplay(t *testing.T)              { runVoiceKAT(t, "voice_signal_replay_kat.json") }
func TestVoiceMediaShieldKeyUnavailable(t *testing.T) { runVoiceKAT(t, "voice_mediashield_key_unavailable_kat.json") }

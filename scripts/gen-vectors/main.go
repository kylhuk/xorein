// Command gen-vectors generates KAT JSON files for protocol families that are
// missing test vectors and updates pin.sha256 in the output directory.
//
// Usage:
//
//	go run scripts/gen-vectors/main.go [--out docs/spec/v0.1/91-test-vectors/]
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	outDir := flag.String("out", "docs/spec/v0.1/91-test-vectors/", "output directory for KAT JSON files and pin.sha256")
	flag.Parse()

	dir := filepath.Clean(*outDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	generated := generateVectors()
	for filename, content := range generated {
		path := filepath.Join(dir, filename)
		data, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			log.Fatalf("marshal %s: %v", filename, err)
		}
		data = append(data, '\n')
		if err := os.WriteFile(path, data, 0o644); err != nil {
			log.Fatalf("write %s: %v", filename, err)
		}
		fmt.Printf("wrote %s\n", path)
	}

	if err := updatePin(dir); err != nil {
		log.Fatalf("update pin: %v", err)
	}
	fmt.Printf("updated %s\n", filepath.Join(dir, "pin.sha256"))
}

// generateVectors returns a map of filename → KAT content for families missing vectors.
func generateVectors() map[string]any {
	return map[string]any{
		"manifest_kat.json": map[string]any{
			"description": "manifest family KAT: publish and fetch by hash; fetch not found",
			"source":      "docs/spec/v0.1/42-family-manifest.md",
			"inputs": map[string]any{
				"cases": []any{
					map[string]any{
						"label":          "publish and fetch by hash",
						"operation":      "manifest.publish",
						"advertised_caps": []string{"cap.manifest"},
						"payload": map[string]any{
							"scope_id":     "sc1",
							"publisher_id": "alice",
							"content":      map[string]any{"key": "value"},
						},
					},
					map[string]any{
						"label":          "fetch not found",
						"operation":      "manifest.fetch",
						"advertised_caps": []string{"cap.manifest"},
						"payload": map[string]any{
							"hash": "nonexistent-hash",
						},
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"ok": true, "hash_non_empty": true},
					map[string]any{"error_code": "MANIFEST_NOT_FOUND"},
				},
			},
		},
		"identity_kat.json": map[string]any{
			"description": "identity family KAT: fetch unknown identity",
			"source":      "docs/spec/v0.1/43-family-identity.md",
			"inputs": map[string]any{
				"cases": []any{
					map[string]any{
						"label":          "fetch unknown identity",
						"operation":      "identity.fetch",
						"advertised_caps": []string{"cap.identity"},
						"payload": map[string]any{
							"peer_id": "ghost",
						},
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"error_code": "OPERATION_FAILED"},
				},
			},
		},
		"chat_kat.json": map[string]any{
			"description": "chat family KAT: join and send and history; send without join",
			"source":      "docs/spec/v0.1/41-family-chat.md",
			"inputs": map[string]any{
				"setup": map[string]any{
					"registered_channels": []string{"ch1"},
				},
				"cases": []any{
					map[string]any{
						"label":          "join channel ch1 as alice",
						"operation":      "chat.join",
						"advertised_caps": []string{"cap.chat"},
						"payload": map[string]any{
							"scope_id":   "ch1",
							"scope_type": "channel",
							"peer_id":    "alice",
						},
					},
					map[string]any{
						"label":          "send message to ch1",
						"operation":      "chat.send",
						"advertised_caps": []string{"cap.chat"},
						"payload": map[string]any{
							"scope_id":   "ch1",
							"scope_type": "channel",
							"sender_id":  "alice",
							"body":       "aGVsbG8=",
						},
					},
					map[string]any{
						"label":          "fetch history for ch1",
						"operation":      "chat.history",
						"advertised_caps": []string{"cap.chat"},
						"payload": map[string]any{
							"scope_id": "ch1",
							"limit":    10,
						},
					},
					map[string]any{
						"label":          "send to unknown channel without join",
						"operation":      "chat.join",
						"advertised_caps": []string{"cap.chat"},
						"payload": map[string]any{
							"scope_id":   "unknown-ch",
							"scope_type": "channel",
							"peer_id":    "alice",
						},
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"ok": true},
					map[string]any{"ok": true},
					map[string]any{"count": 1},
					map[string]any{"error_code": "CHANNEL_NOT_FOUND"},
				},
			},
		},
		"dm_kat.json": map[string]any{
			"description": "dm family KAT: rate limiting (61st DM rejected)",
			"source":      "docs/spec/v0.1/45-family-dm.md",
			"inputs": map[string]any{
				"cases": []any{
					map[string]any{
						"label":                  "rate limit: 61st DM must be rejected",
						"description":            "send 61 DMs from same sender in rapid succession; 61st returns DM_RATE_LIMITED",
						"sender_id":              "alice",
						"recipient_id":           "bob",
						"count":                  61,
						"expect_rate_limited_at": 61,
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"error_code": "DM_RATE_LIMITED"},
				},
			},
		},
		"groupdm_kat.json": map[string]any{
			"description": "groupdm family KAT: create group and send; non-member send rejected",
			"source":      "docs/spec/v0.1/46-family-groupdm.md",
			"inputs": map[string]any{
				"setup": map[string]any{
					"groups": []any{
						map[string]any{
							"group_id":   "g1",
							"creator_id": "alice",
							"members":    []string{"alice", "bob"},
						},
					},
				},
				"cases": []any{
					map[string]any{
						"label":          "create group g1",
						"operation":      "groupdm.create",
						"advertised_caps": []string{"cap.group-dm", "mode.tree"},
						"payload": map[string]any{
							"group_id": "g1",
							"creator":  map[string]any{"peer_id": "alice"},
						},
					},
					map[string]any{
						"label":          "non-member carol cannot send to g1",
						"description":    "carol is not in g1; send must fail with OPERATION_FAILED",
						"operation":      "groupdm.send",
						"advertised_caps": []string{"cap.group-dm", "mode.tree"},
						"payload": map[string]any{
							"group_id":   "g1",
							"sender_id":  "carol",
							"ciphertext": "aW52YWxpZA",
						},
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"ok": true},
					map[string]any{"error_code": "OPERATION_FAILED"},
				},
			},
		},
		"friends_kat.json": map[string]any{
			"description": "friends family KAT: request-accept-list flow; rate limit on 11th request",
			"source":      "docs/spec/v0.1/47-family-friends.md",
			"inputs": map[string]any{
				"cases": []any{
					map[string]any{
						"label":          "alice requests bob",
						"operation":      "friends.request",
						"advertised_caps": []string{"cap.friends"},
						"payload": map[string]any{
							"from_peer_id": "alice",
							"to_peer_id":   "bob",
						},
					},
					map[string]any{
						"label":          "bob accepts alice",
						"operation":      "friends.accept",
						"advertised_caps": []string{"cap.friends"},
						"payload": map[string]any{
							"from_peer_id": "alice",
							"to_peer_id":   "bob",
						},
					},
					map[string]any{
						"label":          "alice lists friends (expects bob)",
						"operation":      "friends.list",
						"advertised_caps": []string{"cap.friends"},
						"payload": map[string]any{
							"peer_id": "alice",
						},
					},
					map[string]any{
						"label":       "rate limit: 11th request from spammer",
						"description": "send 10 requests from spammer to different targets, then 11th must fail",
						"operation":   "friends.request",
						"advertised_caps": []string{"cap.friends"},
						"payload": map[string]any{
							"from_peer_id": "spammer",
							"to_peer_id":   "victim11",
						},
						"pre_requests": []any{
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t1"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t2"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t3"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t4"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t5"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t6"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t7"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t8"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t9"},
							map[string]any{"from_peer_id": "spammer", "to_peer_id": "t10"},
						},
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"ok": true, "status": "pending"},
					map[string]any{"ok": true, "status": "accepted"},
					map[string]any{"contains_bob": true},
					map[string]any{"error_code": "RATE_LIMITED"},
				},
			},
		},
		"presence_kat.json": map[string]any{
			"description": "presence family KAT: announce and query; stale version rejected",
			"source":      "docs/spec/v0.1/48-family-presence.md",
			"inputs": map[string]any{
				"cases": []any{
					map[string]any{
						"label":          "announce alice online",
						"operation":      "presence.update",
						"advertised_caps": []string{"cap.presence"},
						"payload": map[string]any{
							"peer_id":        "alice",
							"status":         "online",
							"status_version": 1,
						},
					},
					map[string]any{
						"label":          "query alice presence",
						"operation":      "presence.query",
						"advertised_caps": []string{"cap.presence"},
						"payload": map[string]any{
							"peer_ids": []string{"alice"},
						},
					},
					map[string]any{
						"label":          "stale version rejected: re-announce with same version",
						"operation":      "presence.update",
						"advertised_caps": []string{"cap.presence"},
						"payload": map[string]any{
							"peer_id":        "alice",
							"status":         "away",
							"status_version": 1,
						},
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"ok": true},
					map[string]any{"presence_count": 1, "status": "online"},
					map[string]any{"ok": false, "reason": "stale_status_version"},
				},
			},
		},
		"notify_kat.json": map[string]any{
			"description": "notify family KAT: push and drain; rate limit on 61st push",
			"source":      "docs/spec/v0.1/49-family-notify.md",
			"inputs": map[string]any{
				"cases": []any{
					map[string]any{
						"label":          "push 3 notifications for alice",
						"operation":      "notify.push",
						"advertised_caps": []string{"cap.notify"},
						"count":          3,
						"payload": map[string]any{
							"recipient_id": "alice",
							"sender_id":   "system",
							"type":        "mention",
							"payload":     map[string]any{},
						},
					},
					map[string]any{
						"label":          "drain alice notifications (expect count=3)",
						"operation":      "notify.drain",
						"advertised_caps": []string{"cap.notify"},
						"payload": map[string]any{
							"recipient_id": "alice",
						},
					},
					map[string]any{
						"label":          "drain alice again (expect count=0)",
						"operation":      "notify.drain",
						"advertised_caps": []string{"cap.notify"},
						"payload": map[string]any{
							"recipient_id": "alice",
						},
					},
					map[string]any{
						"label":       "rate limit: 61st push from same sender returns RATE_LIMITED",
						"description": "push 61 notifications rapidly from sender; 61st is rejected",
						"operation":   "notify.push",
						"advertised_caps": []string{"cap.notify"},
						"count":       61,
						"payload": map[string]any{
							"recipient_id": "alice",
							"sender_id":   "spammer",
							"type":        "dm",
							"payload":     map[string]any{},
						},
					},
				},
			},
			"expected_output": map[string]any{
				"cases": []any{
					map[string]any{"ok": true},
					map[string]any{"count": 3},
					map[string]any{"count": 0},
					map[string]any{"error_code": "RATE_LIMITED"},
				},
			},
		},
	}
}

// updatePin reads all JSON files in dir, computes their SHA-256, and writes pin.sha256.
// Existing entries that are still correct are preserved; new/changed entries are updated.
func updatePin(dir string) error {
	pinPath := filepath.Join(dir, "pin.sha256")

	// Read existing pin entries (for ordering stability).
	existingOrder := []string{}
	existingMap := map[string]string{}
	if data, err := os.ReadFile(pinPath); err == nil {
		sc := bufio.NewScanner(strings.NewReader(string(data)))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) == 2 {
				existingMap[parts[1]] = parts[0]
				existingOrder = append(existingOrder, parts[1])
			}
		}
	}

	// Enumerate all JSON files in dir.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("readdir %s: %w", dir, err)
	}

	newMap := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return fmt.Errorf("read %s: %w", e.Name(), err)
		}
		sum := sha256.Sum256(data)
		newMap[e.Name()] = hex.EncodeToString(sum[:])
	}

	// Build output: preserve existing order first, then append new files.
	seen := map[string]bool{}
	var lines []string
	for _, name := range existingOrder {
		if hash, ok := newMap[name]; ok {
			lines = append(lines, hash+"  "+name)
			seen[name] = true
		}
		// If file was deleted, drop it from pin.
	}
	// Collect new filenames not in existing order.
	var newNames []string
	for name := range newMap {
		if !seen[name] {
			newNames = append(newNames, name)
		}
	}
	sort.Strings(newNames)
	for _, name := range newNames {
		lines = append(lines, newMap[name]+"  "+name)
	}

	out := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(pinPath, []byte(out), 0o644)
}

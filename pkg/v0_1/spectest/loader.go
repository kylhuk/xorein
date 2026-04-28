// Package spectest implements the conformance harness KAT loader.
// Source: docs/spec/v0.1/90-conformance-harness.md §2
package spectest

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FlexMap is a JSON object whose values may be strings, numbers, or booleans.
// Non-string values are stored as their JSON text representation.
type FlexMap map[string]string

func (m *FlexMap) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make(FlexMap, len(raw))
	for k, v := range raw {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			out[k] = s
		} else {
			// Numbers, booleans, etc. — store as JSON text.
			out[k] = string(v)
		}
	}
	*m = out
	return nil
}

// Vector is a single known-answer test vector per spec 90 §2.
type Vector struct {
	ID             string  `json:"id"`
	Description    string  `json:"description"`
	Source         string  `json:"source"`
	Inputs         FlexMap `json:"inputs"`
	ExpectedOutput FlexMap `json:"expected_output"`
}

// Name returns the test name (Description preferred, then ID).
func (v Vector) Name() string {
	if v.Description != "" {
		return v.Description
	}
	return v.ID
}

// outerEnvelope handles JSON files whose top level is a wrapper object
// with a "vectors" array inside (e.g. primitive_x25519.json).
type outerEnvelope struct {
	Description string   `json:"description"`
	Source      string   `json:"source"`
	Vectors     []Vector `json:"vectors"`
}

// LoadVectors reads and parses a KAT JSON file. Supports three shapes:
//  1. []Vector — top-level array of vector objects.
//  2. Outer envelope — top-level object with a "vectors" sub-array.
//  3. Single Vector — top-level object that is a single vector.
func LoadVectors(t *testing.T, path string) []Vector {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("spectest: open vector file %s: %v", path, err)
	}

	// Shape 1: top-level array.
	var vecs []Vector
	if err := json.Unmarshal(data, &vecs); err == nil && len(vecs) > 0 {
		return vecs
	}

	// Shape 2: outer envelope with "vectors" sub-array.
	var outer outerEnvelope
	if err := json.Unmarshal(data, &outer); err == nil && len(outer.Vectors) > 0 {
		return outer.Vectors
	}

	// Shape 3: single vector.
	var v Vector
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("spectest: decode %s: %v", path, err)
	}
	return []Vector{v}
}

// Hex decodes a lowercase hex string (no 0x prefix) per spec 90 §2.
func Hex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("spectest: invalid hex %q: %v", s, err))
	}
	return b
}

// VerifyPin reads pin.sha256 from the test-vectors directory and verifies all
// listed files. Should be called from TestMain.
func VerifyPin(t *testing.T, vectorDir string) {
	t.Helper()
	pinPath := filepath.Join(vectorDir, "pin.sha256")
	f, err := os.Open(pinPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Log("spectest: pin.sha256 not found; skipping pin verification")
			return
		}
		t.Fatalf("spectest: open pin file: %v", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			t.Errorf("spectest: malformed pin line: %q", line)
			continue
		}
		expectedHex, filename := parts[0], parts[1]
		fullPath := filepath.Join(vectorDir, filename)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("spectest: read vector file %s: %v", fullPath, err)
			continue
		}
		got := sha256.Sum256(data)
		gotHex := hex.EncodeToString(got[:])
		if gotHex != expectedHex {
			t.Errorf("spectest: pin mismatch for %s:\n  want %s\n  got  %s", filename, expectedHex, gotHex)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("spectest: scan pin file: %v", err)
	}
}

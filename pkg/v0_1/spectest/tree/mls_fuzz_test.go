package tree_test

import (
	"encoding/json"
	"testing"

	tree "github.com/aether/code_aether/pkg/v0_1/mode/tree"
)

// FuzzMLSMessageDecode feeds arbitrary bytes to MLSMessage JSON decode to verify no panics.
func FuzzMLSMessageDecode(f *testing.F) {
	// Seed corpus: a valid MLSMessage.
	valid := tree.MLSMessage{
		Version:     1,
		ContentType: tree.ContentTypeApplication,
		GroupID:     "test-group",
		EpochID:     0,
		Payload:     []byte("hello"),
	}
	validJSON, _ := json.Marshal(valid)
	f.Add(validJSON)

	// Some invalid seeds.
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"content_type":255,"group_id":"","epoch_id":0}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`{"body":"bm90YmFzZTY0ISQ="}`))
	f.Add([]byte{0x00, 0x01, 0xFF, 0xFE})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic — json.Unmarshal into MLSMessage should only return an error.
		var msg tree.MLSMessage
		_ = json.Unmarshal(data, &msg)
	})
}

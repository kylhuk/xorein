package v24

import (
	"testing"

	"github.com/aether/code_aether/pkg/v24/localapi"
)

func FuzzValidateFrame(f *testing.F) {
	f.Add([]byte{}, 0)
	f.Add([]byte{0, 0, 0, 0}, 128)
	f.Fuzz(func(t *testing.T, data []byte, maxPayload int) {
		localapi.ValidateFrame(data, maxPayload)
	})
}

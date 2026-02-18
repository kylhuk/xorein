package v11perf

import (
	"errors"
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
)

func TestRelayPolicyBoundary(t *testing.T) {
	t.Parallel()

	table := []struct {
		name           string
		mode           relaypolicy.PersistenceMode
		expectError    bool
		forbiddenClass relaypolicy.StorageClass
	}{
		{
			name:        "allowed-mode",
			mode:        relaypolicy.PersistenceModeTransientMetadata,
			expectError: false,
		},
		{
			name:           "forbidden-durable-mode",
			mode:           relaypolicy.PersistenceModeDurableMessageBody,
			expectError:    true,
			forbiddenClass: relaypolicy.StorageClassDurableMessageBody,
		},
		{
			name:           "forbidden-attachment-payload",
			mode:           relaypolicy.PersistenceModeAttachmentPayload,
			expectError:    true,
			forbiddenClass: relaypolicy.StorageClassAttachmentPayload,
		},
		{
			name:           "forbidden-media-frame-archive",
			mode:           relaypolicy.PersistenceModeMediaFrameArchive,
			expectError:    true,
			forbiddenClass: relaypolicy.StorageClassMediaFrameArchive,
		},
	}

	for _, tc := range table {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := relaypolicy.ValidateMode(tc.mode)
			if !tc.expectError {
				if err != nil {
					t.Fatalf("expected no error for mode %q, got %v", tc.mode, err)
				}
				return
			}

			var valErr *relaypolicy.ValidationError
			if err == nil || !errors.As(err, &valErr) {
				t.Fatalf("expected ValidationError for mode %q, got %v", tc.mode, err)
			}

			found := false
			for _, cls := range valErr.ForbiddenClasses {
				if cls == tc.forbiddenClass {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected forbidden class %q in error, got %v", tc.forbiddenClass, valErr.ForbiddenClasses)
			}
		})
	}
}

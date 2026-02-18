package relaypolicy

import (
	"errors"
	"strings"
	"testing"
)

func TestAllowedClassesAreDeterministic(t *testing.T) {
	expected := []StorageClass{StorageClassSessionMetadata, StorageClassTransientMetadata}
	got := AllowedClasses()
	if len(got) != len(expected) {
		t.Fatalf("allowed classes length = %d, want %d", len(got), len(expected))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("allowed class at %d = %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestForbiddenClassesAreDeterministic(t *testing.T) {
	expected := []StorageClass{
		StorageClassDurableMessageBody,
		StorageClassAttachmentPayload,
		StorageClassMediaFrameArchive,
	}
	got := ForbiddenClasses()
	if len(got) != len(expected) {
		t.Fatalf("forbidden classes length = %d, want %d", len(got), len(expected))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("forbidden class at %d = %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestValidateModeAllowsSessionMetadata(t *testing.T) {
	mode, err := ParsePersistenceMode("session-metadata")
	if err != nil {
		t.Fatalf("ParsePersistenceMode() error = %v", err)
	}
	if err := ValidateMode(mode); err != nil {
		t.Fatalf("ValidateMode() error = %v", err)
	}
}

func TestValidateModeRejectsDurableMessageBody(t *testing.T) {
	mode := PersistenceModeDurableMessageBody
	err := ValidateMode(mode)
	if err == nil {
		t.Fatalf("ValidateMode() = nil, want ValidationError")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.Mode != mode {
		t.Fatalf("expected mode %q, got %q", mode, validationErr.Mode)
	}
	if len(validationErr.ForbiddenClasses) != 1 || validationErr.ForbiddenClasses[0] != StorageClassDurableMessageBody {
		t.Fatalf("unexpected forbidden classes list: %v", validationErr.ForbiddenClasses)
	}
}

func TestParsePersistenceModeUnknownValue(t *testing.T) {
	if _, err := ParsePersistenceMode("please-do-not-parse"); err == nil {
		t.Fatalf("expected error for unknown mode")
	} else if !strings.Contains(err.Error(), "valid modes") {
		t.Fatalf("unexpected error text: %v", err)
	}
}

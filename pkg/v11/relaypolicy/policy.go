package relaypolicy

import (
	"fmt"
	"sort"
	"strings"
)

// StorageClass represents a relay persistence category.
type StorageClass string

const (
	// StorageClassSessionMetadata is metadata tied to an active relay session.
	StorageClassSessionMetadata StorageClass = "session-metadata"
	// StorageClassTransientMetadata is ephemeral operational metadata.
	StorageClassTransientMetadata StorageClass = "transient-metadata"
	// StorageClassDurableMessageBody represents a stored message body.
	StorageClassDurableMessageBody StorageClass = "durable-message-body"
	// StorageClassAttachmentPayload represents stored attachment payloads.
	StorageClassAttachmentPayload StorageClass = "attachment-payload"
	// StorageClassMediaFrameArchive represents stored media frames or archives.
	StorageClassMediaFrameArchive StorageClass = "media-frame-archive"
)

var (
	allowedClasses = []StorageClass{
		StorageClassSessionMetadata,
		StorageClassTransientMetadata,
	}
	forbiddenClasses = []StorageClass{
		StorageClassDurableMessageBody,
		StorageClassAttachmentPayload,
		StorageClassMediaFrameArchive,
	}
	forbiddenClassSet = func() map[StorageClass]struct{} {
		m := make(map[StorageClass]struct{}, len(forbiddenClasses))
		for _, cls := range forbiddenClasses {
			m[cls] = struct{}{}
		}
		return m
	}()
)

// AllowedClasses reports the deterministic list of storage classes permitted by Phase 2 relay policy.
func AllowedClasses() []StorageClass {
	result := make([]StorageClass, len(allowedClasses))
	copy(result, allowedClasses)
	return result
}

// ForbiddenClasses reports the deterministic list of storage classes that may never be hosted.
func ForbiddenClasses() []StorageClass {
	result := make([]StorageClass, len(forbiddenClasses))
	copy(result, forbiddenClasses)
	return result
}

// PersistenceMode describes a relay persistence intent.
type PersistenceMode string

const (
	PersistenceModeNone               PersistenceMode = "none"
	PersistenceModeSessionMetadata    PersistenceMode = "session-metadata"
	PersistenceModeTransientMetadata  PersistenceMode = "transient-metadata"
	PersistenceModeDurableMessageBody PersistenceMode = "durable-message-body"
	PersistenceModeAttachmentPayload  PersistenceMode = "attachment-payload"
	PersistenceModeMediaFrameArchive  PersistenceMode = "media-frame-archive"
)

var modeToClasses = map[PersistenceMode][]StorageClass{
	PersistenceModeNone: {
		// No classes.
	},
	PersistenceModeSessionMetadata: {
		StorageClassSessionMetadata,
	},
	PersistenceModeTransientMetadata: {
		StorageClassTransientMetadata,
		StorageClassSessionMetadata,
	},
	PersistenceModeDurableMessageBody: {
		StorageClassDurableMessageBody,
	},
	PersistenceModeAttachmentPayload: {
		StorageClassAttachmentPayload,
	},
	PersistenceModeMediaFrameArchive: {
		StorageClassMediaFrameArchive,
	},
}

// ValidationError indicates a persistence mode would persist forbidden classes.
type ValidationError struct {
	Mode             PersistenceMode
	ForbiddenClasses []StorageClass
}

// Error implements error.
func (e ValidationError) Error() string {
	labels := make([]string, len(e.ForbiddenClasses))
	for i, cls := range e.ForbiddenClasses {
		labels[i] = string(cls)
	}
	return fmt.Sprintf("relay persistence mode %q may not store forbidden classes: %s", e.Mode, strings.Join(labels, ", "))
}

// ParsePersistenceMode converts the raw flag value into a PersistenceMode.
func ParsePersistenceMode(raw string) (PersistenceMode, error) {
	normalized := PersistenceMode(strings.ToLower(strings.TrimSpace(raw)))
	if normalized == "" {
		normalized = PersistenceModeNone
	}

	if _, ok := modeToClasses[normalized]; ok {
		return normalized, nil
	}

	modes := make([]string, 0, len(modeToClasses))
	for m := range modeToClasses {
		modes = append(modes, string(m))
	}
	sort.Strings(modes)
	return "", fmt.Errorf("unknown relay persistence mode %q; valid modes: %s", raw, strings.Join(modes, ", "))
}

// ValidateMode enforces that a persistence mode only stores allowed classes.
func ValidateMode(mode PersistenceMode) error {
	classes, ok := modeToClasses[mode]
	if !ok {
		return fmt.Errorf("unexpected persistence mode %q", mode)
	}

	forbidden := make([]StorageClass, 0, len(classes))
	for _, cls := range classes {
		if _, isForbidden := forbiddenClassSet[cls]; isForbidden {
			forbidden = append(forbidden, cls)
		}
	}

	if len(forbidden) == 0 {
		return nil
	}

	return &ValidationError{Mode: mode, ForbiddenClasses: forbidden}
}

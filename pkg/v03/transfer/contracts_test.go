package transfer

import "testing"

func TestChunkCountLimits(t *testing.T) {
	tests := []struct {
		totalBytes int
		expected   int
	}{
		{0, 0},
		{-100, 0},
		{ChunkSizeBytes, 1},
		{ChunkSizeBytes + 1, 2},
		{ChunkSizeBytes * 4, 4},
	}

	for _, tt := range tests {
		if got := ChunkCount(tt.totalBytes); got != tt.expected {
			t.Fatalf("ChunkCount(%d) = %d, want %d", tt.totalBytes, got, tt.expected)
		}
	}
}

func TestIsTransferSizeAllowed(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		expected bool
	}{
		{"zero", 0, false},
		{"negative", -1, false},
		{"one byte", 1, true},
		{"max allowed", MaxTransferSizeMB * 1024 * 1024, true},
		{"over max", MaxTransferSizeMB*1024*1024 + 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTransferSizeAllowed(tt.total); got != tt.expected {
				t.Fatalf("IsTransferSizeAllowed(%d) = %t, want %t", tt.total, got, tt.expected)
			}
		})
	}
}

func TestRetryIntervalDeterminism(t *testing.T) {
	tests := []struct {
		attempt  int
		expected int
	}{
		{-1, 0},
		{0, 0},
		{1, 250},
		{RetryLimitAttempts, RetryLimitAttempts * 250},
		{RetryLimitAttempts + 2, RetryLimitAttempts * 250},
	}

	for _, tt := range tests {
		if got := RetryInterval(tt.attempt); got != tt.expected {
			t.Fatalf("RetryInterval(%d) = %d, want %d", tt.attempt, got, tt.expected)
		}
	}
}

func TestSecurityDisclosureModes(t *testing.T) {
	if got := SecurityDisclosure(TransferSecurityE2EE); got == SecurityDisclosure(TransferSecurityClear) {
		t.Fatalf("security disclosure should differ between modes; got %q", got)
	}
	if got := SecurityDisclosure(TransferSecurityClear); got != "Clear transfer; server can inspect inline media" {
		t.Fatalf("unexpected clear disclosure: %q", got)
	}
}

func TestIntegrityStatus(t *testing.T) {
	tests := []struct {
		retries  int
		expected string
	}{
		{0, "Integrity verified on first attempt"},
		{RetryLimitAttempts - 1, "Integrity recovered via deterministic retry"},
		{RetryLimitAttempts, "Integrity requires manual resume (max retries reached)"},
		{RetryLimitAttempts + 3, "Integrity requires manual resume (max retries reached)"},
	}

	for _, tt := range tests {
		if got := IntegrityStatus(tt.retries); got != tt.expected {
			t.Fatalf("IntegrityStatus(%d) = %q, want %q", tt.retries, got, tt.expected)
		}
	}
}

func TestPresentationForMIME(t *testing.T) {
	tests := []struct {
		mimeType string
		expected AttachmentPresentation
	}{
		{"image/png", PresentationInlineImage},
		{"IMAGE/JPEG", PresentationInlineImage},
		{"application/pdf", PresentationAttachmentCard},
		{"", PresentationAttachmentCard},
	}

	for _, tt := range tests {
		if got := PresentationForMIME(tt.mimeType); got != tt.expected {
			t.Fatalf("PresentationForMIME(%q) = %q, want %q", tt.mimeType, got, tt.expected)
		}
	}
}

func TestAttachmentMetadataDisclosure(t *testing.T) {
	if got := AttachmentMetadataDisclosure(TransferSecurityE2EE); got != "Attachment metadata minimized; payload encrypted with conversation epoch keys" {
		t.Fatalf("unexpected e2ee metadata disclosure: %q", got)
	}
	if got := AttachmentMetadataDisclosure(TransferSecurityClear); got != "Attachment metadata visible to relay/server in clear mode" {
		t.Fatalf("unexpected clear metadata disclosure: %q", got)
	}
}

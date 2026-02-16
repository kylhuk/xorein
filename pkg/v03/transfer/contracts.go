package transfer

import "strings"

const (
	MaxTransferSizeMB  = 25
	ChunkSizeBytes     = 256 * 1024
	RetryLimitAttempts = 5
)

type TransferSecurityMode string

const (
	TransferSecurityE2EE  TransferSecurityMode = "e2ee"
	TransferSecurityClear TransferSecurityMode = "clear"
)

func ChunkCount(totalBytes int) int {
	if totalBytes <= 0 {
		return 0
	}
	chunks := totalBytes / ChunkSizeBytes
	if totalBytes%ChunkSizeBytes != 0 {
		chunks++
	}
	return chunks
}

func IsTransferSizeAllowed(totalBytes int) bool {
	if totalBytes <= 0 {
		return false
	}
	return totalBytes <= MaxTransferSizeMB*1024*1024
}

func RetryInterval(attempt int) int {
	if attempt <= 0 {
		return 0
	}
	if attempt > RetryLimitAttempts {
		attempt = RetryLimitAttempts
	}
	return attempt * 250
}

func SecurityDisclosure(mode TransferSecurityMode) string {
	if mode == TransferSecurityClear {
		return "Clear transfer; server can inspect inline media"
	}
	return "Media E2EE chunk encryption per conversation epoch"
}

func IntegrityStatus(retries int) string {
	if retries == 0 {
		return "Integrity verified on first attempt"
	}
	if retries < RetryLimitAttempts {
		return "Integrity recovered via deterministic retry"
	}
	return "Integrity requires manual resume (max retries reached)"
}

type AttachmentPresentation string

const (
	PresentationInlineImage    AttachmentPresentation = "inline-image"
	PresentationAttachmentCard AttachmentPresentation = "attachment-card"
)

func PresentationForMIME(mimeType string) AttachmentPresentation {
	if strings.HasPrefix(strings.ToLower(mimeType), "image/") {
		return PresentationInlineImage
	}
	return PresentationAttachmentCard
}

func AttachmentMetadataDisclosure(mode TransferSecurityMode) string {
	if mode == TransferSecurityE2EE {
		return "Attachment metadata minimized; payload encrypted with conversation epoch keys"
	}
	return "Attachment metadata visible to relay/server in clear mode"
}

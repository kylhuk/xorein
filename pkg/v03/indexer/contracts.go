package indexer

type IndexerPosture string

const (
	IndexerPostureOptional  IndexerPosture = "optional"
	IndexerPostureCommunity IndexerPosture = "community-run"
	IndexerPostureNonAuth   IndexerPosture = "non-authoritative"
)

type SignatureVerificationState string

const (
	SignatureValid   SignatureVerificationState = "valid"
	SignatureMissing SignatureVerificationState = "missing"
	SignatureStale   SignatureVerificationState = "stale"
	SignatureInvalid SignatureVerificationState = "invalid"
)

func VerificationDecision(state SignatureVerificationState) string {
	switch state {
	case SignatureValid:
		return "Accept payload; signature verified"
	case SignatureMissing:
		return "Treat payload as advisory; fall back to authoritative directory"
	case SignatureStale:
		return "Signal stale data; request fresh proof"
	default:
		return "Reject payload; signature verification failed"
	}
}

func MultiIndexerMergeHint(authoritative bool) string {
	if authoritative {
		return "Honor authoritative source (if present) before community-run indexes"
	}
	return "Merge sorted by deterministic fingerprint and drop duplicates"
}

package security

// RefusalReasonKeywordBackfillNotAllowed indicates keyword-bearing requests are off by default.
const RefusalReasonKeywordBackfillNotAllowed RefusalReason = "keyword_backfill_not_allowed"

// TimeRange describes the requested backfill window.
type TimeRange struct {
	Start int64
	End   int64
}

// BackfillRequest models the data carried inside a history backfill call.
type BackfillRequest struct {
	SpaceID   string
	ChannelID string
	Range     TimeRange
	MaxRanges int
	Query     string
	Keywords  []string
}

// PrivacyConfig gates keyword-bearing requests unless explicitly allowed.
type PrivacyConfig struct {
	AllowKeywordBackfill bool
}

// ValidateBackfillRequest enforces keyword blocking based on the config.
func (c PrivacyConfig) ValidateBackfillRequest(req BackfillRequest) error {
	if c.AllowKeywordBackfill {
		return nil
	}
	if req.HasKeywordIntent() {
		return &RefusalError{
			Reason:  RefusalReasonKeywordBackfillNotAllowed,
			Details: "keyword-bearing backfill blocked by default",
		}
	}
	return nil
}

// HasKeywordIntent reports whether the request carries keyword/query intent.
func (r BackfillRequest) HasKeywordIntent() bool {
	if r.Query != "" {
		return true
	}
	for _, keyword := range r.Keywords {
		if keyword != "" {
			return true
		}
	}
	return false
}

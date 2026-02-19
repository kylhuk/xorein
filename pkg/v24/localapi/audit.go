package localapi

import "time"

// AuditRecord captures sensitive RPC details without payload content.
type AuditRecord struct {
	Timestamp time.Time
	RPC       string
	SpaceID   string
	ChannelID string
	Reason    RefusalReason
	Outcome   string
}

// Summary returns a deterministic audit summary for logging.
func (a AuditRecord) Summary() string {
	return a.Timestamp.UTC().Format(time.RFC3339) + " " + a.RPC + " reason=" + a.Reason.String() + " outcome=" + a.Outcome + " space=" + a.SpaceID + " channel=" + a.ChannelID
}

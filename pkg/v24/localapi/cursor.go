package localapi

import "fmt"

// EventCursor tracks a deterministic position for stream resumption.
type EventCursor struct {
	Position    uint64
	ResumeToken string
}

// Advance returns the next cursor and resumption token without payload data.
func (c EventCursor) Advance(delta uint64) EventCursor {
	next := c.Position + delta
	return EventCursor{
		Position:    next,
		ResumeToken: fmt.Sprintf("cursor-%d", next),
	}
}

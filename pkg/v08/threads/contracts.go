package threads

import "fmt"

// Lifecycle labels the deterministic state of a conversation thread.
type Lifecycle string

const (
	LifecycleActive   Lifecycle = "active"
	LifecycleDormant  Lifecycle = "dormant"
	LifecycleArchived Lifecycle = "archived"
)

// ThreadTrace captures reply metadata for a given thread.
type ThreadTrace struct {
	ID           string
	CreatedDepth int
	ReplyDepth   int
	Lifecycle    Lifecycle
}

// ValidateReplyLineage enforces that replies progress monotonically.
func ValidateReplyLineage(trace ThreadTrace) error {
	if trace.ID == "" {
		return fmt.Errorf("thread trace missing ID")
	}
	if trace.ReplyDepth < 0 {
		return fmt.Errorf("reply depth %d must be non-negative", trace.ReplyDepth)
	}
	if trace.ReplyDepth > trace.CreatedDepth+1 {
		return fmt.Errorf("reply depth %d exceeds allowed growth from depth %d", trace.ReplyDepth, trace.CreatedDepth)
	}
	return nil
}

// ClassifyLifecycle returns the lifecycle label a thread should obey.
func ClassifyLifecycle(depth int) Lifecycle {
	switch {
	case depth == 0:
		return LifecycleDormant
	case depth < 4:
		return LifecycleActive
	default:
		return LifecycleArchived
	}
}

// DescribeLifecycle renders the lifecycle label for telemetry.
func DescribeLifecycle(lifecycle Lifecycle) string {
	return fmt.Sprintf("thread-lifecycle:%s", lifecycle)
}

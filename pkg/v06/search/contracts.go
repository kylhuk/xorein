package search

import (
	"fmt"
	"sort"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

type SearchReasonClass string

const (
	SearchReasonPartial SearchReasonClass = "VA-S1:search.partial"
	SearchReasonExplore SearchReasonClass = "VA-S2:search.explore"
	SearchReasonPreview SearchReasonClass = "VA-S3:search.preview"
)

type SearchContract struct {
	ScopeID  string
	TaskID   string
	Artifact string
	Gate     conformance.GateID
	Reason   SearchReasonClass
}

func NewSearchContract(scope, task, artifact string, gate conformance.GateID, reason SearchReasonClass) SearchContract {
	return SearchContract{
		ScopeID:  scope,
		TaskID:   task,
		Artifact: artifact,
		Gate:     gate,
		Reason:   reason,
	}
}

func (c SearchContract) EvidenceAnchor() string {
	return fmt.Sprintf("%s#%s", c.Artifact, c.TaskID)
}

func (c SearchContract) ReasonLabel() string {
	return string(c.Reason)
}

func SearchReasonClasses() []SearchReasonClass {
	return []SearchReasonClass{
		SearchReasonPartial,
		SearchReasonExplore,
		SearchReasonPreview,
	}
}

type PartialFailureStatus string

const (
	PartialFailureSuccess  PartialFailureStatus = "search.partial.success"
	PartialFailureFailure  PartialFailureStatus = "search.partial.failure"
	PartialFailureFallback PartialFailureStatus = "search.partial.fallback"
)

type PartialFailureClassification struct {
	Status        PartialFailureStatus
	Successful    int
	Total         int
	ReasonLabel   string
	NeedsFallback bool
}

func ClassifyPartialFailure(successful, total int) PartialFailureClassification {
	status := PartialFailureSuccess
	needsFallback := false
	reason := string(PartialFailureSuccess)
	if successful < total {
		status = PartialFailureFailure
		needsFallback = true
		reason = string(PartialFailureFailure)
	}
	if successful == 0 {
		status = PartialFailureFallback
		reason = string(PartialFailureFallback)
	}
	return PartialFailureClassification{
		Status:        status,
		Successful:    successful,
		Total:         total,
		ReasonLabel:   reason,
		NeedsFallback: needsFallback,
	}
}

type FeedItem struct {
	ID        string
	Freshness time.Time
	Priority  int
}

func OrderExploreFeed(items []FeedItem, freshnessCutoff time.Time) []FeedItem {
	sorted := make([]FeedItem, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Freshness.Before(sorted[j].Freshness) {
			return true
		}
		if sorted[i].Freshness.After(sorted[j].Freshness) {
			return false
		}
		return sorted[i].Priority < sorted[j].Priority
	})
	for index, item := range sorted {
		if item.Freshness.After(freshnessCutoff) {
			sorted[index].Priority += 1
		}
	}
	return sorted
}

type PreviewMismatchDecision struct {
	Matched   bool
	Reason    string
	Recovery  string
	AnchorIDs [2]string
}

func DecidePreviewMismatch(expected, actual string) PreviewMismatchDecision {
	if expected == actual {
		return PreviewMismatchDecision{
			Matched:   true,
			Reason:    string(SearchReasonPreview) + ".success",
			Recovery:  "search.preview.align",
			AnchorIDs: [2]string{expected, actual},
		}
	}
	return PreviewMismatchDecision{
		Matched:   false,
		Reason:    string(SearchReasonPreview) + ".mismatch",
		Recovery:  "search.preview.align",
		AnchorIDs: [2]string{expected, actual},
	}
}

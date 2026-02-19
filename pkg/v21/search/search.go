package search

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

const defaultQueryLimit = 64

var ErrSearchQueryTimeout = errors.New("SEARCH_QUERY_TIMEOUT")

type Document struct {
	ID        string
	Channel   string
	Sender    string
	Timestamp time.Time
	Body      string
}

type QueryOptions struct {
	Channel string
	Sender  string
	Since   time.Time
	Until   time.Time
	Limit   int
}

type CoverageWindow struct {
	Status string
	Label  string
	Since  time.Time
	Until  time.Time
}

type Index struct {
	mu       sync.RWMutex
	docs     map[string]Document
	coverage map[string]*channelCoverage
}

type channelCoverage struct {
	since time.Time
	until time.Time
	count int
}

func NewIndex() *Index {
	return &Index{
		docs:     make(map[string]Document),
		coverage: make(map[string]*channelCoverage),
	}
}

func (idx *Index) Add(doc Document) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.docs[doc.ID] = doc
	idx.updateCoverage(doc)
}

func (idx *Index) Remove(id string) bool {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	entry, ok := idx.docs[id]
	if !ok {
		return false
	}
	channel := entry.Channel
	delete(idx.docs, id)
	idx.rebuildCoverage(channel)
	return true
}

func (idx *Index) Coverage(channel string) (CoverageWindow, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	if cov, ok := idx.coverage[channel]; ok {
		return cov.window(), true
	}
	return CoverageWindow{Status: "COVERAGE_EMPTY", Label: "No documents indexed"}, false
}

func (idx *Index) Search(ctx context.Context, opts QueryOptions) ([]Document, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultQueryLimit
	}
	idx.mu.RLock()
	filtered := make([]Document, 0, len(idx.docs))
	for _, doc := range idx.docs {
		if opts.Channel != "" && doc.Channel != opts.Channel {
			continue
		}
		if opts.Sender != "" && doc.Sender != opts.Sender {
			continue
		}
		if !opts.Since.IsZero() && doc.Timestamp.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && doc.Timestamp.After(opts.Until) {
			continue
		}
		filtered = append(filtered, doc)
	}
	idx.mu.RUnlock()
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Timestamp.Equal(filtered[j].Timestamp) {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].Timestamp.Before(filtered[j].Timestamp)
	})
	if len(filtered) > limit {
		return nil, ErrSearchQueryTimeout
	}
	res := make([]Document, 0, len(filtered))
	for _, doc := range filtered {
		select {
		case <-ctx.Done():
			return nil, ErrSearchQueryTimeout
		default:
		}
		res = append(res, doc)
	}
	return res, nil
}

func (idx *Index) updateCoverage(doc Document) {
	channel := doc.Channel
	cov, ok := idx.coverage[channel]
	if !ok {
		cov = &channelCoverage{since: doc.Timestamp, until: doc.Timestamp, count: 0}
		idx.coverage[channel] = cov
	}
	if cov.count == 0 || doc.Timestamp.Before(cov.since) {
		cov.since = doc.Timestamp
	}
	if cov.count == 0 || doc.Timestamp.After(cov.until) {
		cov.until = doc.Timestamp
	}
	cov.count++
}

func (idx *Index) rebuildCoverage(channel string) {
	min, max := time.Time{}, time.Time{}
	count := 0
	for _, doc := range idx.docs {
		if doc.Channel != channel {
			continue
		}
		if count == 0 || doc.Timestamp.Before(min) {
			min = doc.Timestamp
		}
		if count == 0 || doc.Timestamp.After(max) {
			max = doc.Timestamp
		}
		count++
	}
	if count == 0 {
		delete(idx.coverage, channel)
		return
	}
	idx.coverage[channel] = &channelCoverage{since: min, until: max, count: count}
}

func (cov *channelCoverage) window() CoverageWindow {
	if cov.count == 0 {
		return CoverageWindow{Status: "COVERAGE_EMPTY", Label: "No documents indexed"}
	}
	label := "Coverage spans a single entry"
	status := "COVERAGE_PARTIAL"
	if cov.count > 1 || !cov.since.Equal(cov.until) {
		status = "COVERAGE_FULL"
		label = "Indexed range from " + cov.since.UTC().Format(time.RFC3339) + " to " + cov.until.UTC().Format(time.RFC3339)
	}
	return CoverageWindow{Status: status, Label: label, Since: cov.since, Until: cov.until}
}

package v21

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/search"
)

func TestLargeDBQueryLimit(t *testing.T) {
	idx := search.NewIndex()
	for i := 0; i < 10; i++ {
		idx.Add(search.Document{ID: fmt.Sprintf("perf-%d", i), Channel: "perf", Sender: "load", Timestamp: time.Now().Add(time.Duration(i) * time.Second), Body: "data"})
	}
	_, err := idx.Search(context.Background(), search.QueryOptions{Limit: 5})
	if err != search.ErrSearchQueryTimeout {
		t.Fatalf("expected limit guard, got %v", err)
	}
}

func TestLargeDBQueryTimeout(t *testing.T) {
	idx := search.NewIndex()
	idx.Add(search.Document{ID: "perf", Channel: "perf", Sender: "load", Timestamp: time.Now(), Body: "data"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := idx.Search(ctx, search.QueryOptions{Limit: 10})
	if err != search.ErrSearchQueryTimeout {
		t.Fatalf("expected timeout guard, got %v", err)
	}
}

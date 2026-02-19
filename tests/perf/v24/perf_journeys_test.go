package v24

import (
	"context"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v24/harmolyn/attach"
	"github.com/aether/code_aether/pkg/v24/harmolyn/journeys"
	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestPerfJourneysCoordinatorSequence(t *testing.T) {
	ctx := context.Background()
	token := localapi.SessionToken{Value: "perf-token", ExpiresAt: time.Now().Add(10 * time.Minute)}
	coord := journeys.NewCoordinator(&stubAttach{result: attach.AttachResult{Token: token, Attached: true}}, &recordingClient{})

	client := coord.Client.(*recordingClient)

	if err := coord.Send(ctx, "payload"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if got, err := coord.Read(ctx, "cursor"); err != nil {
		t.Fatalf("Read failed: %v", err)
	} else if got != "cursor-value" {
		t.Fatalf("unexpected read payload: %s", got)
	}
	if results, err := coord.Search(ctx, "query"); err != nil {
		t.Fatalf("Search failed: %v", err)
	} else if len(results) != 1 || results[0] != "search-result" {
		t.Fatalf("unexpected search results: %v", results)
	}
	if history, err := coord.FetchHistory(ctx, journeys.HistoryRequest{ChannelID: "chan", Limit: 1}); err != nil {
		t.Fatalf("FetchHistory failed: %v", err)
	} else if len(history) != 1 || history[0] != "history-entry" {
		t.Fatalf("unexpected history entries: %v", history)
	}
	if err := coord.MediaControl(ctx, "pause"); err != nil {
		t.Fatalf("MediaControl failed: %v", err)
	}

	expectedOps := []string{"send", "read", "search", "fetch-history", "media-control"}
	if len(client.operations) != len(expectedOps) {
		t.Fatalf("operation count mismatch: got %d ops", len(client.operations))
	}
	for idx, op := range expectedOps {
		if client.operations[idx] != op {
			t.Fatalf("operation %d: expected %s got %s", idx, op, client.operations[idx])
		}
	}
	for _, recorded := range client.tokens {
		if recorded.Value != token.Value {
			t.Fatalf("token drifted: %s", recorded.Value)
		}
	}
}

func TestPerfJourneysDegradedReasonStability(t *testing.T) {
	ctx := context.Background()
	failure := &attach.Failure{Reason: attach.FailureReasonDaemonStartFailed, NextAction: attach.NextActionRetry, Detail: "daemon offline"}
	coord := journeys.NewCoordinator(&stubAttach{result: attach.AttachResult{Failure: failure}}, &recordingClient{})

	cases := []struct {
		name string
		call func() error
	}{
		{name: "send", call: func() error { return coord.Send(ctx, "payload") }},
		{name: "read", call: func() error { _, err := coord.Read(ctx, "cursor"); return err }},
		{name: "search", call: func() error { _, err := coord.Search(ctx, "query"); return err }},
		{name: "history", call: func() error {
			_, err := coord.FetchHistory(ctx, journeys.HistoryRequest{ChannelID: "chan", Limit: 1})
			return err
		}},
		{name: "media", call: func() error { return coord.MediaControl(ctx, "resume") }},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for iter := 0; iter < 8; iter++ {
				err := tc.call()
				if err == nil {
					t.Fatalf("expected error on iteration %d", iter)
				}
				je, ok := err.(*journeys.JourneyError)
				if !ok {
					t.Fatalf("iteration %d: expected JourneyError, got %T", iter, err)
				}
				if je.Reason != failure.Reason || je.NextAction != failure.NextAction {
					t.Fatalf("iteration %d: unexpected reason/action %s/%s", iter, je.Reason, je.NextAction)
				}
			}
		})
	}
}

type stubAttach struct {
	result attach.AttachResult
}

func (s *stubAttach) Attach(ctx context.Context) attach.AttachResult {
	return s.result
}

type recordingClient struct {
	operations []string
	tokens     []localapi.SessionToken
}

func (c *recordingClient) record(op string, token localapi.SessionToken) {
	c.operations = append(c.operations, op)
	c.tokens = append(c.tokens, token)
}

func (c *recordingClient) Send(ctx context.Context, token localapi.SessionToken, payload string) error {
	c.record("send", token)
	return nil
}

func (c *recordingClient) Read(ctx context.Context, token localapi.SessionToken, cursor string) (string, error) {
	c.record("read", token)
	return "cursor-value", nil
}

func (c *recordingClient) Search(ctx context.Context, token localapi.SessionToken, query string) ([]string, error) {
	c.record("search", token)
	return []string{"search-result"}, nil
}

func (c *recordingClient) FetchHistory(ctx context.Context, token localapi.SessionToken, req journeys.HistoryRequest) ([]string, error) {
	c.record("fetch-history", token)
	return []string{"history-entry"}, nil
}

func (c *recordingClient) MediaControl(ctx context.Context, token localapi.SessionToken, action string) error {
	c.record("media-control", token)
	return nil
}

package v24

import (
	"context"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v24/harmolyn/attach"
	"github.com/aether/code_aether/pkg/v24/harmolyn/journeys"
	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestJourneysFlowRoutesThroughLocalAPI(t *testing.T) {
	ctx := context.Background()
	token := localapi.SessionToken{Value: "flow-token", ExpiresAt: time.Now().Add(time.Hour)}
	attachResult := attach.AttachResult{Attached: true, Token: token}
	client := &flowClient{
		readReply:    "timeline-body",
		searchReply:  []string{"match"},
		historyReply: []string{"evt-1"},
	}
	coord := journeys.NewCoordinator(&flowAttach{result: attachResult}, client)

	if err := coord.Send(ctx, "message"); err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if client.lastSendPayload != "message" {
		t.Fatalf("unexpected send payload: %q", client.lastSendPayload)
	}

	got, err := coord.Read(ctx, "cursor")
	if err != nil || got != "timeline-body" {
		t.Fatalf("read result mismatch: %q %v", got, err)
	}

	results, err := coord.Search(ctx, "query")
	if err != nil || len(results) != 1 || results[0] != "match" {
		t.Fatalf("search results unexpected: %v %v", results, err)
	}

	history, err := coord.FetchHistory(ctx, journeys.HistoryRequest{ChannelID: "chan", Limit: 1})
	if err != nil || len(history) != 1 {
		t.Fatalf("history fetch unexpected: %v %v", history, err)
	}

	if err := coord.MediaControl(ctx, "pause"); err != nil {
		t.Fatalf("media control failed: %v", err)
	}
}

type flowAttach struct {
	result attach.AttachResult
}

func (f *flowAttach) Attach(ctx context.Context) attach.AttachResult {
	return f.result
}

type flowClient struct {
	readReply       string
	searchReply     []string
	historyReply    []string
	lastSendPayload string
}

func (f *flowClient) Send(ctx context.Context, token localapi.SessionToken, payload string) error {
	f.lastSendPayload = payload
	return nil
}

func (f *flowClient) Read(ctx context.Context, token localapi.SessionToken, cursor string) (string, error) {
	return f.readReply, nil
}

func (f *flowClient) Search(ctx context.Context, token localapi.SessionToken, query string) ([]string, error) {
	return f.searchReply, nil
}

func (f *flowClient) FetchHistory(ctx context.Context, token localapi.SessionToken, req journeys.HistoryRequest) ([]string, error) {
	return f.historyReply, nil
}

func (f *flowClient) MediaControl(ctx context.Context, token localapi.SessionToken, action string) error {
	return nil
}

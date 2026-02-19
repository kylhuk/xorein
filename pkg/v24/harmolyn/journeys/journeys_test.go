package journeys

import (
	"context"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v24/harmolyn/attach"
	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestCoordinatorSendRoutesThroughLocalAPI(t *testing.T) {
	ctx := context.Background()
	token := localapi.SessionToken{Value: "token", ExpiresAt: time.Now().Add(time.Hour)}
	attachResult := attach.AttachResult{Attached: true, Token: token}
	client := &fakeLocalClient{}
	coord := NewCoordinator(&fakeAttach{result: attachResult}, client)

	if err := coord.Send(ctx, "payload"); err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !client.sendCalled {
		t.Fatalf("send should hit local API")
	}
	if _, err := coord.Read(ctx, "cursor-1"); err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if _, err := coord.Search(ctx, "query"); err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if _, err := coord.FetchHistory(ctx, HistoryRequest{ChannelID: "chan", Limit: 5}); err != nil {
		t.Fatalf("history fetch failed: %v", err)
	}
	if err := coord.MediaControl(ctx, "pause"); err != nil {
		t.Fatalf("media control failed: %v", err)
	}

	if client.sendToken != token {
		t.Fatalf("expected token to reach client, got %v", client.sendToken)
	}
}

func TestCoordinatorDegradedPathReturnsReasonWithoutClientCall(t *testing.T) {
	ctx := context.Background()
	failure := &attach.Failure{
		Reason:     attach.FailureReasonSocketPermissionDenied,
		NextAction: attach.NextActionOpenLogs,
		Detail:     "permission denied",
	}
	coord := NewCoordinator(&fakeAttach{result: attach.AttachResult{Failure: failure}}, &fakeLocalClient{})

	_, err := coord.Search(ctx, "query")
	if err == nil {
		t.Fatalf("expected search to fail when attach degraded")
	}
	journeyErr, ok := err.(*JourneyError)
	if !ok {
		t.Fatalf("expected JourneyError, got %T", err)
	}
	if journeyErr.Reason != failure.Reason {
		t.Fatalf("unexpected reason: %v", journeyErr.Reason)
	}
	if journeyErr.NextAction != failure.NextAction {
		t.Fatalf("unexpected action: %v", journeyErr.NextAction)
	}
	client := coord.Client.(*fakeLocalClient)
	if client.searchCalled {
		t.Fatalf("search should not hit local API when attach degraded")
	}
}

type fakeAttach struct {
	result attach.AttachResult
}

func (f *fakeAttach) Attach(ctx context.Context) attach.AttachResult {
	return f.result
}

type fakeLocalClient struct {
	sendCalled    bool
	readCalled    bool
	searchCalled  bool
	historyCalled bool
	controlCalled bool
	sendToken     localapi.SessionToken
}

func (f *fakeLocalClient) Send(ctx context.Context, token localapi.SessionToken, payload string) error {
	f.sendCalled = true
	f.sendToken = token
	return nil
}

func (f *fakeLocalClient) Read(ctx context.Context, token localapi.SessionToken, cursor string) (string, error) {
	f.readCalled = true
	return "ok", nil
}

func (f *fakeLocalClient) Search(ctx context.Context, token localapi.SessionToken, query string) ([]string, error) {
	f.searchCalled = true
	return []string{"result"}, nil
}

func (f *fakeLocalClient) FetchHistory(ctx context.Context, token localapi.SessionToken, req HistoryRequest) ([]string, error) {
	f.historyCalled = true
	return []string{"history"}, nil
}

func (f *fakeLocalClient) MediaControl(ctx context.Context, token localapi.SessionToken, action string) error {
	f.controlCalled = true
	return nil
}

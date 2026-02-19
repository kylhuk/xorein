package v24

import (
	"context"
	"testing"

	"github.com/aether/code_aether/pkg/v24/harmolyn/attach"
	"github.com/aether/code_aether/pkg/v24/harmolyn/journeys"
	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestJourneysDegradedReasonMatrix(t *testing.T) {
	ctx := context.Background()
	matrix := []struct {
		name   string
		reason attach.FailureReason
		action attach.NextAction
		detail string
	}{
		{name: "daemon start", reason: attach.FailureReasonDaemonStartFailed, action: attach.NextActionRetry, detail: "start failed"},
		{name: "daemon incompatible", reason: attach.FailureReasonDaemonIncompatible, action: attach.NextActionRepair, detail: "version mismatch"},
		{name: "auth failed", reason: attach.FailureReasonAuthFailed, action: attach.NextActionReset, detail: "token invalid"},
		{name: "permission denied", reason: attach.FailureReasonSocketPermissionDenied, action: attach.NextActionOpenLogs, detail: "acl"},
	}

	for _, tc := range matrix {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			failure := &attach.Failure{Reason: tc.reason, NextAction: tc.action, Detail: tc.detail}
			attachResult := attach.AttachResult{Failure: failure}
			client := &degradedClient{}
			coord := journeys.NewCoordinator(&degradedAttach{result: attachResult}, client)

			err := coord.MediaControl(ctx, "resume")
			if err == nil {
				t.Fatalf("expected degraded error for %s", tc.name)
			}
			je, ok := err.(*journeys.JourneyError)
			if !ok {
				t.Fatalf("expected JourneyError, got %T", err)
			}
			if je.Reason != tc.reason {
				t.Fatalf("unexpected reason: %s", je.Reason)
			}
			if je.NextAction != tc.action {
				t.Fatalf("unexpected next action: %s", je.NextAction)
			}
			if client.called {
				t.Fatalf("client should not be called when attach %s", tc.name)
			}
		})
	}
}

type degradedAttach struct {
	result attach.AttachResult
}

func (d *degradedAttach) Attach(ctx context.Context) attach.AttachResult {
	return d.result
}

type degradedClient struct {
	called bool
}

func (d *degradedClient) Send(ctx context.Context, token localapi.SessionToken, payload string) error {
	d.called = true
	return nil
}

func (d *degradedClient) Read(ctx context.Context, token localapi.SessionToken, cursor string) (string, error) {
	d.called = true
	return "", nil
}

func (d *degradedClient) Search(ctx context.Context, token localapi.SessionToken, query string) ([]string, error) {
	d.called = true
	return nil, nil
}

func (d *degradedClient) FetchHistory(ctx context.Context, token localapi.SessionToken, req journeys.HistoryRequest) ([]string, error) {
	d.called = true
	return nil, nil
}

func (d *degradedClient) MediaControl(ctx context.Context, token localapi.SessionToken, action string) error {
	d.called = true
	return nil
}

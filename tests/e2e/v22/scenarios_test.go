package v22

import (
	"errors"
	"testing"

	"github.com/aether/code_aether/pkg/v22/archivist/replicate"
	"github.com/aether/code_aether/pkg/v22/archivist/store"
	"github.com/aether/code_aether/pkg/v22/history/backfill"
	"github.com/aether/code_aether/pkg/v22/history/retrieve"
)

func TestScenarioOfflineCatchup(t *testing.T) {
	manager := backfill.NewManager(3)
	req := backfill.BackfillRequest{
		SpaceID:   "space",
		ChannelID: "chan",
		Range:     backfill.TimeRange{Start: 0, End: 10},
	}
	attempts := 0
	fetch := func() ([]backfill.Segment, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("offline")
		}
		return []backfill.Segment{{ID: "seg"}}, nil
	}
	apply := func(backfill.Segment) error { return nil }

	if err := manager.Backfill(req, fetch, apply); err != backfill.ErrBackfillIncomplete {
		t.Fatalf("expected offline rejection, got %v", err)
	}
	if err := manager.Backfill(req, fetch, apply); err != nil {
		t.Fatalf("offline reattempt failed: %v", err)
	}
	if report := manager.Progress(req); !report.Completed {
		t.Fatalf("progress report did not mark completion: %+v", report)
	}
}

func TestScenarioMultiArchivistFailover(t *testing.T) {
	policy := replicate.Policy{R: 3, RMin: 2}
	result := replicate.Replicate(policy, []replicate.EndpointID{"primary", "secondary", "tertiary"}, func(id replicate.EndpointID) error {
		if id == "primary" {
			return errors.New("primary down")
		}
		return nil
	})
	if result.Health != replicate.HealthDegraded || result.Reason != replicate.ResultReplicaWritePartial {
		t.Fatalf("unexpected failover result %+v", result)
	}
	healing := replicate.Heal(policy, []replicate.EndpointID{"secondary"}, []replicate.EndpointID{"tertiary"}, func(replicate.EndpointID) error {
		return nil
	})
	if len(healing.HealedTokens) == 0 {
		t.Fatalf("expected healing token, got %+v", healing)
	}
}

func TestScenarioQuotaRefusal(t *testing.T) {
	s := store.NewStore(store.Config{QuotaPerSpace: map[store.SpaceID]int64{"space": 1}})
	if err := s.Put("space", "chan", "seg", 2); err == nil {
		t.Fatalf("expected quota refusal")
	}
}

func TestScenarioReplicaHealing(t *testing.T) {
	policy := replicate.Policy{R: 2, RMin: 1}
	healing := replicate.Heal(policy, []replicate.EndpointID{"a"}, []replicate.EndpointID{"b", "c"}, func(replicate.EndpointID) error {
		return nil
	})
	if len(healing.HealedTokens) == 0 {
		t.Fatalf("expected healing tokens, got %+v", healing)
	}
}

func TestScenarioRelayNoHistoryHosting(t *testing.T) {
	store := retrieve.NewRetrievalStore(1)
	if _, err := store.RetrieveHead(retrieve.RetrievalRequest{SpaceID: "space", ChannelID: "chan", Key: "any"}); err != retrieve.ErrRetrievalFailure {
		t.Fatalf("relay should not host history: %v", err)
	}
}

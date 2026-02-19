package v22

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aether/code_aether/pkg/v22/archivist/replicate"
	"github.com/aether/code_aether/pkg/v22/archivist/store"
	"github.com/aether/code_aether/pkg/v22/history/retrieve"
)

func TestPerfQuotaEnforcementBurst(t *testing.T) {
	s := store.NewStore(store.Config{QuotaPerSpace: map[store.SpaceID]int64{"space": 10}})
	var lastErr error
	for i := 0; i < 25; i++ {
		lastErr = s.Put("space", "chan", store.SegmentID(fmt.Sprintf("seg-%d", i)), 1)
		if lastErr != nil {
			break
		}
	}
	if lastErr == nil {
		t.Fatalf("expected quota error but did not observe one")
	}
	var se store.StoreError
	if !errors.As(lastErr, &se) || se.Reason != store.ReasonQuotaExceeded {
		t.Fatalf("unexpected quota error %v", lastErr)
	}
}

func TestPerfAntiEnumerationBurst(t *testing.T) {
	store := retrieve.NewRetrievalStore(0)
	req := retrieve.RetrievalRequest{SpaceID: "space", ChannelID: "chan", Key: "bad"}
	for i := 0; i < 50; i++ {
		if _, err := store.RetrieveHead(req); err != retrieve.ErrRetrievalFailure {
			t.Fatalf("expected consistent failure, got %v", err)
		}
	}
}

func TestPerfReplicaHealingCycles(t *testing.T) {
	policy := replicate.Policy{R: 2, RMin: 1}
	for i := 0; i < 20; i++ {
		result := replicate.Replicate(policy, []replicate.EndpointID{"a", "b"}, func(id replicate.EndpointID) error {
			if id == "a" && i%2 == 0 {
				return errors.New("rotate")
			}
			return nil
		})
		if result.Health == "" {
			t.Fatalf("unexpected empty health")
		}
		healing := replicate.Heal(policy, []replicate.EndpointID{"a"}, []replicate.EndpointID{"c"}, func(replicate.EndpointID) error {
			return nil
		})
		if healing.SuccessTotal == 0 {
			t.Fatalf("expected healing success, got %+v", healing)
		}
	}
}

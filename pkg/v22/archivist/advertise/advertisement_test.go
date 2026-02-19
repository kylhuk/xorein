package advertise

import (
	"testing"
	"time"
)

func TestSelectAdvertisementPrefersSpaceAndQuota(t *testing.T) {
	now := time.Now()
	ads := []Advertisement{
		{ID: "a1", Space: "space-a", Operator: "op-a", Available: true, QuotaRemaining: 50, TTL: time.Hour, LastRefreshed: now, PolicyAllowed: true},
		{ID: "a2", Space: "space-b", Operator: "op-b", Available: true, QuotaRemaining: 200, TTL: time.Hour, LastRefreshed: now.Add(-time.Minute), PolicyAllowed: true},
		{ID: "a3", Space: "space-a", Operator: "op-c", Available: true, QuotaRemaining: 100, TTL: time.Hour, LastRefreshed: now.Add(-time.Minute * 2), PolicyAllowed: true},
	}

	result := SelectAdvertisement(ads, "space-a", 10, now)
	if !result.Selected {
		t.Fatalf("expected selection")
	}
	if result.Advertisement.ID != "a3" {
		t.Fatalf("expected highest quota same space, got %s", result.Advertisement.ID)
	}
}

func TestSelectAdvertisementPolicyDenied(t *testing.T) {
	now := time.Now()
	ads := []Advertisement{
		{ID: "policy", PolicyAllowed: false},
	}
	result := SelectAdvertisement(ads, "space-a", 0, now)
	if result.Selected {
		t.Fatalf("expected no selection")
	}
	if result.Reason != RefusalPolicyDenied {
		t.Fatalf("expected policy denial, got %s", result.Reason)
	}
}

func TestSelectAdvertisementPolicyAllowedButUnavailable(t *testing.T) {
	now := time.Now()
	ads := []Advertisement{
		{ID: "denied", PolicyAllowed: false},
		{ID: "allowed", PolicyAllowed: true, Available: false},
	}
	result := SelectAdvertisement(ads, "space-a", 0, now)
	if result.Selected {
		t.Fatalf("expected no selection")
	}
	if result.Reason != RefusalNoArchivistAvailable {
		t.Fatalf("expected no archivist reason, got %s", result.Reason)
	}
}

func TestSelectAdvertisementStalenessAndQuota(t *testing.T) {
	now := time.Now()
	ads := []Advertisement{
		{ID: "stale", PolicyAllowed: true, Available: true, TTL: time.Minute, LastRefreshed: now.Add(-time.Hour), QuotaRemaining: 100},
		{ID: "low", PolicyAllowed: true, Available: true, TTL: time.Minute * 10, LastRefreshed: now, QuotaRemaining: 5},
	}
	result := SelectAdvertisement(ads, "space-a", 50, now)
	if result.Selected {
		t.Fatalf("expected no selection due to quota")
	}
	if result.Reason != RefusalNoArchivistAvailable {
		t.Fatalf("expected no archivist reason, got %s", result.Reason)
	}
}

func TestAdvertisementRefresh(t *testing.T) {
	now := time.Now()
	ad := Advertisement{LastRefreshed: now.Add(-time.Hour), TTL: time.Minute}
	ad.Refresh(now)
	if !ad.LastRefreshed.Equal(now) {
		t.Fatalf("expected refreshed timestamp updated")
	}
	if !ad.IsFresh(now.Add(time.Second)) {
		t.Fatalf("expected advertisement fresh after refresh")
	}
}

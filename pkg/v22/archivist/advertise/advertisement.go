package advertise

import (
	"sort"
	"time"
)

type SpaceID string
type OperatorID string

type Advertisement struct {
	ID             string
	Space          SpaceID
	Operator       OperatorID
	Available      bool
	QuotaRemaining int64
	TTL            time.Duration
	LastRefreshed  time.Time
	PolicyAllowed  bool
}

func (a Advertisement) IsFresh(now time.Time) bool {
	if a.TTL <= 0 {
		return true
	}
	return now.Sub(a.LastRefreshed) <= a.TTL
}

func (a *Advertisement) Refresh(now time.Time) {
	a.LastRefreshed = now
}

type RefusalReason string

const (
	RefusalNoArchivistAvailable RefusalReason = "NO_ARCHIVIST_AVAILABLE"
	RefusalPolicyDenied         RefusalReason = "ARCHIVIST_POLICY_DENIED"
)

type SelectionResult struct {
	Advertisement Advertisement
	Reason        RefusalReason
	Selected      bool
}

func SelectAdvertisement(adverts []Advertisement, requestedSpace SpaceID, minQuota int64, now time.Time) SelectionResult {
	var candidates []Advertisement
	var hasPolicyAllowed bool

	for _, ad := range adverts {
		if !ad.PolicyAllowed {
			continue
		}
		hasPolicyAllowed = true
		if !ad.Available {
			continue
		}
		if !ad.IsFresh(now) {
			continue
		}
		if ad.QuotaRemaining < minQuota {
			continue
		}
		candidates = append(candidates, ad)
	}

	if len(candidates) == 0 {
		if len(adverts) > 0 && !hasPolicyAllowed {
			return SelectionResult{Reason: RefusalPolicyDenied}
		}
		return SelectionResult{Reason: RefusalNoArchivistAvailable}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		iSame := candidates[i].Space == requestedSpace
		jSame := candidates[j].Space == requestedSpace
		if iSame != jSame {
			return iSame
		}
		if candidates[i].QuotaRemaining != candidates[j].QuotaRemaining {
			return candidates[i].QuotaRemaining > candidates[j].QuotaRemaining
		}
		return candidates[i].LastRefreshed.After(candidates[j].LastRefreshed)
	})

	return SelectionResult{Advertisement: candidates[0], Selected: true}
}

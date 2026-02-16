package conformance

import (
	"fmt"
	"sort"
)

type ScopeBullet string

const (
	BulletDM            ScopeBullet = "dm"
	BulletPrekey        ScopeBullet = "prekey"
	BulletDMTransport   ScopeBullet = "dm-transport"
	BulletGroupDM       ScopeBullet = "group-dm"
	BulletFriends       ScopeBullet = "friends"
	BulletPresence      ScopeBullet = "presence"
	BulletCustomStatus  ScopeBullet = "custom-status"
	BulletFriendsList   ScopeBullet = "friends-list"
	BulletNotifications ScopeBullet = "notifications"
	BulletMentions      ScopeBullet = "mentions"
	BulletRBAC          ScopeBullet = "rbac"
	BulletModeration    ScopeBullet = "moderation"
	BulletSlowMode      ScopeBullet = "slow-mode"
)

var allBullets = []ScopeBullet{
	BulletDM,
	BulletPrekey,
	BulletDMTransport,
	BulletGroupDM,
	BulletFriends,
	BulletPresence,
	BulletCustomStatus,
	BulletFriendsList,
	BulletNotifications,
	BulletMentions,
	BulletRBAC,
	BulletModeration,
	BulletSlowMode,
}

// ValidateTraceMatrix enforces one-to-one scope bullet mapping.
func ValidateTraceMatrix(trace map[ScopeBullet]string) error {
	if len(trace) != len(allBullets) {
		return fmt.Errorf("trace entries mismatch: got %d want %d", len(trace), len(allBullets))
	}
	for _, bullet := range allBullets {
		if trace[bullet] == "" {
			return fmt.Errorf("missing mapping for bullet %q", bullet)
		}
	}
	return nil
}

type RequirementEvidence struct {
	Positive bool
	Negative bool
}

// ValidateRequirementMatrix ensures each requirement has positive and negative evidence.
func ValidateRequirementMatrix(items map[string]RequirementEvidence) error {
	if len(items) == 0 {
		return fmt.Errorf("requirement matrix empty")
	}
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		item := items[key]
		if !item.Positive || !item.Negative {
			return fmt.Errorf("requirement %q missing positive/negative coverage", key)
		}
	}
	return nil
}

type GateStatus struct {
	P0 bool
	P1 bool
	P2 bool
	P3 bool
	P4 bool
	P5 bool
	P6 bool
	P7 bool
}

func (s GateStatus) Complete() bool {
	return s.P0 && s.P1 && s.P2 && s.P3 && s.P4 && s.P5 && s.P6 && s.P7
}

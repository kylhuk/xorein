package policy

import "fmt"

type VersionID string

type PolicyVersion struct {
	ID        VersionID
	Major     int
	Minor     int
	Immutable bool
}

func (v PolicyVersion) String() string {
	return fmt.Sprintf("%d.%d:%s", v.Major, v.Minor, v.ID)
}

func (v PolicyVersion) CanMigrateTo(target PolicyVersion) bool {
	if !v.Immutable {
		return false
	}
	return v.Major == target.Major && target.Minor >= v.Minor
}

type PolicyTrace struct {
	Current PolicyVersion
	History []PolicyVersion
}

func (t PolicyTrace) RollbackTarget() (PolicyVersion, bool) {
	if len(t.History) == 0 {
		return PolicyVersion{}, false
	}
	return t.History[len(t.History)-1], true
}

func (t PolicyTrace) Append(version PolicyVersion) PolicyTrace {
	updated := make([]PolicyVersion, len(t.History))
	copy(updated, t.History)
	updated = append(updated, t.Current)
	return PolicyTrace{Current: version, History: updated}
}

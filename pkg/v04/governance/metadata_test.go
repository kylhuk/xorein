package governance

import (
	"reflect"
	"testing"
)

func TestRequiredGates(t *testing.T) {
	expected := []GateID{GateV4G0, GateV4G1, GateV4G2, GateV4G3, GateV4G4, GateV4G5, GateV4G6}
	if !reflect.DeepEqual(RequiredGates, expected) {
		t.Fatalf("RequiredGates order changed: got %v want %v", RequiredGates, expected)
	}
}

func TestReleaseChecklist(t *testing.T) {
	checklist := NewReleaseChecklist()
	if len(checklist.Items) != len(ReleaseChecklistItems) {
		t.Fatalf("initial items: got %d want %d", len(checklist.Items), len(ReleaseChecklistItems))
	}

	first := ReleaseChecklistItems[0]
	checklist.Mark(first)
	if !checklist.Items[first] {
		t.Fatalf("item %s should be marked true", first)
	}

	checklist.Mark("unknown item")
	if len(checklist.Items) != len(ReleaseChecklistItems) {
		t.Fatalf("unknown mark mutated list: got %d want %d", len(checklist.Items), len(ReleaseChecklistItems))
	}

	if checklist.IsComplete() {
		t.Fatal("IsComplete should be false before all items marked")
	}

	for _, item := range ReleaseChecklistItems {
		checklist.Mark(item)
	}
	if !checklist.IsComplete() {
		t.Fatal("IsComplete should be true after all marked")
	}
}

func TestOpenDecisionReminder(t *testing.T) {
	open := OpenDecisionReminder{Status: "Open"}
	if !open.IsOpen() {
		t.Fatal("IsOpen should return true for status Open")
	}

	closed := OpenDecisionReminder{Status: "Closed"}
	if closed.IsOpen() {
		t.Fatal("IsOpen should return false for non-Open status")
	}
}

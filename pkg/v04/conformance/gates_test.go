package conformance

import (
	"reflect"
	"testing"
)

func TestGateChecklistRequiredTasks(t *testing.T) {
	checklist := NewGateChecklist(GateV4G0)
	want := GateScopeMapping[GateV4G0]
	got := checklist.RequiredTasks()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RequiredTasks: got %v want %v", got, want)
	}

	got = append(got, "extra")
	if len(checklist.RequiredTasks()) != len(want) {
		t.Fatalf("RequiredTasks should return a copy: got %v want %v", checklist.RequiredTasks(), want)
	}
}

func TestGateChecklistRequiredTasksUnknown(t *testing.T) {
	checklist := GateChecklist{Gate: GateID("V4-GX")}
	if got := checklist.RequiredTasks(); got != nil {
		t.Fatalf("RequiredTasks for unknown gate: got %v want nil", got)
	}
}

func TestGateChecklistRecordEvidenceInitializes(t *testing.T) {
	checklist := GateChecklist{Gate: GateV4G0}
	checklist.RecordEvidence("P0-T1", "evidence")
	if checklist.Evidence == nil {
		t.Fatal("Evidence map should be initialized")
	}
	if checklist.Evidence["P0-T1"] != "evidence" {
		t.Fatalf("recorded evidence mismatch: got %s", checklist.Evidence["P0-T1"])
	}
}

func TestGateChecklistIsSatisfied(t *testing.T) {
	checklist := NewGateChecklist(GateV4G0)
	if checklist.IsSatisfied() {
		t.Fatal("empty checklist should not be satisfied")
	}

	checklist.Completed = true
	checklist.Evidence = nil
	if checklist.IsSatisfied() {
		t.Fatal("checklist without evidence should not be satisfied")
	}

	checklist = NewGateChecklist(GateV4G0)
	checklist.Completed = true
	for _, task := range checklist.RequiredTasks() {
		checklist.RecordEvidence(task, "ok")
	}
	if !checklist.IsSatisfied() {
		t.Fatal("checklist should be satisfied after completing all tasks")
	}
}

func TestGateChecklistIsSatisfiedRequiresScopeCoverage(t *testing.T) {
	checklist := NewGateChecklist(GateV4G1)
	checklist.Completed = true

	checklist.RecordEvidence("P1-T1", "ok")
	checklist.RecordEvidence("P1-T2", "ok")
	checklist.RecordEvidence("P1-T3", "ok")
	checklist.RecordEvidence("unexpected", "ok")

	if checklist.IsSatisfied() {
		t.Fatal("checklist with missing required task evidence should not be satisfied")
	}
}

func TestGateChecklistIsSatisfiedRequiresCompleted(t *testing.T) {
	checklist := NewGateChecklist(GateV4G2)
	for _, task := range checklist.RequiredTasks() {
		checklist.RecordEvidence(task, "ok")
	}

	if checklist.IsSatisfied() {
		t.Fatal("checklist without completed flag should not be satisfied")
	}
}

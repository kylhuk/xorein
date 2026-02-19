package apply

import "testing"

func TestApplierHonorsTombstones(t *testing.T) {
	applier := NewApplier()
	applier.ApplyTombstones([]Tombstone{{MessageID: "m-1", Reason: "removed", Timestamp: 1}})
	applier.ApplyEvents([]BackfillEvent{{MessageID: "m-1", Content: "secret"}, {MessageID: "m-2", Content: "keep"}})

	events := applier.VisibleEvents()
	if len(events) != 1 || events[0].MessageID != "m-2" {
		t.Fatalf("unexpected visible events: %+v", events)
	}

	bodies := applier.IndexableBodies()
	if len(bodies) != 1 || bodies[0] != "keep" {
		t.Fatalf("unexpected bodies: %+v", bodies)
	}
}

func TestDeletionDisclosure(t *testing.T) {
	disclosure := DeletionDisclosure("limited retention")
	if disclosure == "" {
		t.Fatalf("expected disclosure text")
	}
}

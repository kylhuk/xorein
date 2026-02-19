package apply

import "fmt"

type BackfillEvent struct {
	MessageID string
	Content   string
	Timestamp int64
}

type Tombstone struct {
	MessageID string
	Reason    string
	Timestamp int64
}

type Applier struct {
	tombstones map[string]Tombstone
	events     map[string]BackfillEvent
}

func NewApplier() *Applier {
	return &Applier{
		tombstones: make(map[string]Tombstone),
		events:     make(map[string]BackfillEvent),
	}
}

func (a *Applier) ApplyTombstones(tombs []Tombstone) {
	for _, tomb := range tombs {
		a.tombstones[tomb.MessageID] = tomb
		delete(a.events, tomb.MessageID)
	}
}

func (a *Applier) ApplyEvents(events []BackfillEvent) {
	for _, event := range events {
		if _, tombstoned := a.tombstones[event.MessageID]; tombstoned {
			continue
		}
		a.events[event.MessageID] = event
	}
}

func (a *Applier) VisibleEvents() []BackfillEvent {
	visible := make([]BackfillEvent, 0, len(a.events))
	for _, ev := range a.events {
		if _, tombstoned := a.tombstones[ev.MessageID]; tombstoned {
			continue
		}
		visible = append(visible, ev)
	}
	return visible
}

func (a *Applier) IndexableBodies() []string {
	bodies := make([]string, 0, len(a.events))
	for _, ev := range a.VisibleEvents() {
		bodies = append(bodies, ev.Content)
	}
	return bodies
}

func DeletionDisclosure(limit string) string {
	if limit == "" {
		limit = "controlled by policy"
	}
	return fmt.Sprintf("Deletion cannot be fully guaranteed (%s).", limit)
}

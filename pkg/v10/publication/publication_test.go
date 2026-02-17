package publication

import "testing"

func TestSectionCollectionNonEmpty(t *testing.T) {
	t.Parallel()

	if len(SectionMap()) == 0 {
		t.Fatal("expected section map to be populated")
	}
	if got := len(SectionList()); got != 3 {
		t.Fatalf("section list length = %d, want 3", got)
	}
}

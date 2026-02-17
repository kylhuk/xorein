package docs

import "testing"

func TestGuideChecklistsHaveExpectedCounts(t *testing.T) {
	t.Parallel()

	t.Run("user", func(t *testing.T) {
		if got, want := len(UserGuideChecklist()), 3; got != want {
			t.Fatalf("user guide checklist length = %d, want %d", got, want)
		}
	})
	t.Run("admin", func(t *testing.T) {
		if got, want := len(AdminGuideChecklist()), 3; got != want {
			t.Fatalf("admin guide checklist length = %d, want %d", got, want)
		}
	})
	t.Run("developer", func(t *testing.T) {
		if got, want := len(DeveloperGuideChecklist()), 4; got != want {
			t.Fatalf("developer guide checklist length = %d, want %d", got, want)
		}
	})
}

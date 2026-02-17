package scenario

import (
	"fmt"
	"strings"
	"time"

	"github.com/aether/code_aether/pkg/v08/accessibility"
	"github.com/aether/code_aether/pkg/v08/bookmarks"
	"github.com/aether/code_aether/pkg/v08/conformance"
	"github.com/aether/code_aether/pkg/v08/i18n"
	"github.com/aether/code_aether/pkg/v08/linkpreview"
	"github.com/aether/code_aether/pkg/v08/pinning"
	"github.com/aether/code_aether/pkg/v08/themes"
	"github.com/aether/code_aether/pkg/v08/threads"
	"github.com/aether/code_aether/pkg/v08/voice"
)

// RunEchoContracts executes a deterministic subset of v0.8 helpers.
func RunEchoContracts() error {
	gates := conformance.Gates()
	if errs := conformance.ValidateChecklist(gates); len(errs) > 0 {
		return fmt.Errorf("gate checklist issues: %s", strings.Join(errs, ", "))
	}

	trace := threads.ThreadTrace{ID: "thread-echo", CreatedDepth: 1, ReplyDepth: 1, Lifecycle: threads.ClassifyLifecycle(1)}
	if err := threads.ValidateReplyLineage(trace); err != nil {
		return fmt.Errorf("thread trace validation failed: %w", err)
	}

	authorities := []pinning.PinAuthority{
		{ID: "pin-a", Scope: pinning.ScopePersonal, Priority: 10},
		{ID: "pin-b", Scope: pinning.ScopeTeam, Priority: 5},
		{ID: "pin-c", Scope: pinning.ScopeGlobal, Priority: 1},
	}
	ordered := pinning.DeterministicOrder(authorities)
	if len(ordered) != len(authorities) {
		return fmt.Errorf("pin ordering lost entries")
	}

	bookmark := bookmarks.Bookmark{ID: "bm-echo", Owner: "tester", Privacy: bookmarks.PrivacyPersonal, CreatedAt: time.Now()}
	if err := bookmarks.ValidatePrivacy(bookmark); err != nil {
		return err
	}
	_ = bookmarks.Lifecycle(bookmark)

	normalized, err := linkpreview.NormalizeURL("https://example.com/v8")
	if err != nil {
		return err
	}
	if !linkpreview.PreviewEligibility(normalized) {
		return fmt.Errorf("link preview ineligible %s", normalized)
	}
	meta := linkpreview.Metadata{
		OG:      map[string]string{"title": "OG Title"},
		Twitter: map[string]string{"title": "Twitter Title", "description": "desc"},
	}
	_ = linkpreview.MetadataPrecedence(meta)
	_ = linkpreview.CacheKey(linkpreview.RenderState{URL: normalized, Cached: true, CacheState: "fresh"})

	if _, err := themes.ValidateCustomTheme([]byte(`{"name":"echo","tokens":{"background":"#112233","text":"#ffffff"}}`), "night"); err != nil {
		return err
	}

	announcer := accessibility.NewAnnouncer(1 * time.Second)
	now := time.Now()
	if !announcer.ShouldAnnounce("announce-echo", now) {
		return fmt.Errorf("announcement should initially pass")
	}
	if announcer.ShouldAnnounce("announce-echo", now.Add(500*time.Millisecond)) {
		return fmt.Errorf("announcement throttle failed")
	}
	_ = accessibility.HighContrastToken(true, "#ffffff")
	graph := accessibility.FocusGraph{}
	graph.AddEdge("start", "next")
	_ = graph.Neighbors("start")

	if msg := i18n.MissingKeyMessage("es-ES", "welcome"); msg == "" {
		return fmt.Errorf("missing key handler empty")
	}
	_ = i18n.FormatLocalizedNumber("fr-FR", 42)

	decision := voice.SelectNoiseReducer(false, fmt.Errorf("detected instability"))
	if decision.Selected != voice.DTLN {
		return fmt.Errorf("voice selection unexpected: %s", decision.Selected)
	}

	return nil
}

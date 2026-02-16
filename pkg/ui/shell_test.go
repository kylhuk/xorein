package ui

import (
	"errors"
	"strings"
	"testing"
)

func newShellWithIdentity(t *testing.T) *Shell {
	t.Helper()
	shell := NewShell()
	if err := shell.SetIdentity("tester"); err != nil {
		t.Fatalf("set identity: %v", err)
	}
	return shell
}

func TestShellEntryPointsAndDefaults(t *testing.T) {
	entries := EntryPoints()
	if len(entries) != 4 {
		t.Fatalf("expected 4 shell entry points, got %d", len(entries))
	}
	want := map[Route]bool{
		RouteIdentitySetup: true,
		RouteServerList:    true,
		RouteChannelView:   true,
		RouteSettings:      true,
	}
	for _, entry := range entries {
		if !want[entry] {
			t.Fatalf("unexpected entry point %q", entry)
		}
		delete(want, entry)
	}
	if len(want) != 0 {
		t.Fatalf("missing expected entry points: %v", want)
	}

	shell := NewShell()
	state := shell.State()
	if state.CurrentRoute != RouteIdentitySetup {
		t.Fatalf("expected default route %q, got %q", RouteIdentitySetup, state.CurrentRoute)
	}
	if len(state.Servers) != 0 {
		t.Fatalf("expected empty server list by default")
	}
}

func TestShellRouteTransitionsPreserveState(t *testing.T) {
	shell := NewShell()

	if err := shell.Navigate(RouteServerList); !errors.Is(err, ErrIdentitySetupRequired) {
		t.Fatalf("expected identity requirement before server list, got %v", err)
	}
	if err := shell.SetIdentity("alice"); err != nil {
		t.Fatalf("set identity: %v", err)
	}
	if err := shell.UpsertServer("srv-1", "Main server"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.Navigate(RouteServerList); err != nil {
		t.Fatalf("navigate to server list: %v", err)
	}
	if got := shell.State().CurrentRoute; got != RouteServerList {
		t.Fatalf("expected server list route, got %q", got)
	}

	if err := shell.Navigate(RouteChannelView); !errors.Is(err, ErrServerSelectionRequired) {
		t.Fatalf("expected server selection requirement, got %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.Navigate(RouteChannelView); !errors.Is(err, ErrChannelSelectionMissing) {
		t.Fatalf("expected channel requirement, got %v", err)
	}
	if err := shell.SelectChannel("general"); err != nil {
		t.Fatalf("select channel: %v", err)
	}
	if err := shell.Navigate(RouteChannelView); err != nil {
		t.Fatalf("navigate to channel view: %v", err)
	}
	state := shell.State()
	if state.CurrentRoute != RouteChannelView {
		t.Fatalf("expected channel view route, got %q", state.CurrentRoute)
	}
	if state.SelectedServerID != "srv-1" || state.SelectedChannelID != "general" {
		t.Fatalf("state not preserved, got server=%q channel=%q", state.SelectedServerID, state.SelectedChannelID)
	}
	if state.IdentityDisplay != "alice" {
		t.Fatalf("identity not preserved across route transitions")
	}
}

func TestShellGuardsAndStateErrors(t *testing.T) {
	var nilShell *Shell
	if err := nilShell.SetIdentity("alice"); !errors.Is(err, ErrShellRequired) {
		t.Fatalf("expected ErrShellRequired from nil shell, got %v", err)
	}
	if err := nilShell.Navigate(RouteIdentitySetup); !errors.Is(err, ErrShellRequired) {
		t.Fatalf("expected ErrShellRequired from nil shell navigate, got %v", err)
	}

	shell := NewShell()
	if err := shell.SetIdentity(" "); !errors.Is(err, ErrIdentitySetupRequired) {
		t.Fatalf("expected ErrIdentitySetupRequired, got %v", err)
	}
	if err := shell.SetIdentity("bob"); err != nil {
		t.Fatalf("set identity: %v", err)
	}
	if err := shell.UpsertServer("", ""); !errors.Is(err, ErrServerSelectionRequired) {
		t.Fatalf("expected ErrServerSelectionRequired for blank server, got %v", err)
	}
	if err := shell.UpsertServer("srv-1", "One"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.SelectServer("srv-unknown"); !errors.Is(err, ErrServerSelectionRequired) {
		t.Fatalf("expected ErrServerSelectionRequired for unknown server, got %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select known server: %v", err)
	}
	if err := shell.SelectChannel(" "); !errors.Is(err, ErrChannelSelectionMissing) {
		t.Fatalf("expected ErrChannelSelectionMissing, got %v", err)
	}
	if err := shell.Navigate(Route("invalid")); !errors.Is(err, ErrRouteInvalid) {
		t.Fatalf("expected ErrRouteInvalid, got %v", err)
	}
}

func TestServerRailReflectsJoinedServers(t *testing.T) {
	shell := newShellWithIdentity(t)
	servers := []struct {
		id   string
		name string
	}{
		{"srv-1", "Alpha"},
		{"srv-2", "Beta"},
		{"srv-3", "Gamma"},
	}
	for _, srv := range servers {
		if err := shell.UpsertServer(srv.id, srv.name); err != nil {
			t.Fatalf("upsert server %s: %v", srv.id, err)
		}
	}
	if err := shell.SelectServer("srv-2"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	rail := shell.ServerRail()
	if len(rail) != len(servers) {
		t.Fatalf("expected %d rail entries, got %d", len(servers), len(rail))
	}
	for i, item := range rail {
		if item.ID != servers[i].id || item.Name != servers[i].name {
			t.Fatalf("rail[%d] mismatch: got %q/%q want %q/%q", i, item.ID, item.Name, servers[i].id, servers[i].name)
		}
		expectedSelected := item.ID == "srv-2"
		if item.Selected != expectedSelected {
			t.Fatalf("rail[%d] selected mismatch for %s: got %v want %v", i, item.ID, item.Selected, expectedSelected)
		}
	}
}

func TestChannelSidebarReflectsSelectedServer(t *testing.T) {
	shell := newShellWithIdentity(t)
	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert srv-1: %v", err)
	}
	if err := shell.UpsertServer("srv-2", "Beta"); err != nil {
		t.Fatalf("upsert srv-2: %v", err)
	}
	channels := map[string][]ChannelSummary{
		"srv-1": {
			{ServerID: "srv-1", ID: "general", Name: "General"},
			{ServerID: "srv-1", ID: "random", Name: "Random"},
		},
		"srv-2": {
			{ServerID: "srv-2", ID: "ops", Name: "Ops"},
		},
	}
	for serverID, chs := range channels {
		for _, ch := range chs {
			if err := shell.UpsertChannel(serverID, ch.ID, ch.Name); err != nil {
				t.Fatalf("upsert channel %s/%s: %v", serverID, ch.ID, err)
			}
		}
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select srv-1: %v", err)
	}
	if err := shell.SelectChannel("random"); err != nil {
		t.Fatalf("select channel random: %v", err)
	}
	sidebar := shell.ChannelSidebar()
	if len(sidebar) != len(channels["srv-1"]) {
		t.Fatalf("expected %d channels for srv-1, got %d", len(channels["srv-1"]), len(sidebar))
	}
	for i, item := range sidebar {
		expected := channels["srv-1"][i]
		if item.ID != expected.ID {
			t.Fatalf("srv-1 sidebar[%d] mismatch: got %s want %s", i, item.ID, expected.ID)
		}
		wantSelected := expected.ID == "random"
		if item.Selected != wantSelected {
			t.Fatalf("srv-1 sidebar[%d] selected mismatch for %s: got %v want %v", i, item.ID, item.Selected, wantSelected)
		}
	}
	if err := shell.SelectServer("srv-2"); err != nil {
		t.Fatalf("select srv-2: %v", err)
	}
	sidebar = shell.ChannelSidebar()
	if len(sidebar) != len(channels["srv-2"]) {
		t.Fatalf("expected %d channels for srv-2, got %d", len(channels["srv-2"]), len(sidebar))
	}
	if len(sidebar) != 1 || sidebar[0].ID != "ops" {
		t.Fatalf("srv-2 sidebar unexpected payload: %+v", sidebar)
	}
	if sidebar[0].Selected {
		t.Fatalf("expected no selected channel for srv-2 after server switch")
	}
}

func TestSelectionSynchronization(t *testing.T) {
	shell := newShellWithIdentity(t)
	for _, srv := range []string{"srv-1", "srv-2"} {
		if err := shell.UpsertServer(srv, strings.ToUpper(srv)); err != nil {
			t.Fatalf("upsert server %s: %v", srv, err)
		}
	}
	if err := shell.UpsertChannel("srv-1", "general", "General"); err != nil {
		t.Fatalf("upsert srv-1/general: %v", err)
	}
	if err := shell.UpsertChannel("srv-2", "ops", "Ops"); err != nil {
		t.Fatalf("upsert srv-2/ops: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select srv-1: %v", err)
	}
	if err := shell.SelectChannel("general"); err != nil {
		t.Fatalf("select general: %v", err)
	}
	state := shell.State()
	if state.Subscription.ServerID != "srv-1" || state.Subscription.ChannelID != "general" {
		t.Fatalf("subscription mismatch after selecting channel: %+v", state.Subscription)
	}
	if state.CurrentRoute != RouteChannelView {
		t.Fatalf("expected channel view route after selecting channel, got %s", state.CurrentRoute)
	}
	if err := shell.SelectServer("srv-2"); err != nil {
		t.Fatalf("select srv-2: %v", err)
	}
	state = shell.State()
	if state.SelectedChannelID != "" {
		t.Fatalf("expected channel cleared after switching server, got %q", state.SelectedChannelID)
	}
	if state.CurrentRoute != RouteServerList {
		t.Fatalf("expected server list route after switching server, got %s", state.CurrentRoute)
	}
	if state.Subscription.ServerID != "srv-2" || state.Subscription.ChannelID != "" {
		t.Fatalf("subscription mismatch after switching server: %+v", state.Subscription)
	}
	if err := shell.SelectChannel("ops"); err != nil {
		t.Fatalf("select srv-2/ops: %v", err)
	}
	state = shell.State()
	if state.Subscription.ServerID != "srv-2" || state.Subscription.ChannelID != "ops" {
		t.Fatalf("subscription mismatch after selecting srv-2/ops: %+v", state.Subscription)
	}
	if state.CurrentRoute != RouteChannelView {
		t.Fatalf("expected channel view route after selecting srv-2/ops, got %s", state.CurrentRoute)
	}
}

func TestChannelSelectionFailsAfterInventoryExists(t *testing.T) {
	shell := newShellWithIdentity(t)
	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.SelectChannel("ad-hoc"); err != nil {
		t.Fatalf("select ad-hoc before inventory: %v", err)
	}
	if err := shell.UpsertChannel("srv-1", "general", "General"); err != nil {
		t.Fatalf("upsert general: %v", err)
	}
	if err := shell.UpsertChannel("srv-1", "random", "Random"); err != nil {
		t.Fatalf("upsert random: %v", err)
	}
	if err := shell.SelectChannel("missing"); !errors.Is(err, ErrChannelSelectionMissing) {
		t.Fatalf("expected ErrChannelSelectionMissing selecting unknown channel, got %v", err)
	}
}

func TestComposeVirtualMessageWindow(t *testing.T) {
	inventory := []string{"m0", "m1", "m2", "m3", "m4"}

	t.Run("normal window", func(t *testing.T) {
		got := ComposeVirtualMessageWindow(inventory, 1, 3)
		if got.Start != 1 || got.End != 4 || got.Total != 5 {
			t.Fatalf("unexpected bounds: %+v", got)
		}
		if !got.HasMoreAbove || !got.HasMoreBelow {
			t.Fatalf("expected both directions available: %+v", got)
		}
		if len(got.Items) != 3 || got.Items[0] != "m1" || got.Items[2] != "m3" {
			t.Fatalf("unexpected items: %+v", got.Items)
		}
	})

	t.Run("empty inventory", func(t *testing.T) {
		got := ComposeVirtualMessageWindow[string](nil, 0, 3)
		if got.Total != 0 || got.Start != 0 || got.End != 0 {
			t.Fatalf("unexpected empty window metadata: %+v", got)
		}
		if got.HasMoreAbove || got.HasMoreBelow || len(got.Items) != 0 {
			t.Fatalf("unexpected empty window flags/items: %+v", got)
		}
	})

	t.Run("anchor clamped low", func(t *testing.T) {
		got := ComposeVirtualMessageWindow(inventory, -50, 2)
		if got.Start != 0 || got.End != 2 {
			t.Fatalf("unexpected clamped low bounds: %+v", got)
		}
		if got.HasMoreAbove {
			t.Fatalf("expected no items above at start: %+v", got)
		}
	})

	t.Run("anchor clamped high", func(t *testing.T) {
		got := ComposeVirtualMessageWindow(inventory, 99, 2)
		if got.Start != 3 || got.End != 5 {
			t.Fatalf("unexpected clamped high bounds: %+v", got)
		}
		if got.HasMoreBelow {
			t.Fatalf("expected no items below at end: %+v", got)
		}
	})

	t.Run("tiny window size clamps to one", func(t *testing.T) {
		got := ComposeVirtualMessageWindow(inventory, 2, 0)
		if got.Start != 2 || got.End != 3 || len(got.Items) != 1 || got.Items[0] != "m2" {
			t.Fatalf("unexpected tiny window result: %+v", got)
		}
	})

	t.Run("large window size clamps to total", func(t *testing.T) {
		got := ComposeVirtualMessageWindow(inventory, 4, 999)
		if got.Start != 0 || got.End != 5 || len(got.Items) != 5 {
			t.Fatalf("unexpected large window result: %+v", got)
		}
		if got.HasMoreAbove || got.HasMoreBelow {
			t.Fatalf("expected no hidden items when full inventory rendered: %+v", got)
		}
	})
}

func TestHandleComposerEnter(t *testing.T) {
	got := HandleComposerEnter("hello", false)
	if !got.ShouldSend || got.InsertedNew || got.Draft != "hello" {
		t.Fatalf("unexpected enter-send behavior: %+v", got)
	}

	got = HandleComposerEnter("hello", true)
	if got.ShouldSend || !got.InsertedNew || got.Draft != "hello\n" {
		t.Fatalf("unexpected shift-enter behavior: %+v", got)
	}
}

func TestValidateComposerMessage(t *testing.T) {
	normalized, err := ValidateComposerMessage("  hello world  ", 50)
	if err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}
	if normalized != "hello world" {
		t.Fatalf("unexpected normalized message %q", normalized)
	}

	if _, err := ValidateComposerMessage("   \n\t", 50); !errors.Is(err, ErrComposerMessageEmpty) {
		t.Fatalf("expected ErrComposerMessageEmpty, got %v", err)
	}

	if _, err := ValidateComposerMessage("abcdef", 5); !errors.Is(err, ErrComposerMessageTooLong) {
		t.Fatalf("expected ErrComposerMessageTooLong, got %v", err)
	}

	if _, err := ValidateComposerMessage(strings.Repeat("x", DefaultComposerMaxLength+1), 0); !errors.Is(err, ErrComposerMessageTooLong) {
		t.Fatalf("expected default max length enforcement, got %v", err)
	}
}

func TestRenderMessageMarkdownSubset(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantPlain       string
		wantTypes       []MarkdownTokenType
		wantUnsupported bool
		wantErr         error
	}{
		{
			name:      "supports basic inline markdown tokens",
			input:     "hello **bold** *italic* `code`",
			wantPlain: "hello bold italic code",
			wantTypes: []MarkdownTokenType{MarkdownTokenText, MarkdownTokenBold, MarkdownTokenText, MarkdownTokenItalic, MarkdownTokenText, MarkdownTokenCode},
		},
		{
			name:            "unsupported headings degrade gracefully",
			input:           "# heading",
			wantPlain:       "# heading",
			wantTypes:       []MarkdownTokenType{MarkdownTokenText},
			wantUnsupported: true,
		},
		{
			name:      "multiline inserts line break token",
			input:     "line1\nline2",
			wantPlain: "line1\nline2",
			wantTypes: []MarkdownTokenType{MarkdownTokenText, MarkdownTokenLineBreak, MarkdownTokenText},
		},
		{
			name:    "empty normalized input rejected",
			input:   "  \n\t  ",
			wantErr: ErrMessageRenderEmpty,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RenderMessageMarkdownSubset(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.PlainText != tc.wantPlain {
				t.Fatalf("plain text mismatch: got %q want %q", got.PlainText, tc.wantPlain)
			}
			if got.UnsupportedMarkdown != tc.wantUnsupported {
				t.Fatalf("unsupported flag mismatch: got %v want %v", got.UnsupportedMarkdown, tc.wantUnsupported)
			}
			if len(got.Tokens) != len(tc.wantTypes) {
				t.Fatalf("token count mismatch: got %d want %d (%+v)", len(got.Tokens), len(tc.wantTypes), got.Tokens)
			}
			for i := range tc.wantTypes {
				if got.Tokens[i].Type != tc.wantTypes[i] {
					t.Fatalf("token[%d] type mismatch: got %q want %q", i, got.Tokens[i].Type, tc.wantTypes[i])
				}
			}
		})
	}
}

func TestRenderMessageMarkdownSubsetEscapesHTML(t *testing.T) {
	got, err := RenderMessageMarkdownSubset("<b>unsafe</b>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PlainText != "&lt;b&gt;unsafe&lt;/b&gt;" {
		t.Fatalf("expected escaped plain text, got %q", got.PlainText)
	}
	if len(got.Tokens) != 1 || got.Tokens[0].Value != "&lt;b&gt;unsafe&lt;/b&gt;" {
		t.Fatalf("expected escaped single text token, got %+v", got.Tokens)
	}
}

func TestBuildReplyReference(t *testing.T) {
	tests := []struct {
		name            string
		messageID       string
		author          string
		excerpt         string
		maxRunes        int
		wantErr         error
		wantAuthor      string
		wantExcerpt     string
		wantTruncated   bool
		wantSummaryPart string
	}{
		{
			name:            "builds sanitized summary",
			messageID:       "msg-1",
			author:          "Alice <admin>",
			excerpt:         "hello <world>",
			maxRunes:        40,
			wantAuthor:      "Alice &lt;admin&gt;",
			wantExcerpt:     "hello &lt;world&gt;",
			wantSummaryPart: "↪ Alice &lt;admin&gt;: hello &lt;world&gt;",
		},
		{
			name:            "truncates excerpt by rune count",
			messageID:       "msg-2",
			author:          "Bob",
			excerpt:         "abcdef",
			maxRunes:        3,
			wantAuthor:      "Bob",
			wantExcerpt:     "abc",
			wantTruncated:   true,
			wantSummaryPart: "↪ Bob: abc…",
		},
		{
			name:      "missing message id rejected",
			messageID: "   ",
			wantErr:   ErrReplyReferenceIDMissing,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildReplyReference(tc.messageID, tc.author, tc.excerpt, tc.maxRunes)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.MessageID != strings.TrimSpace(tc.messageID) {
				t.Fatalf("message id mismatch: got %q", got.MessageID)
			}
			if got.AuthorDisplay != tc.wantAuthor {
				t.Fatalf("author mismatch: got %q want %q", got.AuthorDisplay, tc.wantAuthor)
			}
			if got.Excerpt != tc.wantExcerpt {
				t.Fatalf("excerpt mismatch: got %q want %q", got.Excerpt, tc.wantExcerpt)
			}
			if got.Truncated != tc.wantTruncated {
				t.Fatalf("truncated mismatch: got %v want %v", got.Truncated, tc.wantTruncated)
			}
			if got.DisplaySummary != tc.wantSummaryPart {
				t.Fatalf("summary mismatch: got %q want %q", got.DisplaySummary, tc.wantSummaryPart)
			}
		})
	}
}

func TestDraftPersistenceAcrossServerChannelSwitches(t *testing.T) {
	shell := newShellWithIdentity(t)
	for _, server := range []string{"srv-1", "srv-2"} {
		if err := shell.UpsertServer(server, server); err != nil {
			t.Fatalf("upsert server %s: %v", server, err)
		}
	}
	if err := shell.UpsertChannel("srv-1", "general", "General"); err != nil {
		t.Fatalf("upsert srv-1/general: %v", err)
	}
	if err := shell.UpsertChannel("srv-2", "ops", "Ops"); err != nil {
		t.Fatalf("upsert srv-2/ops: %v", err)
	}

	if err := shell.SaveDraft("srv-1", "general", "draft one"); err != nil {
		t.Fatalf("save srv-1/general draft: %v", err)
	}
	if err := shell.SaveDraft("srv-2", "ops", "draft two"); err != nil {
		t.Fatalf("save srv-2/ops draft: %v", err)
	}

	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select srv-1: %v", err)
	}
	if err := shell.SelectChannel("general"); err != nil {
		t.Fatalf("select srv-1/general: %v", err)
	}
	if err := shell.SelectServer("srv-2"); err != nil {
		t.Fatalf("select srv-2: %v", err)
	}
	if err := shell.SelectChannel("ops"); err != nil {
		t.Fatalf("select srv-2/ops: %v", err)
	}

	draft, err := shell.LoadDraft("srv-1", "general")
	if err != nil {
		t.Fatalf("load srv-1/general draft: %v", err)
	}
	if draft != "draft one" {
		t.Fatalf("unexpected srv-1/general draft %q", draft)
	}

	draft, err = shell.LoadDraft("srv-2", "ops")
	if err != nil {
		t.Fatalf("load srv-2/ops draft: %v", err)
	}
	if draft != "draft two" {
		t.Fatalf("unexpected srv-2/ops draft %q", draft)
	}

	err = shell.ClearDraft("srv-1", "general")
	if err != nil {
		t.Fatalf("clear srv-1/general draft: %v", err)
	}
	draft, err = shell.LoadDraft("srv-1", "general")
	if err != nil {
		t.Fatalf("load cleared draft: %v", err)
	}
	if draft != "" {
		t.Fatalf("expected cleared draft to be empty, got %q", draft)
	}
}

func TestDraftScopeKeyValidation(t *testing.T) {
	if _, err := DraftScopeKey("", "general"); !errors.Is(err, ErrServerSelectionRequired) {
		t.Fatalf("expected ErrServerSelectionRequired, got %v", err)
	}
	if _, err := DraftScopeKey("srv-1", ""); !errors.Is(err, ErrChannelSelectionMissing) {
		t.Fatalf("expected ErrChannelSelectionMissing, got %v", err)
	}
	key, err := DraftScopeKey(" srv-1 ", " general ")
	if err != nil {
		t.Fatalf("unexpected key generation error: %v", err)
	}
	if key != "srv-1/general" {
		t.Fatalf("unexpected key %q", key)
	}
}

func TestSettingsPersistenceFlow(t *testing.T) {
	shell := newShellWithIdentity(t)

	profile := SettingsProfileState{
		DisplayName:   "  Alice  ",
		StatusMessage: "  Ready  ",
	}
	audio := SettingsAudioState{
		InputMuted:       true,
		OutputMuted:      false,
		InputVolume:      -20,
		OutputVolume:     180,
		NoiseSuppression: false,
		EchoCancellation: false,
	}

	if err := shell.AdjustProfileSettings(profile); err != nil {
		t.Fatalf("adjust profile: %v", err)
	}
	if err := shell.AdjustAudioSettings(audio); err != nil {
		t.Fatalf("adjust audio: %v", err)
	}

	view := shell.SettingsView()
	if view.Profile.DisplayName != "Alice" || view.Profile.StatusMessage != "Ready" {
		t.Fatalf("unexpected profile view: %+v", view.Profile)
	}
	if view.Audio.InputVolume != 0 || view.Audio.OutputVolume != 100 {
		t.Fatalf("expected clamped audio volumes, got input=%d output=%d", view.Audio.InputVolume, view.Audio.OutputVolume)
	}
	if view.SavedVersion != 0 || !view.Dirty {
		t.Fatalf("expected dirty unsaved settings with version 0, got version=%d dirty=%v", view.SavedVersion, view.Dirty)
	}

	if err := shell.SaveSettings(); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	view = shell.SettingsView()
	if view.SavedVersion != 1 || view.Dirty {
		t.Fatalf("expected version increment and clean state, got version=%d dirty=%v", view.SavedVersion, view.Dirty)
	}

	if err := shell.AdjustAudioSettings(SettingsAudioState{InputVolume: 101, OutputVolume: 20}); err != nil {
		t.Fatalf("adjust audio again: %v", err)
	}
	view = shell.SettingsView()
	if !view.Dirty {
		t.Fatalf("expected dirty settings after adjustment, got clean")
	}
	if view.Audio.InputVolume != 100 || view.Audio.OutputVolume != 20 {
		t.Fatalf("unexpected updated audio view: %+v", view.Audio)
	}

	if err := shell.ReloadSettings(); err != nil {
		t.Fatalf("reload settings: %v", err)
	}
	view = shell.SettingsView()
	if view.Dirty {
		t.Fatalf("expected clean settings after reload")
	}
	if view.Profile.DisplayName != "Alice" || view.Profile.StatusMessage != "Ready" {
		t.Fatalf("expected persisted profile restored after reload, got %+v", view.Profile)
	}
	if view.Audio.InputVolume != 0 || view.Audio.OutputVolume != 100 {
		t.Fatalf("expected persisted audio restored after reload, got %+v", view.Audio)
	}
}

func TestNetworkDiagnosticsValidationAndView(t *testing.T) {
	shell := newShellWithIdentity(t)

	valid := NetworkDiagnosticsState{
		DirectPathStatus: NetworkPathStatusActive,
		RelayPathStatus:  NetworkPathStatusConnecting,
		ActivePath:       NetworkPathDirect,
		ReasonClass:      NetworkReasonClassNATTraversal,
		Summary:          "  ok ",
	}
	if err := shell.SetNetworkDiagnostics(valid); err != nil {
		t.Fatalf("set network diagnostics: %v", err)
	}
	view := shell.NetworkDiagnosticsView()
	if view.DirectPathStatus != NetworkPathStatusActive || view.RelayPathStatus != NetworkPathStatusConnecting {
		t.Fatalf("unexpected path statuses: %+v", view)
	}
	if view.ActivePath != NetworkPathDirect || view.ReasonClass != NetworkReasonClassNATTraversal {
		t.Fatalf("unexpected active path or reason: %+v", view)
	}
	if view.Summary != "ok" {
		t.Fatalf("expected trimmed summary, got %q", view.Summary)
	}

	cases := []struct {
		name  string
		state NetworkDiagnosticsState
		want  error
	}{
		{
			name: "invalid direct status",
			state: NetworkDiagnosticsState{
				DirectPathStatus: NetworkPathStatus("bogus"),
				RelayPathStatus:  NetworkPathStatusUnavailable,
				ActivePath:       NetworkPathNone,
				ReasonClass:      NetworkReasonClassNone,
			},
			want: ErrNetworkPathStatusInvalid,
		},
		{
			name: "invalid relay status",
			state: NetworkDiagnosticsState{
				DirectPathStatus: NetworkPathStatusUnavailable,
				RelayPathStatus:  NetworkPathStatus("bogus"),
				ActivePath:       NetworkPathNone,
				ReasonClass:      NetworkReasonClassNone,
			},
			want: ErrNetworkPathStatusInvalid,
		},
		{
			name: "invalid active path",
			state: NetworkDiagnosticsState{
				DirectPathStatus: NetworkPathStatusUnavailable,
				RelayPathStatus:  NetworkPathStatusUnavailable,
				ActivePath:       NetworkPath("warp"),
				ReasonClass:      NetworkReasonClassNone,
			},
			want: ErrNetworkPathInvalid,
		},
		{
			name: "invalid reason class",
			state: NetworkDiagnosticsState{
				DirectPathStatus: NetworkPathStatusUnavailable,
				RelayPathStatus:  NetworkPathStatusUnavailable,
				ActivePath:       NetworkPathNone,
				ReasonClass:      NetworkReasonClass("mystery"),
			},
			want: ErrNetworkReasonClassInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := shell.SetNetworkDiagnostics(tc.state); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestDiagnosticsRetentionExportAndRedaction(t *testing.T) {
	shell := newShellWithIdentity(t)
	if err := shell.SetDiagnosticRetentionLimit(2); err != nil {
		t.Fatalf("set retention limit: %v", err)
	}

	records := []DiagnosticRecord{
		{
			Category:     " auth ",
			Message:      "token=abc",
			Metadata:     map[string]string{"token": "abc", "note": "authorization=bad", "safe": "ok"},
			OccurredUnix: 10,
		},
		{
			Category:     "net",
			Message:      "ok",
			Metadata:     map[string]string{"password": "secret", "note": "secret=abc"},
			OccurredUnix: 20,
		},
		{
			Category:     "net",
			Message:      "session_key=xyz",
			Metadata:     map[string]string{"meta": "token=xyz"},
			OccurredUnix: 30,
		},
	}
	for i, record := range records {
		if err := shell.AddDiagnosticRecord(record); err != nil {
			t.Fatalf("add diagnostic record %d: %v", i, err)
		}
	}

	if _, err := shell.ExportDiagnostics(false, "auto"); !errors.Is(err, ErrDiagnosticExportUserTriggerRequired) {
		t.Fatalf("expected ErrDiagnosticExportUserTriggerRequired, got %v", err)
	}

	envelope, err := shell.ExportDiagnostics(true, "  user request  ")
	if err != nil {
		t.Fatalf("export diagnostics: %v", err)
	}
	if envelope.Version != "v1" || !envelope.TriggeredByUser {
		t.Fatalf("unexpected envelope header: %+v", envelope)
	}
	if envelope.Reason != "user request" {
		t.Fatalf("expected trimmed reason, got %q", envelope.Reason)
	}
	if envelope.RetentionLimit != 2 || envelope.RecordCount != 2 || envelope.GeneratedUnix != 30 {
		t.Fatalf("unexpected envelope counts: limit=%d count=%d generated=%d", envelope.RetentionLimit, envelope.RecordCount, envelope.GeneratedUnix)
	}
	if envelope.Redaction.Token != "[REDACTED]" || envelope.Redaction.Mode == "" {
		t.Fatalf("unexpected redaction policy: %+v", envelope.Redaction)
	}
	if !containsString(envelope.Redaction.RedactedKeys, "token") || !containsString(envelope.Redaction.RedactedKeys, "password") {
		t.Fatalf("expected redaction keys to include token/password, got %+v", envelope.Redaction.RedactedKeys)
	}
	if len(envelope.Records) != 2 {
		t.Fatalf("expected 2 retained records, got %d", len(envelope.Records))
	}

	first := envelope.Records[0]
	if first.Category != "net" || first.Message != "ok" {
		t.Fatalf("unexpected first retained record: %+v", first)
	}
	if first.Metadata["password"] != "[REDACTED]" {
		t.Fatalf("expected password metadata redacted, got %q", first.Metadata["password"])
	}
	if first.Metadata["note"] != "secret=[REDACTED]" {
		t.Fatalf("expected marker redaction in metadata, got %q", first.Metadata["note"])
	}

	second := envelope.Records[1]
	if second.Message != "session_key=[REDACTED]" {
		t.Fatalf("expected message redaction, got %q", second.Message)
	}
	if second.Metadata["meta"] != "token=[REDACTED]" {
		t.Fatalf("expected metadata marker redaction, got %q", second.Metadata["meta"])
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestVoiceBarVisibilityAcrossNavigation(t *testing.T) {
	shell := newShellWithIdentity(t)
	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert voice channel: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.SelectChannel("voice"); err != nil {
		t.Fatalf("select channel: %v", err)
	}
	if err := shell.StartVoiceSessionForScope("srv-1", "voice"); err != nil {
		t.Fatalf("start scoped voice session: %v", err)
	}
	if err := shell.Navigate(RouteSettings); err != nil {
		t.Fatalf("navigate settings: %v", err)
	}
	bar := shell.PersistentVoiceBar()
	if !bar.Visible {
		t.Fatalf("expected voice bar visible after navigation, got hidden reason=%s", bar.HiddenReason)
	}
	if bar.ServerID != "srv-1" || bar.ChannelID != "voice" {
		t.Fatalf("unexpected voice bar scope: %+v", bar)
	}
}

func TestVoiceParticipantStatusUpdatesDeterministic(t *testing.T) {
	shell := newShellWithIdentity(t)
	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert voice channel: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.SelectChannel("voice"); err != nil {
		t.Fatalf("select channel: %v", err)
	}
	if err := shell.StartVoiceSession(); err != nil {
		t.Fatalf("start voice session: %v", err)
	}

	if err := shell.VoiceParticipantJoin("user-b", "Beta", false); err != nil {
		t.Fatalf("join user-b: %v", err)
	}
	if err := shell.VoiceParticipantJoin("user-a", "Alpha", true); err != nil {
		t.Fatalf("join user-a: %v", err)
	}
	if err := shell.UpdateVoiceParticipantStatus("user-b", true, true, false, VoiceConnectionConnecting); err != nil {
		t.Fatalf("update user-b: %v", err)
	}
	if err := shell.UpdateVoiceParticipantStatus("user-a", false, false, true, VoiceConnectionDisconnected); err != nil {
		t.Fatalf("update user-a: %v", err)
	}

	bar := shell.PersistentVoiceBar()
	if !bar.Visible {
		t.Fatalf("expected voice bar visible, got hidden reason=%s", bar.HiddenReason)
	}
	if len(bar.Participants) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(bar.Participants))
	}
	if bar.Participants[0].ID != "user-a" || bar.Participants[1].ID != "user-b" {
		t.Fatalf("participants not sorted deterministically: %+v", bar.Participants)
	}
	userA := bar.Participants[0]
	if userA.Display != "Alpha" || userA.Speaking || userA.Muted || !userA.Deafened || userA.ConnectionStatus != VoiceConnectionDisconnected || !userA.Self {
		t.Fatalf("unexpected user-a state: %+v", userA)
	}
	userB := bar.Participants[1]
	if userB.Display != "Beta" || !userB.Speaking || !userB.Muted || userB.Deafened || userB.ConnectionStatus != VoiceConnectionConnecting || userB.Self {
		t.Fatalf("unexpected user-b state: %+v", userB)
	}
}

func TestVoiceSessionSwitchAndLeaveBehavior(t *testing.T) {
	shell := newShellWithIdentity(t)
	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert srv-1: %v", err)
	}
	if err := shell.UpsertServer("srv-2", "Beta"); err != nil {
		t.Fatalf("upsert srv-2: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert voice channel: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-2", "lobby", "Lobby", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert lobby channel: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select srv-1: %v", err)
	}
	if err := shell.SelectChannel("voice"); err != nil {
		t.Fatalf("select voice: %v", err)
	}
	if err := shell.StartVoiceSession(); err != nil {
		t.Fatalf("start voice session: %v", err)
	}
	if err := shell.VoiceParticipantJoin("user-a", "Alpha", true); err != nil {
		t.Fatalf("join user-a: %v", err)
	}

	if err := shell.SelectServer("srv-2"); err != nil {
		t.Fatalf("select srv-2: %v", err)
	}
	bar := shell.PersistentVoiceBar()
	if bar.Visible {
		t.Fatalf("expected voice bar hidden after server switch, got visible")
	}
	if bar.HiddenReason != VoiceBarHiddenReasonNoActiveSession {
		t.Fatalf("unexpected hidden reason after server switch: %s", bar.HiddenReason)
	}

	if err := shell.SelectChannel("lobby"); err != nil {
		t.Fatalf("select lobby: %v", err)
	}
	if err := shell.StartVoiceSession(); err != nil {
		t.Fatalf("start lobby session: %v", err)
	}
	if err := shell.VoiceParticipantJoin("user-b", "Beta", false); err != nil {
		t.Fatalf("join user-b: %v", err)
	}
	if err := shell.LeaveVoiceSession(); err != nil {
		t.Fatalf("leave voice session: %v", err)
	}
	bar = shell.PersistentVoiceBar()
	if bar.Visible {
		t.Fatalf("expected voice bar hidden after leave, got visible")
	}
	if bar.HiddenReason != VoiceBarHiddenReasonNoActiveSession {
		t.Fatalf("unexpected hidden reason after leave: %s", bar.HiddenReason)
	}
}

func TestVoiceParticipantViewAndVoiceBarComponent(t *testing.T) {
	shell := newShellWithIdentity(t)
	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert voice channel: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.SelectChannel("voice"); err != nil {
		t.Fatalf("select channel: %v", err)
	}
	if err := shell.StartVoiceSession(); err != nil {
		t.Fatalf("start voice session: %v", err)
	}
	if err := shell.VoiceParticipantJoin("user-b", "Beta", false); err != nil {
		t.Fatalf("join user-b: %v", err)
	}
	if err := shell.VoiceParticipantJoin("user-a", "Alpha", true); err != nil {
		t.Fatalf("join user-a: %v", err)
	}
	if err := shell.UpdateVoiceParticipantStatus("user-a", false, true, true, VoiceConnectionConnected); err != nil {
		t.Fatalf("update user-a: %v", err)
	}
	if err := shell.SetPushToTalkMode(VoicePushToTalkHold); err != nil {
		t.Fatalf("set push-to-talk mode: %v", err)
	}
	if err := shell.SetSelfMute(true); err != nil {
		t.Fatalf("set self mute: %v", err)
	}
	if err := shell.SetSelfDeafen(true); err != nil {
		t.Fatalf("set self deafen: %v", err)
	}

	participantView := shell.VoiceParticipantView()
	if !participantView.Visible {
		t.Fatalf("expected participant view visible, got hidden reason=%s", participantView.HiddenReason)
	}
	if participantView.ServerID != "srv-1" || participantView.ChannelID != "voice" {
		t.Fatalf("unexpected participant view scope: %+v", participantView)
	}
	if participantView.ActiveRoute != RouteChannelView {
		t.Fatalf("expected active route channel view, got %s", participantView.ActiveRoute)
	}
	if !participantView.SessionRouteBound {
		t.Fatalf("expected session route bound for matching selection")
	}
	if participantView.ParticipantCount != 2 || len(participantView.Participants) != 2 {
		t.Fatalf("expected 2 participants in participant view, got count=%d len=%d", participantView.ParticipantCount, len(participantView.Participants))
	}
	if participantView.Participants[0].ID != "user-a" || participantView.Participants[1].ID != "user-b" {
		t.Fatalf("expected deterministic participant order, got %+v", participantView.Participants)
	}

	if err := shell.Navigate(RouteSettings); err != nil {
		t.Fatalf("navigate settings: %v", err)
	}
	participantView = shell.VoiceParticipantView()
	if participantView.ActiveRoute != RouteSettings {
		t.Fatalf("expected route settings in participant view, got %s", participantView.ActiveRoute)
	}
	if participantView.SessionRouteBound {
		t.Fatalf("expected session route bound false after navigating away")
	}

	barComponent := shell.VoiceBarComponent()
	if !barComponent.Visible {
		t.Fatalf("expected voice bar component visible, got hidden reason=%s", barComponent.HiddenReason)
	}
	if barComponent.ServerID != "srv-1" || barComponent.ChannelID != "voice" {
		t.Fatalf("unexpected voice bar component scope: %+v", barComponent)
	}
	if barComponent.ParticipantCount != 2 || len(barComponent.Participants) != 2 {
		t.Fatalf("expected 2 participants in voice bar component, got count=%d len=%d", barComponent.ParticipantCount, len(barComponent.Participants))
	}
	if !barComponent.SelfMuted || !barComponent.SelfDeafened {
		t.Fatalf("expected voice bar component to reflect self mute/deafen, got %+v", barComponent)
	}
	if barComponent.PushToTalkMode != VoicePushToTalkHold {
		t.Fatalf("expected hold push-to-talk mode, got %s", barComponent.PushToTalkMode)
	}

	if err := shell.LeaveVoiceSession(); err != nil {
		t.Fatalf("leave voice session: %v", err)
	}
	participantView = shell.VoiceParticipantView()
	if participantView.Visible {
		t.Fatalf("expected participant view hidden after leave")
	}
	if participantView.HiddenReason != VoiceBarHiddenReasonNoActiveSession {
		t.Fatalf("expected hidden reason no_active_session, got %s", participantView.HiddenReason)
	}
	barComponent = shell.VoiceBarComponent()
	if barComponent.Visible {
		t.Fatalf("expected voice bar component hidden after leave")
	}
	if barComponent.HiddenReason != VoiceBarHiddenReasonNoActiveSession {
		t.Fatalf("expected hidden reason no_active_session for bar component, got %s", barComponent.HiddenReason)
	}
}

func TestVoiceParticipantViewAndVoiceBarComponentNilShell(t *testing.T) {
	var nilShell *Shell

	participantView := nilShell.VoiceParticipantView()
	if participantView.Visible {
		t.Fatalf("expected nil-shell participant view hidden")
	}
	if participantView.HiddenReason != VoiceBarHiddenReasonNoActiveSession {
		t.Fatalf("unexpected nil-shell participant view hidden reason: %s", participantView.HiddenReason)
	}

	barComponent := nilShell.VoiceBarComponent()
	if barComponent.Visible {
		t.Fatalf("expected nil-shell voice bar component hidden")
	}
	if barComponent.HiddenReason != VoiceBarHiddenReasonNoActiveSession {
		t.Fatalf("unexpected nil-shell voice bar component hidden reason: %s", barComponent.HiddenReason)
	}
}

func TestVoiceSessionGuards(t *testing.T) {
	var nilShell *Shell
	if err := nilShell.StartVoiceSession(); !errors.Is(err, ErrShellRequired) {
		t.Fatalf("expected ErrShellRequired start voice, got %v", err)
	}
	if err := nilShell.VoiceParticipantJoin("user", "User", true); !errors.Is(err, ErrShellRequired) {
		t.Fatalf("expected ErrShellRequired join, got %v", err)
	}

	shell := newShellWithIdentity(t)
	if err := shell.VoiceParticipantJoin("user", "User", true); !errors.Is(err, ErrVoiceSessionInactive) {
		t.Fatalf("expected ErrVoiceSessionInactive, got %v", err)
	}
	if err := shell.UpdateVoiceParticipantStatus("user", false, false, false, VoiceConnectionConnected); !errors.Is(err, ErrVoiceSessionInactive) {
		t.Fatalf("expected ErrVoiceSessionInactive update, got %v", err)
	}
	if err := shell.VoiceParticipantLeave("user"); !errors.Is(err, ErrVoiceSessionInactive) {
		t.Fatalf("expected ErrVoiceSessionInactive leave, got %v", err)
	}

	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert voice channel: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.SelectChannel("voice"); err != nil {
		t.Fatalf("select channel: %v", err)
	}
	if err := shell.StartVoiceSession(); err != nil {
		t.Fatalf("start voice session: %v", err)
	}

	if err := shell.VoiceParticipantJoin(" ", "User", true); !errors.Is(err, ErrVoiceParticipantMissing) {
		t.Fatalf("expected ErrVoiceParticipantMissing join, got %v", err)
	}
	if err := shell.VoiceParticipantLeave(" "); !errors.Is(err, ErrVoiceParticipantMissing) {
		t.Fatalf("expected ErrVoiceParticipantMissing leave, got %v", err)
	}
	if err := shell.UpdateVoiceParticipantStatus(" ", false, false, false, VoiceConnectionConnected); !errors.Is(err, ErrVoiceParticipantMissing) {
		t.Fatalf("expected ErrVoiceParticipantMissing update, got %v", err)
	}
	if err := shell.UpdateVoiceParticipantStatus("user", false, false, false, VoiceConnectionStatus("unknown")); !errors.Is(err, ErrVoiceConnectionInvalid) {
		t.Fatalf("expected ErrVoiceConnectionInvalid, got %v", err)
	}
	if err := shell.UpdateVoiceParticipantStatus("missing", false, false, false, VoiceConnectionConnected); !errors.Is(err, ErrVoiceParticipantUnknown) {
		t.Fatalf("expected ErrVoiceParticipantUnknown, got %v", err)
	}
}

func TestVoiceControlsLifecycleAndPushToTalk(t *testing.T) {
	shell := newShellWithIdentity(t)

	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert channel: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.SelectChannel("voice"); err != nil {
		t.Fatalf("select channel: %v", err)
	}
	if err := shell.StartVoiceSession(); err != nil {
		t.Fatalf("start voice session: %v", err)
	}

	if err := shell.SetPushToTalkMode(VoicePushToTalkHold); err != nil {
		t.Fatalf("set push-to-talk mode hold: %v", err)
	}
	if err := shell.PressPushToTalk(); err != nil {
		t.Fatalf("press push-to-talk: %v", err)
	}
	controls := shell.VoiceControls()
	if controls.PushToTalkMode != VoicePushToTalkHold || !controls.PushToTalkPressed {
		t.Fatalf("expected hold mode with pressed=true, got %+v", controls)
	}

	if err := shell.ReleasePushToTalk(); err != nil {
		t.Fatalf("release push-to-talk: %v", err)
	}
	if err := shell.SetSelfMute(true); err != nil {
		t.Fatalf("set self mute true: %v", err)
	}
	if err := shell.SetSelfDeafen(true); err != nil {
		t.Fatalf("set self deafen true: %v", err)
	}
	controls = shell.VoiceControls()
	if !controls.SelfMuted || !controls.SelfDeafened {
		t.Fatalf("expected self mute/deafen true, got %+v", controls)
	}

	if err := shell.SetPushToTalkMode(VoicePushToTalkDisabled); err != nil {
		t.Fatalf("set push-to-talk disabled: %v", err)
	}
	controls = shell.VoiceControls()
	if controls.PushToTalkPressed {
		t.Fatalf("expected push-to-talk pressed cleared in disabled mode, got %+v", controls)
	}

	if err := shell.LeaveVoiceSession(); err != nil {
		t.Fatalf("leave voice session: %v", err)
	}
	controls = shell.VoiceControls()
	if controls.SelfMuted || controls.SelfDeafened || controls.PushToTalkPressed {
		t.Fatalf("expected controls reset on leave, got %+v", controls)
	}
}

func TestVoiceDeviceSelectionBehavior(t *testing.T) {
	shell := newShellWithIdentity(t)

	if err := shell.SetVoiceDevices([]string{"mic-b", "mic-a", "mic-a"}, []string{"spk-b", "spk-a"}); err != nil {
		t.Fatalf("set voice devices: %v", err)
	}
	controls := shell.VoiceControls()
	if len(controls.AvailableInputDevices) != 2 || controls.AvailableInputDevices[0] != "mic-a" || controls.AvailableInputDevices[1] != "mic-b" {
		t.Fatalf("unexpected normalized input device list: %+v", controls.AvailableInputDevices)
	}
	if len(controls.AvailableOutputDevices) != 2 || controls.AvailableOutputDevices[0] != "spk-a" || controls.AvailableOutputDevices[1] != "spk-b" {
		t.Fatalf("unexpected normalized output device list: %+v", controls.AvailableOutputDevices)
	}
	if controls.InputDeviceID != "mic-a" || controls.OutputDeviceID != "spk-a" {
		t.Fatalf("expected fallback to first sorted devices, got input=%q output=%q", controls.InputDeviceID, controls.OutputDeviceID)
	}

	if err := shell.SwitchInputDevice("mic-b"); err != nil {
		t.Fatalf("switch input device: %v", err)
	}
	if err := shell.SwitchOutputDevice("spk-b"); err != nil {
		t.Fatalf("switch output device: %v", err)
	}
	controls = shell.VoiceControls()
	if controls.InputDeviceID != "mic-b" || controls.OutputDeviceID != "spk-b" {
		t.Fatalf("expected switched devices, got input=%q output=%q", controls.InputDeviceID, controls.OutputDeviceID)
	}
	if controls.InputSwitchInProgress || controls.OutputSwitchInProgress {
		t.Fatalf("expected switch-in-progress flags false after switch completion, got %+v", controls)
	}
}

func TestVoiceControlAndDeviceValidationGuards(t *testing.T) {
	var nilShell *Shell
	if err := nilShell.SetSelfMute(true); !errors.Is(err, ErrShellRequired) {
		t.Fatalf("expected ErrShellRequired set self mute, got %v", err)
	}
	if err := nilShell.SwitchInputDevice("mic"); !errors.Is(err, ErrShellRequired) {
		t.Fatalf("expected ErrShellRequired switch input device, got %v", err)
	}

	shell := newShellWithIdentity(t)

	if err := shell.SetSelfMute(true); !errors.Is(err, ErrVoiceControlUnavailable) {
		t.Fatalf("expected ErrVoiceControlUnavailable for set self mute without session, got %v", err)
	}
	if err := shell.SetSelfDeafen(true); !errors.Is(err, ErrVoiceControlUnavailable) {
		t.Fatalf("expected ErrVoiceControlUnavailable for set self deafen without session, got %v", err)
	}
	if err := shell.PressPushToTalk(); !errors.Is(err, ErrVoiceControlUnavailable) {
		t.Fatalf("expected ErrVoiceControlUnavailable for press push-to-talk without session, got %v", err)
	}

	if err := shell.SetPushToTalkMode(VoicePushToTalkMode("toggle")); !errors.Is(err, ErrVoicePushToTalkModeInvalid) {
		t.Fatalf("expected ErrVoicePushToTalkModeInvalid, got %v", err)
	}
	if err := shell.SwitchInputDevice(" "); !errors.Is(err, ErrVoiceDeviceIDMissing) {
		t.Fatalf("expected ErrVoiceDeviceIDMissing switch input, got %v", err)
	}
	if err := shell.SwitchOutputDevice(" "); !errors.Is(err, ErrVoiceDeviceIDMissing) {
		t.Fatalf("expected ErrVoiceDeviceIDMissing switch output, got %v", err)
	}

	if err := shell.SetVoiceDevices([]string{"mic-a"}, []string{"spk-a"}); err != nil {
		t.Fatalf("set initial voice devices: %v", err)
	}
	if err := shell.SwitchInputDevice("mic-z"); err != nil {
		t.Fatalf("unexpected switch input fallback error: %v", err)
	}
	if err := shell.SwitchOutputDevice("spk-z"); err != nil {
		t.Fatalf("unexpected switch output fallback error: %v", err)
	}
	controls := shell.VoiceControls()
	if controls.InputDeviceID != "mic-a" || controls.OutputDeviceID != "spk-a" {
		t.Fatalf("expected fallback to active devices, got input=%q output=%q", controls.InputDeviceID, controls.OutputDeviceID)
	}
	if err := shell.SetVoiceDevices([]string{}, []string{"spk-a"}); !errors.Is(err, ErrVoiceInputDeviceUnknown) {
		t.Fatalf("expected ErrVoiceInputDeviceUnknown for empty input list, got %v", err)
	}
	if err := shell.SetVoiceDevices([]string{"mic-a"}, []string{}); !errors.Is(err, ErrVoiceOutputDeviceUnknown) {
		t.Fatalf("expected ErrVoiceOutputDeviceUnknown for empty output list, got %v", err)
	}
}

func TestFirstLaunchGuidedFlowFromBlankState(t *testing.T) {
	shell := NewShell()

	if got := shell.State().CurrentRoute; got != RouteIdentitySetup {
		t.Fatalf("expected blank state to start at identity setup, got %q", got)
	}
	if err := shell.Navigate(RouteServerList); !errors.Is(err, ErrIdentitySetupRequired) {
		t.Fatalf("expected identity guard before server list, got %v", err)
	}

	if err := shell.SetIdentity("alice"); err != nil {
		t.Fatalf("set identity: %v", err)
	}
	if err := shell.Navigate(RouteServerList); err != nil {
		t.Fatalf("navigate server list after identity: %v", err)
	}

	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "general", "General", ChannelTypeText); err != nil {
		t.Fatalf("upsert text channel: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert voice channel: %v", err)
	}

	if err := shell.Navigate(RouteChannelView); !errors.Is(err, ErrServerSelectionRequired) {
		t.Fatalf("expected server selection guard for channel view, got %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select server: %v", err)
	}
	if err := shell.Navigate(RouteChannelView); !errors.Is(err, ErrChannelSelectionMissing) {
		t.Fatalf("expected channel selection guard for channel view, got %v", err)
	}
	if err := shell.SelectChannel("general"); err != nil {
		t.Fatalf("select text channel: %v", err)
	}
	if err := shell.Navigate(RouteChannelView); err != nil {
		t.Fatalf("navigate channel view: %v", err)
	}

	if err := shell.StartVoiceSessionForScope("srv-1", "voice"); err != nil {
		t.Fatalf("start voice session for scoped channel: %v", err)
	}
	bar := shell.PersistentVoiceBar()
	if !bar.Visible {
		t.Fatalf("expected persistent voice bar to be visible with active scoped session, hidden reason=%s", bar.HiddenReason)
	}
	if bar.ServerID != "srv-1" || bar.ChannelID != "voice" {
		t.Fatalf("unexpected voice bar scope: %+v", bar)
	}

	if err := shell.Navigate(RouteSettings); err != nil {
		t.Fatalf("navigate settings: %v", err)
	}
	settings := shell.SettingsView()
	if settings.Profile.DisplayName != "" || settings.Audio.InputVolume != DefaultAudioVolume {
		t.Fatalf("unexpected default settings payload after first-launch flow: %+v", settings)
	}
}

func TestFirstLaunchInterruptedFlowRecovery(t *testing.T) {
	shell := NewShell()
	if err := shell.SetIdentity("alice"); err != nil {
		t.Fatalf("set identity: %v", err)
	}
	if err := shell.UpsertServer("srv-1", "Alpha"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertServer("srv-2", "Beta"); err != nil {
		t.Fatalf("upsert server: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "general", "General", ChannelTypeText); err != nil {
		t.Fatalf("upsert srv-1/general: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-1", "voice", "Voice", ChannelTypeVoice); err != nil {
		t.Fatalf("upsert srv-1/voice: %v", err)
	}
	if err := shell.UpsertChannelWithType("srv-2", "ops", "Ops", ChannelTypeText); err != nil {
		t.Fatalf("upsert srv-2/ops: %v", err)
	}

	if err := shell.SaveDraft("srv-1", "general", "draft-msg"); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("select srv-1: %v", err)
	}
	if err := shell.SelectChannel("voice"); err != nil {
		t.Fatalf("select srv-1/voice: %v", err)
	}
	if err := shell.StartVoiceSession(); err != nil {
		t.Fatalf("start voice session: %v", err)
	}
	if err := shell.Navigate(RouteSettings); err != nil {
		t.Fatalf("navigate settings: %v", err)
	}

	if err := shell.SelectServer("srv-2"); err != nil {
		t.Fatalf("switch to srv-2: %v", err)
	}
	bar := shell.PersistentVoiceBar()
	if bar.Visible {
		t.Fatalf("expected voice session to terminate when switching server, got visible bar: %+v", bar)
	}
	if bar.HiddenReason != VoiceBarHiddenReasonNoActiveSession {
		t.Fatalf("expected no_active_session hidden reason after interruption, got %s", bar.HiddenReason)
	}

	if err := shell.SelectServer("srv-1"); err != nil {
		t.Fatalf("recover srv-1 selection: %v", err)
	}
	if err := shell.SelectChannel("general"); err != nil {
		t.Fatalf("recover text channel selection: %v", err)
	}
	draft, err := shell.LoadDraft("srv-1", "general")
	if err != nil {
		t.Fatalf("load recovered draft: %v", err)
	}
	if draft != "draft-msg" {
		t.Fatalf("expected recovered draft to persist across interruption, got %q", draft)
	}

	if err := shell.Navigate(RouteChannelView); err != nil {
		t.Fatalf("recover navigation to channel view: %v", err)
	}
	state := shell.State()
	if state.Subscription.ServerID != "srv-1" || state.Subscription.ChannelID != "general" {
		t.Fatalf("expected recovered subscription on srv-1/general, got %+v", state.Subscription)
	}
}

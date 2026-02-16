package ui

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Route represents a top-level shell entry point for the v0.1 user journey.
type Route string

const (
	RouteIdentitySetup Route = "identity_setup"
	RouteServerList    Route = "server_list"
	RouteChannelView   Route = "channel_view"
	RouteSettings      Route = "settings"
)

var (
	ErrShellRequired                       = errors.New("shell required")
	ErrRouteInvalid                        = errors.New("invalid route")
	ErrIdentitySetupRequired               = errors.New("identity setup required")
	ErrServerSelectionRequired             = errors.New("server selection required")
	ErrChannelSelectionMissing             = errors.New("channel selection required")
	ErrChannelTypeInvalid                  = errors.New("channel type invalid")
	ErrComposerMessageEmpty                = errors.New("composer message empty")
	ErrComposerMessageTooLong              = errors.New("composer message too long")
	ErrVoiceSessionInactive                = errors.New("voice session inactive")
	ErrVoiceParticipantMissing             = errors.New("voice participant missing")
	ErrVoiceParticipantUnknown             = errors.New("voice participant unknown")
	ErrVoiceConnectionInvalid              = errors.New("voice connection status invalid")
	ErrVoiceControlUnavailable             = errors.New("voice controls unavailable without active voice session")
	ErrVoicePushToTalkModeInvalid          = errors.New("voice push-to-talk mode invalid")
	ErrVoiceDeviceIDMissing                = errors.New("voice device id missing")
	ErrVoiceInputDeviceUnknown             = errors.New("voice input device unknown")
	ErrVoiceOutputDeviceUnknown            = errors.New("voice output device unknown")
	ErrNetworkPathStatusInvalid            = errors.New("network path status invalid")
	ErrNetworkPathInvalid                  = errors.New("network path invalid")
	ErrNetworkReasonClassInvalid           = errors.New("network reason class invalid")
	ErrDiagnosticLimitInvalid              = errors.New("diagnostic retention limit invalid")
	ErrDiagnosticCategoryMissing           = errors.New("diagnostic category missing")
	ErrDiagnosticExportUserTriggerRequired = errors.New("diagnostic export requires explicit user trigger")
	ErrMessageRenderEmpty                  = errors.New("message render input empty")
	ErrReplyReferenceIDMissing             = errors.New("reply reference id missing")
)

const DefaultComposerMaxLength = 4000

const (
	DefaultAudioVolume              = 70
	DefaultDiagnosticRetentionLimit = 128
	DefaultVoiceInputDeviceID       = "default-input"
	DefaultVoiceOutputDeviceID      = "default-output"
)

// VirtualMessageWindow captures a deterministic message render window and metadata.
// Start and End use half-open indexing [Start, End).
type VirtualMessageWindow[T any] struct {
	Items        []T
	Start        int
	End          int
	Total        int
	HasMoreAbove bool
	HasMoreBelow bool
}

// ComposeVirtualMessageWindow returns a clamped window over message inventory.
func ComposeVirtualMessageWindow[T any](inventory []T, anchor, windowSize int) VirtualMessageWindow[T] {
	total := len(inventory)
	if total == 0 {
		return VirtualMessageWindow[T]{}
	}

	if windowSize < 1 {
		windowSize = 1
	}
	if windowSize > total {
		windowSize = total
	}

	if anchor < 0 {
		anchor = 0
	}
	if anchor >= total {
		anchor = total - 1
	}

	start := anchor
	maxStart := total - windowSize
	if start > maxStart {
		start = maxStart
	}
	end := start + windowSize

	items := append([]T(nil), inventory[start:end]...)
	return VirtualMessageWindow[T]{
		Items:        items,
		Start:        start,
		End:          end,
		Total:        total,
		HasMoreAbove: start > 0,
		HasMoreBelow: end < total,
	}
}

// ComposerKeyResult captures deterministic key handling output for the composer.
type ComposerKeyResult struct {
	Draft       string
	ShouldSend  bool
	InsertedNew bool
}

// MarkdownTokenType captures the supported markdown subset render primitives.
type MarkdownTokenType string

const (
	MarkdownTokenText      MarkdownTokenType = "text"
	MarkdownTokenBold      MarkdownTokenType = "bold"
	MarkdownTokenItalic    MarkdownTokenType = "italic"
	MarkdownTokenCode      MarkdownTokenType = "code"
	MarkdownTokenLineBreak MarkdownTokenType = "line_break"
)

// MarkdownToken is a deterministic render token for message text.
type MarkdownToken struct {
	Type  MarkdownTokenType
	Value string
}

// ReplyReference stores deterministic reply metadata for UI rendering.
type ReplyReference struct {
	MessageID      string
	AuthorDisplay  string
	Excerpt        string
	Truncated      bool
	DisplaySummary string
}

// RenderedMessage is the deterministic markdown+reply render payload for a chat message.
type RenderedMessage struct {
	PlainText           string
	Tokens              []MarkdownToken
	UnsupportedMarkdown bool
	Reply               *ReplyReference
}

var htmlEscapeReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#39;",
)

// HandleComposerEnter applies Enter/Shift+Enter behavior to a composer draft.
func HandleComposerEnter(draft string, shiftHeld bool) ComposerKeyResult {
	if shiftHeld {
		return ComposerKeyResult{Draft: draft + "\n", InsertedNew: true}
	}
	return ComposerKeyResult{Draft: draft, ShouldSend: true}
}

// ValidateComposerMessage validates outgoing message content and returns normalized text.
func ValidateComposerMessage(draft string, maxLength int) (string, error) {
	normalized := strings.TrimSpace(draft)
	if normalized == "" {
		return "", ErrComposerMessageEmpty
	}
	if maxLength <= 0 {
		maxLength = DefaultComposerMaxLength
	}
	if utf8.RuneCountInString(normalized) > maxLength {
		return "", fmt.Errorf("%w: max=%d", ErrComposerMessageTooLong, maxLength)
	}
	return normalized, nil
}

// RenderMessageMarkdownSubset renders a deterministic markdown subset for v0.1.
// Supported inline tokens: **bold**, *italic*, and `code`.
// Unsupported markdown constructs degrade to escaped literal text and set UnsupportedMarkdown=true.
func RenderMessageMarkdownSubset(input string) (RenderedMessage, error) {
	normalized := strings.TrimSpace(input)
	if normalized == "" {
		return RenderedMessage{}, ErrMessageRenderEmpty
	}

	lines := strings.Split(normalized, "\n")
	tokens := make([]MarkdownToken, 0, len(lines)*2)
	plainLines := make([]string, 0, len(lines))
	unsupported := false

	for i, line := range lines {
		lineTokens, linePlain, lineUnsupported := parseMarkdownLine(line)
		tokens = append(tokens, lineTokens...)
		plainLines = append(plainLines, linePlain)
		unsupported = unsupported || lineUnsupported
		if i < len(lines)-1 {
			tokens = append(tokens, MarkdownToken{Type: MarkdownTokenLineBreak, Value: "\n"})
		}
	}

	return RenderedMessage{
		PlainText:           strings.Join(plainLines, "\n"),
		Tokens:              tokens,
		UnsupportedMarkdown: unsupported,
	}, nil
}

// BuildReplyReference produces deterministic reply metadata with sanitization and excerpt truncation.
func BuildReplyReference(messageID, authorDisplay, excerpt string, maxExcerptRunes int) (*ReplyReference, error) {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return nil, ErrReplyReferenceIDMissing
	}
	authorDisplay = strings.TrimSpace(authorDisplay)
	if authorDisplay == "" {
		authorDisplay = "unknown"
	}
	excerpt = strings.TrimSpace(excerpt)
	if maxExcerptRunes <= 0 {
		maxExcerptRunes = 80
	}

	truncatedExcerpt, truncated := truncateRunes(excerpt, maxExcerptRunes)
	escapedAuthor := sanitizeMarkdownLiteral(authorDisplay)
	escapedExcerpt := sanitizeMarkdownLiteral(truncatedExcerpt)
	summary := "↪ " + escapedAuthor
	if escapedExcerpt != "" {
		summary += ": " + escapedExcerpt
	}
	if truncated {
		summary += "…"
	}

	return &ReplyReference{
		MessageID:      messageID,
		AuthorDisplay:  escapedAuthor,
		Excerpt:        escapedExcerpt,
		Truncated:      truncated,
		DisplaySummary: summary,
	}, nil
}

func parseMarkdownLine(line string) ([]MarkdownToken, string, bool) {
	runes := []rune(line)
	tokens := make([]MarkdownToken, 0, 8)
	plain := strings.Builder{}
	var segment strings.Builder
	unsupported := false

	flushText := func() {
		if segment.Len() == 0 {
			return
		}
		text := sanitizeMarkdownLiteral(segment.String())
		tokens = append(tokens, MarkdownToken{Type: MarkdownTokenText, Value: text})
		plain.WriteString(text)
		segment.Reset()
	}

	for i := 0; i < len(runes); {
		if runes[i] == '*' && i+1 < len(runes) && runes[i+1] == '*' {
			end := findMarker(runes, i+2, "**")
			if end > i+2 {
				flushText()
				content := sanitizeMarkdownLiteral(string(runes[i+2 : end]))
				tokens = append(tokens, MarkdownToken{Type: MarkdownTokenBold, Value: content})
				plain.WriteString(content)
				i = end + 2
				continue
			}
		}
		if runes[i] == '*' {
			end := findMarker(runes, i+1, "*")
			if end > i+1 {
				flushText()
				content := sanitizeMarkdownLiteral(string(runes[i+1 : end]))
				tokens = append(tokens, MarkdownToken{Type: MarkdownTokenItalic, Value: content})
				plain.WriteString(content)
				i = end + 1
				continue
			}
		}
		if runes[i] == '`' {
			end := findMarker(runes, i+1, "`")
			if end > i+1 {
				flushText()
				content := sanitizeMarkdownLiteral(string(runes[i+1 : end]))
				tokens = append(tokens, MarkdownToken{Type: MarkdownTokenCode, Value: content})
				plain.WriteString(content)
				i = end + 1
				continue
			}
		}

		if i == 0 {
			trimmed := strings.TrimSpace(string(runes))
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "-") {
				unsupported = true
			}
		}
		if runes[i] == '[' || runes[i] == ']' || runes[i] == '(' || runes[i] == ')' {
			unsupported = true
		}
		segment.WriteRune(runes[i])
		i++
	}

	flushText()
	return tokens, plain.String(), unsupported
}

func findMarker(runes []rune, start int, marker string) int {
	if marker == "*" || marker == "`" {
		m := rune(marker[0])
		for i := start; i < len(runes); i++ {
			if runes[i] == m {
				return i
			}
		}
		return -1
	}
	if marker == "**" {
		for i := start; i+1 < len(runes); i++ {
			if runes[i] == '*' && runes[i+1] == '*' {
				return i
			}
		}
	}
	return -1
}

func sanitizeMarkdownLiteral(v string) string {
	return htmlEscapeReplacer.Replace(v)
}

func truncateRunes(input string, maxRunes int) (string, bool) {
	runes := []rune(input)
	if len(runes) <= maxRunes {
		return input, false
	}
	return string(runes[:maxRunes]), true
}

// ServerSummary captures server rail state used by the shell baseline.
type ServerSummary struct {
	ID   string
	Name string
}

// ChannelType classifies channel entries for sidebar indicator selection.
type ChannelType string

const (
	ChannelTypeText  ChannelType = "text"
	ChannelTypeVoice ChannelType = "voice"
)

// ChannelSummary captures channel sidebar state for a specific server.
type ChannelSummary struct {
	ServerID string
	ID       string
	Name     string
	Type     ChannelType
}

// ServerRailItem is the render model for the server rail.
type ServerRailItem struct {
	ID       string
	Name     string
	Selected bool
}

// ChannelSidebarItem is the render model for the selected server channel list.
type ChannelSidebarItem struct {
	ID        string
	Name      string
	Type      ChannelType
	Indicator string
	Selected  bool
}

// SubscriptionState tracks the active channel subscription target.
type SubscriptionState struct {
	ServerID  string
	ChannelID string
}

// VoiceConnectionStatus represents participant connectivity state in an active voice session.
type VoiceConnectionStatus string

const (
	VoiceConnectionConnected    VoiceConnectionStatus = "connected"
	VoiceConnectionConnecting   VoiceConnectionStatus = "connecting"
	VoiceConnectionDisconnected VoiceConnectionStatus = "disconnected"
)

// VoiceParticipantTileState is the deterministic participant tile model for voice channels.
type VoiceParticipantTileState struct {
	ID               string
	Display          string
	Speaking         bool
	Muted            bool
	Deafened         bool
	ConnectionStatus VoiceConnectionStatus
	Self             bool
}

// VoiceSessionState tracks an active voice context anchored to server/channel selection.
type VoiceSessionState struct {
	Active       bool
	ServerID     string
	ChannelID    string
	Participants []VoiceParticipantTileState
}

// VoiceBarHiddenReason provides deterministic hidden-state explanations.
type VoiceBarHiddenReason string

const (
	VoiceBarHiddenReasonNone              VoiceBarHiddenReason = "none"
	VoiceBarHiddenReasonNoActiveSession   VoiceBarHiddenReason = "no_active_session"
	VoiceBarHiddenReasonSessionScopeEmpty VoiceBarHiddenReason = "session_scope_empty"
)

// VoiceBarState models the persistent voice bar, independent from current route.
type VoiceBarState struct {
	Visible      bool
	HiddenReason VoiceBarHiddenReason
	ServerID     string
	ChannelID    string
	Participants []VoiceParticipantTileState
}

// VoiceParticipantViewState models the participant panel component for the active voice scope.
type VoiceParticipantViewState struct {
	Visible           bool
	HiddenReason      VoiceBarHiddenReason
	ServerID          string
	ChannelID         string
	Participants      []VoiceParticipantTileState
	ParticipantCount  int
	ActiveRoute       Route
	SessionRouteBound bool
}

// VoiceBarComponentState models the persistent voice bar component payload.
type VoiceBarComponentState struct {
	Visible          bool
	HiddenReason     VoiceBarHiddenReason
	ServerID         string
	ChannelID        string
	Participants     []VoiceParticipantTileState
	ParticipantCount int
	SelfMuted        bool
	SelfDeafened     bool
	PushToTalkMode   VoicePushToTalkMode
}

// VoicePushToTalkMode describes push-to-talk interaction semantics.
type VoicePushToTalkMode string

const (
	VoicePushToTalkDisabled VoicePushToTalkMode = "disabled"
	VoicePushToTalkHold     VoicePushToTalkMode = "hold"
)

// VoiceControlState models deterministic local voice controls and selected devices.
type VoiceControlState struct {
	SelfMuted              bool
	SelfDeafened           bool
	PushToTalkMode         VoicePushToTalkMode
	PushToTalkPressed      bool
	InputDeviceID          string
	OutputDeviceID         string
	AvailableInputDevices  []string
	AvailableOutputDevices []string
	InputSwitchInProgress  bool
	OutputSwitchInProgress bool
}

// SettingsProfileState models editable profile settings.
type SettingsProfileState struct {
	DisplayName   string
	StatusMessage string
}

// SettingsAudioState models editable audio settings.
type SettingsAudioState struct {
	InputMuted       bool
	OutputMuted      bool
	InputVolume      int
	OutputVolume     int
	NoiseSuppression bool
	EchoCancellation bool
}

// SettingsState captures deterministic settings view state and persistence metadata.
type SettingsState struct {
	Profile      SettingsProfileState
	Audio        SettingsAudioState
	SavedVersion uint64
	Dirty        bool
}

// NetworkPathStatus represents availability of a concrete network path.
type NetworkPathStatus string

const (
	NetworkPathStatusUnavailable NetworkPathStatus = "unavailable"
	NetworkPathStatusConnecting  NetworkPathStatus = "connecting"
	NetworkPathStatusActive      NetworkPathStatus = "active"
)

// NetworkPath indicates the currently active transport path.
type NetworkPath string

const (
	NetworkPathNone   NetworkPath = "none"
	NetworkPathDirect NetworkPath = "direct"
	NetworkPathRelay  NetworkPath = "relay"
)

// NetworkReasonClass classifies direct/relay path behavior for diagnostics.
type NetworkReasonClass string

const (
	NetworkReasonClassNone                NetworkReasonClass = "none"
	NetworkReasonClassNATTraversal        NetworkReasonClass = "nat_traversal"
	NetworkReasonClassFirewallPolicy      NetworkReasonClass = "firewall_policy"
	NetworkReasonClassTimeout             NetworkReasonClass = "timeout"
	NetworkReasonClassConnectivityDegrade NetworkReasonClass = "connectivity_degrade"
	NetworkReasonClassUnknown             NetworkReasonClass = "unknown"
)

// NetworkDiagnosticsState is the deterministic settings diagnostics model for direct/relay status.
type NetworkDiagnosticsState struct {
	DirectPathStatus NetworkPathStatus
	RelayPathStatus  NetworkPathStatus
	ActivePath       NetworkPath
	ReasonClass      NetworkReasonClass
	Summary          string
}

// DiagnosticRecord is a redacted local diagnostic item retained in a bounded ring buffer.
type DiagnosticRecord struct {
	Category     string
	Message      string
	Metadata     map[string]string
	OccurredUnix int64
}

// DiagnosticRedactionPolicy documents the privacy-preserving export policy.
type DiagnosticRedactionPolicy struct {
	Mode         string
	RedactedKeys []string
	Token        string
}

// DiagnosticExportEnvelope is the explicit local export payload.
type DiagnosticExportEnvelope struct {
	Version         string
	TriggeredByUser bool
	Reason          string
	RetentionLimit  int
	RecordCount     int
	GeneratedUnix   int64
	Redaction       DiagnosticRedactionPolicy
	Records         []DiagnosticRecord
}

// AppState is the global state model used by P10-T1 route transitions.
type AppState struct {
	CurrentRoute       Route
	IdentityDisplay    string
	Servers            []ServerSummary
	ChannelsByServer   map[string][]ChannelSummary
	DraftsByScope      map[string]string
	Settings           SettingsState
	NetworkDiagnostics NetworkDiagnosticsState
	VoiceSession       VoiceSessionState
	VoiceBar           VoiceBarState
	VoiceControls      VoiceControlState
	SelectedServerID   string
	SelectedChannelID  string
	Subscription       SubscriptionState
}

func (s AppState) hasIdentity() bool {
	return strings.TrimSpace(s.IdentityDisplay) != ""
}

// Shell holds the in-memory state for baseline route transitions.
type Shell struct {
	state                    AppState
	persistedSettings        SettingsState
	diagnosticRetentionLimit int
	diagnosticRecords        []DiagnosticRecord
}

// NewShell creates the baseline shell in identity setup route.
func NewShell() *Shell {
	return &Shell{
		state: AppState{
			CurrentRoute:       RouteIdentitySetup,
			Servers:            []ServerSummary{},
			ChannelsByServer:   map[string][]ChannelSummary{},
			DraftsByScope:      map[string]string{},
			Settings:           defaultSettingsState(),
			NetworkDiagnostics: defaultNetworkDiagnosticsState(),
			VoiceControls:      defaultVoiceControlState(),
		},
		persistedSettings:        defaultSettingsState(),
		diagnosticRetentionLimit: DefaultDiagnosticRetentionLimit,
		diagnosticRecords:        []DiagnosticRecord{},
	}
}

// EntryPoints returns the mandatory shell entry points for v0.1.
func EntryPoints() []Route {
	return []Route{
		RouteIdentitySetup,
		RouteServerList,
		RouteChannelView,
		RouteSettings,
	}
}

// State returns a defensive copy of the current shell state.
func (s *Shell) State() AppState {
	if s == nil {
		return AppState{}
	}
	copyState := s.state
	copyState.Servers = append([]ServerSummary(nil), s.state.Servers...)
	copyState.ChannelsByServer = make(map[string][]ChannelSummary, len(s.state.ChannelsByServer))
	for serverID, channels := range s.state.ChannelsByServer {
		copyState.ChannelsByServer[serverID] = append([]ChannelSummary(nil), channels...)
	}
	copyState.DraftsByScope = make(map[string]string, len(s.state.DraftsByScope))
	for key, draft := range s.state.DraftsByScope {
		copyState.DraftsByScope[key] = draft
	}
	copyState.VoiceSession.Participants = append([]VoiceParticipantTileState(nil), s.state.VoiceSession.Participants...)
	copyState.VoiceBar = s.PersistentVoiceBar()
	copyState.Settings = normalizeSettingsState(copyState.Settings)
	copyState.NetworkDiagnostics = normalizeNetworkDiagnosticsState(copyState.NetworkDiagnostics)
	return copyState
}

// SettingsView returns current deterministic settings state.
func (s *Shell) SettingsView() SettingsState {
	if s == nil {
		return defaultSettingsState()
	}
	return normalizeSettingsState(s.state.Settings)
}

// AdjustProfileSettings updates profile settings in memory.
func (s *Shell) AdjustProfileSettings(profile SettingsProfileState) error {
	if s == nil {
		return ErrShellRequired
	}
	next := s.state.Settings
	next.Profile = SettingsProfileState{
		DisplayName:   strings.TrimSpace(profile.DisplayName),
		StatusMessage: strings.TrimSpace(profile.StatusMessage),
	}
	next.Dirty = next != s.persistedSettings
	s.state.Settings = normalizeSettingsState(next)
	return nil
}

// AdjustAudioSettings updates audio settings in memory.
func (s *Shell) AdjustAudioSettings(audio SettingsAudioState) error {
	if s == nil {
		return ErrShellRequired
	}
	next := s.state.Settings
	next.Audio = audio
	next = normalizeSettingsState(next)
	next.Dirty = next != s.persistedSettings
	s.state.Settings = next
	return nil
}

// SaveSettings persists in-memory settings into shell-local durable state.
func (s *Shell) SaveSettings() error {
	if s == nil {
		return ErrShellRequired
	}
	next := normalizeSettingsState(s.state.Settings)
	next.SavedVersion++
	next.Dirty = false
	s.persistedSettings = next
	s.state.Settings = next
	return nil
}

// ReloadSettings restores persisted settings into the current settings view model.
func (s *Shell) ReloadSettings() error {
	if s == nil {
		return ErrShellRequired
	}
	s.state.Settings = normalizeSettingsState(s.persistedSettings)
	s.state.Settings.Dirty = false
	return nil
}

// NetworkDiagnosticsView returns direct/relay path diagnostics.
func (s *Shell) NetworkDiagnosticsView() NetworkDiagnosticsState {
	if s == nil {
		return defaultNetworkDiagnosticsState()
	}
	return normalizeNetworkDiagnosticsState(s.state.NetworkDiagnostics)
}

// SetNetworkDiagnostics updates direct/relay diagnostics and reason-class status.
func (s *Shell) SetNetworkDiagnostics(next NetworkDiagnosticsState) error {
	if s == nil {
		return ErrShellRequired
	}
	normalized, err := validateNetworkDiagnosticsState(next)
	if err != nil {
		return err
	}
	s.state.NetworkDiagnostics = normalized
	return nil
}

// SetDiagnosticRetentionLimit configures bounded local diagnostics retention.
func (s *Shell) SetDiagnosticRetentionLimit(limit int) error {
	if s == nil {
		return ErrShellRequired
	}
	if limit < 1 {
		return ErrDiagnosticLimitInvalid
	}
	s.diagnosticRetentionLimit = limit
	if len(s.diagnosticRecords) > limit {
		s.diagnosticRecords = append([]DiagnosticRecord(nil), s.diagnosticRecords[len(s.diagnosticRecords)-limit:]...)
	}
	return nil
}

// DiagnosticRetentionLimit returns current local diagnostics retention bound.
func (s *Shell) DiagnosticRetentionLimit() int {
	if s == nil || s.diagnosticRetentionLimit < 1 {
		return DefaultDiagnosticRetentionLimit
	}
	return s.diagnosticRetentionLimit
}

// AddDiagnosticRecord appends a diagnostic record into bounded local retention with redaction.
func (s *Shell) AddDiagnosticRecord(record DiagnosticRecord) error {
	if s == nil {
		return ErrShellRequired
	}
	record.Category = strings.TrimSpace(record.Category)
	if record.Category == "" {
		return ErrDiagnosticCategoryMissing
	}
	record = redactDiagnosticRecord(record)
	if s.diagnosticRetentionLimit < 1 {
		s.diagnosticRetentionLimit = DefaultDiagnosticRetentionLimit
	}
	s.diagnosticRecords = append(s.diagnosticRecords, record)
	if len(s.diagnosticRecords) > s.diagnosticRetentionLimit {
		s.diagnosticRecords = append([]DiagnosticRecord(nil), s.diagnosticRecords[len(s.diagnosticRecords)-s.diagnosticRetentionLimit:]...)
	}
	return nil
}

// ExportDiagnostics produces a local export envelope and requires explicit user-trigger control.
func (s *Shell) ExportDiagnostics(userTriggered bool, reason string) (DiagnosticExportEnvelope, error) {
	if s == nil {
		return DiagnosticExportEnvelope{}, ErrShellRequired
	}
	if !userTriggered {
		return DiagnosticExportEnvelope{}, ErrDiagnosticExportUserTriggerRequired
	}
	reason = strings.TrimSpace(reason)
	records := append([]DiagnosticRecord(nil), s.diagnosticRecords...)
	maxUnix := int64(0)
	for _, record := range records {
		if record.OccurredUnix > maxUnix {
			maxUnix = record.OccurredUnix
		}
	}
	return DiagnosticExportEnvelope{
		Version:         "v1",
		TriggeredByUser: true,
		Reason:          reason,
		RetentionLimit:  s.DiagnosticRetentionLimit(),
		RecordCount:     len(records),
		GeneratedUnix:   maxUnix,
		Redaction:       defaultDiagnosticRedactionPolicy(),
		Records:         records,
	}, nil
}

// StartVoiceSession anchors an active voice session to current selected server/channel.
func (s *Shell) StartVoiceSession() error {
	return s.StartVoiceSessionForSelection()
}

// StartVoiceSessionForSelection anchors an active voice session to the current selected server/channel.
func (s *Shell) StartVoiceSessionForSelection() error {
	if s == nil {
		return ErrShellRequired
	}
	if s.state.SelectedServerID == "" {
		return ErrServerSelectionRequired
	}
	if s.state.SelectedChannelID == "" {
		return ErrChannelSelectionMissing
	}
	if !s.hasServer(s.state.SelectedServerID) {
		return fmt.Errorf("%w: %s", ErrServerSelectionRequired, s.state.SelectedServerID)
	}
	if !s.hasChannel(s.state.SelectedServerID, s.state.SelectedChannelID) {
		return fmt.Errorf("%w: %s", ErrChannelSelectionMissing, s.state.SelectedChannelID)
	}
	s.state.VoiceSession = VoiceSessionState{
		Active:       true,
		ServerID:     s.state.SelectedServerID,
		ChannelID:    s.state.SelectedChannelID,
		Participants: []VoiceParticipantTileState{},
	}
	return nil
}

// StartVoiceSessionForScope anchors an active voice session to a specific server/channel scope
// without mutating current route/server/channel selection. This supports persistent voice bar
// visibility across navigation contexts.
func (s *Shell) StartVoiceSessionForScope(serverID, channelID string) error {
	if s == nil {
		return ErrShellRequired
	}
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		return ErrServerSelectionRequired
	}
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return ErrChannelSelectionMissing
	}
	if !s.hasServer(serverID) {
		return fmt.Errorf("%w: %s", ErrServerSelectionRequired, serverID)
	}
	if !s.hasChannel(serverID, channelID) {
		return fmt.Errorf("%w: %s", ErrChannelSelectionMissing, channelID)
	}
	s.state.VoiceSession = VoiceSessionState{
		Active:       true,
		ServerID:     serverID,
		ChannelID:    channelID,
		Participants: []VoiceParticipantTileState{},
	}
	return nil
}

// LeaveVoiceSession resets the active voice session state.
func (s *Shell) LeaveVoiceSession() error {
	if s == nil {
		return ErrShellRequired
	}
	s.clearVoiceSession()
	return nil
}

// VoiceParticipantJoin inserts or updates a participant tile in the active voice session.
func (s *Shell) VoiceParticipantJoin(participantID, display string, self bool) error {
	if s == nil {
		return ErrShellRequired
	}
	if !s.state.VoiceSession.Active {
		return ErrVoiceSessionInactive
	}
	participant, err := normalizeVoiceParticipantTileState(VoiceParticipantTileState{
		ID:               participantID,
		Display:          display,
		ConnectionStatus: VoiceConnectionConnected,
		Self:             self,
	})
	if err != nil {
		return err
	}
	s.state.VoiceSession.Participants = upsertVoiceParticipant(s.state.VoiceSession.Participants, participant)
	return nil
}

// VoiceParticipantLeave removes a participant tile from the active voice session.
func (s *Shell) VoiceParticipantLeave(participantID string) error {
	if s == nil {
		return ErrShellRequired
	}
	if !s.state.VoiceSession.Active {
		return ErrVoiceSessionInactive
	}
	participantID = strings.TrimSpace(participantID)
	if participantID == "" {
		return ErrVoiceParticipantMissing
	}
	before := len(s.state.VoiceSession.Participants)
	s.state.VoiceSession.Participants = removeVoiceParticipant(s.state.VoiceSession.Participants, participantID)
	if len(s.state.VoiceSession.Participants) == before {
		return fmt.Errorf("%w: %s", ErrVoiceParticipantUnknown, participantID)
	}
	return nil
}

// UpdateVoiceParticipantStatus applies deterministic status updates to an existing participant.
func (s *Shell) UpdateVoiceParticipantStatus(participantID string, speaking, muted, deafened bool, connectionStatus VoiceConnectionStatus) error {
	if s == nil {
		return ErrShellRequired
	}
	if !s.state.VoiceSession.Active {
		return ErrVoiceSessionInactive
	}
	participantID = strings.TrimSpace(participantID)
	if participantID == "" {
		return ErrVoiceParticipantMissing
	}
	normalizedStatus, err := normalizeVoiceConnectionStatus(connectionStatus)
	if err != nil {
		return err
	}
	for i := range s.state.VoiceSession.Participants {
		if s.state.VoiceSession.Participants[i].ID == participantID {
			s.state.VoiceSession.Participants[i].Speaking = speaking
			s.state.VoiceSession.Participants[i].Muted = muted
			s.state.VoiceSession.Participants[i].Deafened = deafened
			s.state.VoiceSession.Participants[i].ConnectionStatus = normalizedStatus
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrVoiceParticipantUnknown, participantID)
}

// VoiceControls returns a defensive copy of deterministic voice control state.
func (s *Shell) VoiceControls() VoiceControlState {
	if s == nil {
		return defaultVoiceControlState()
	}
	return normalizeVoiceControlState(s.state.VoiceControls)
}

// SetVoiceDevices replaces available device inventories and enforces active selections.
func (s *Shell) SetVoiceDevices(inputIDs, outputIDs []string) error {
	if s == nil {
		return ErrShellRequired
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	next.AvailableInputDevices = normalizeDeviceList(inputIDs)
	next.AvailableOutputDevices = normalizeDeviceList(outputIDs)
	if _, err := resolveDeviceSelection(next.AvailableInputDevices, next.InputDeviceID, ErrVoiceInputDeviceUnknown); err != nil {
		return err
	}
	if _, err := resolveDeviceSelection(next.AvailableOutputDevices, next.OutputDeviceID, ErrVoiceOutputDeviceUnknown); err != nil {
		return err
	}
	next.InputSwitchInProgress = false
	next.OutputSwitchInProgress = false
	s.state.VoiceControls = next
	return nil
}

// SetPushToTalkMode updates push-to-talk mode and normalizes pressed-state semantics.
func (s *Shell) SetPushToTalkMode(mode VoicePushToTalkMode) error {
	if s == nil {
		return ErrShellRequired
	}
	normalizedMode, err := normalizeVoicePushToTalkMode(mode)
	if err != nil {
		return err
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	next.PushToTalkMode = normalizedMode
	if normalizedMode == VoicePushToTalkDisabled {
		next.PushToTalkPressed = false
	}
	s.state.VoiceControls = next
	return nil
}

// SetSelfMute toggles local self-mute state in active voice sessions.
func (s *Shell) SetSelfMute(muted bool) error {
	if s == nil {
		return ErrShellRequired
	}
	if !s.state.VoiceSession.Active {
		return ErrVoiceControlUnavailable
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	next.SelfMuted = muted
	if !muted {
		next.PushToTalkPressed = false
	}
	s.state.VoiceControls = next
	return nil
}

// SetSelfDeafen toggles local self-deafen state in active voice sessions.
func (s *Shell) SetSelfDeafen(deafened bool) error {
	if s == nil {
		return ErrShellRequired
	}
	if !s.state.VoiceSession.Active {
		return ErrVoiceControlUnavailable
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	next.SelfDeafened = deafened
	s.state.VoiceControls = next
	return nil
}

// PressPushToTalk marks local push-to-talk press state when hold mode is active.
func (s *Shell) PressPushToTalk() error {
	if s == nil {
		return ErrShellRequired
	}
	if !s.state.VoiceSession.Active {
		return ErrVoiceControlUnavailable
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	if next.PushToTalkMode != VoicePushToTalkHold {
		return ErrVoicePushToTalkModeInvalid
	}
	next.PushToTalkPressed = true
	s.state.VoiceControls = next
	return nil
}

// ReleasePushToTalk clears local push-to-talk press state.
func (s *Shell) ReleasePushToTalk() error {
	if s == nil {
		return ErrShellRequired
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	next.PushToTalkPressed = false
	s.state.VoiceControls = next
	return nil
}

// SwitchInputDevice updates active input device with session-safe lifecycle semantics.
func (s *Shell) SwitchInputDevice(deviceID string) error {
	if s == nil {
		return ErrShellRequired
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return ErrVoiceDeviceIDMissing
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	resolved, err := resolveDeviceSelection(next.AvailableInputDevices, deviceID, ErrVoiceInputDeviceUnknown)
	if err != nil {
		return err
	}
	next.InputSwitchInProgress = true
	next.InputDeviceID = resolved
	next.InputSwitchInProgress = false
	s.state.VoiceControls = next
	return nil
}

// SwitchOutputDevice updates active output device with session-safe lifecycle semantics.
func (s *Shell) SwitchOutputDevice(deviceID string) error {
	if s == nil {
		return ErrShellRequired
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return ErrVoiceDeviceIDMissing
	}
	next := normalizeVoiceControlState(s.state.VoiceControls)
	resolved, err := resolveDeviceSelection(next.AvailableOutputDevices, deviceID, ErrVoiceOutputDeviceUnknown)
	if err != nil {
		return err
	}
	next.OutputSwitchInProgress = true
	next.OutputDeviceID = resolved
	next.OutputSwitchInProgress = false
	s.state.VoiceControls = next
	return nil
}

// PersistentVoiceBar returns a deterministic voice bar state independent of current route.
func (s *Shell) PersistentVoiceBar() VoiceBarState {
	if s == nil {
		return VoiceBarState{HiddenReason: VoiceBarHiddenReasonNoActiveSession}
	}
	if !s.state.VoiceSession.Active {
		return VoiceBarState{HiddenReason: VoiceBarHiddenReasonNoActiveSession}
	}
	if strings.TrimSpace(s.state.VoiceSession.ServerID) == "" || strings.TrimSpace(s.state.VoiceSession.ChannelID) == "" {
		return VoiceBarState{HiddenReason: VoiceBarHiddenReasonSessionScopeEmpty}
	}
	return VoiceBarState{
		Visible:      true,
		HiddenReason: VoiceBarHiddenReasonNone,
		ServerID:     s.state.VoiceSession.ServerID,
		ChannelID:    s.state.VoiceSession.ChannelID,
		Participants: append([]VoiceParticipantTileState(nil), s.state.VoiceSession.Participants...),
	}
}

// VoiceParticipantView returns a deterministic participant panel model for the active voice scope.
func (s *Shell) VoiceParticipantView() VoiceParticipantViewState {
	if s == nil {
		return VoiceParticipantViewState{HiddenReason: VoiceBarHiddenReasonNoActiveSession}
	}
	bar := s.PersistentVoiceBar()
	view := VoiceParticipantViewState{
		Visible:           bar.Visible,
		HiddenReason:      bar.HiddenReason,
		ServerID:          bar.ServerID,
		ChannelID:         bar.ChannelID,
		Participants:      append([]VoiceParticipantTileState(nil), bar.Participants...),
		ParticipantCount:  len(bar.Participants),
		ActiveRoute:       s.state.CurrentRoute,
		SessionRouteBound: s.state.VoiceSession.Active && s.state.CurrentRoute == RouteChannelView && s.state.SelectedServerID == s.state.VoiceSession.ServerID && s.state.SelectedChannelID == s.state.VoiceSession.ChannelID,
	}
	return view
}

// VoiceBarComponent returns a deterministic persistent voice bar component payload.
func (s *Shell) VoiceBarComponent() VoiceBarComponentState {
	if s == nil {
		return VoiceBarComponentState{HiddenReason: VoiceBarHiddenReasonNoActiveSession}
	}
	bar := s.PersistentVoiceBar()
	controls := normalizeVoiceControlState(s.state.VoiceControls)
	return VoiceBarComponentState{
		Visible:          bar.Visible,
		HiddenReason:     bar.HiddenReason,
		ServerID:         bar.ServerID,
		ChannelID:        bar.ChannelID,
		Participants:     append([]VoiceParticipantTileState(nil), bar.Participants...),
		ParticipantCount: len(bar.Participants),
		SelfMuted:        controls.SelfMuted,
		SelfDeafened:     controls.SelfDeafened,
		PushToTalkMode:   controls.PushToTalkMode,
	}
}

// DraftScopeKey creates a deterministic per-server/channel draft key.
func DraftScopeKey(serverID, channelID string) (string, error) {
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		return "", ErrServerSelectionRequired
	}
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return "", ErrChannelSelectionMissing
	}
	return serverID + "/" + channelID, nil
}

// SaveDraft stores a draft for a server/channel scope.
func (s *Shell) SaveDraft(serverID, channelID, draft string) error {
	if s == nil {
		return ErrShellRequired
	}
	key, err := DraftScopeKey(serverID, channelID)
	if err != nil {
		return err
	}
	if s.state.DraftsByScope == nil {
		s.state.DraftsByScope = map[string]string{}
	}
	s.state.DraftsByScope[key] = draft
	return nil
}

// LoadDraft returns a saved draft for a server/channel scope.
func (s *Shell) LoadDraft(serverID, channelID string) (string, error) {
	if s == nil {
		return "", ErrShellRequired
	}
	key, err := DraftScopeKey(serverID, channelID)
	if err != nil {
		return "", err
	}
	return s.state.DraftsByScope[key], nil
}

// ClearDraft removes a saved draft for a server/channel scope.
func (s *Shell) ClearDraft(serverID, channelID string) error {
	if s == nil {
		return ErrShellRequired
	}
	key, err := DraftScopeKey(serverID, channelID)
	if err != nil {
		return err
	}
	delete(s.state.DraftsByScope, key)
	return nil
}

// SetIdentity updates the global identity state used by route guards.
func (s *Shell) SetIdentity(display string) error {
	if s == nil {
		return ErrShellRequired
	}
	display = strings.TrimSpace(display)
	if display == "" {
		return ErrIdentitySetupRequired
	}
	s.state.IdentityDisplay = display
	return nil
}

// UpsertServer records a server in shell state.
func (s *Shell) UpsertServer(serverID, name string) error {
	if s == nil {
		return ErrShellRequired
	}
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		return ErrServerSelectionRequired
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = serverID
	}
	for i := range s.state.Servers {
		if s.state.Servers[i].ID == serverID {
			s.state.Servers[i].Name = name
			if _, ok := s.state.ChannelsByServer[serverID]; !ok {
				s.state.ChannelsByServer[serverID] = []ChannelSummary{}
			}
			return nil
		}
	}
	s.state.Servers = append(s.state.Servers, ServerSummary{ID: serverID, Name: name})
	s.state.ChannelsByServer[serverID] = []ChannelSummary{}
	return nil
}

// UpsertChannel records a channel in shell state for a server.
func (s *Shell) UpsertChannel(serverID, channelID, name string) error {
	return s.UpsertChannelWithType(serverID, channelID, name, ChannelTypeText)
}

// UpsertChannelWithType records a channel and its visual type classification.
func (s *Shell) UpsertChannelWithType(serverID, channelID, name string, channelType ChannelType) error {
	if s == nil {
		return ErrShellRequired
	}
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		return ErrServerSelectionRequired
	}
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return ErrChannelSelectionMissing
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = channelID
	}

	normalizedType, err := normalizeChannelType(channelType)
	if err != nil {
		return err
	}

	if !s.hasServer(serverID) {
		return fmt.Errorf("%w: %s", ErrServerSelectionRequired, serverID)
	}

	channels := s.state.ChannelsByServer[serverID]
	for i := range channels {
		if channels[i].ID == channelID {
			channels[i].Name = name
			channels[i].Type = normalizedType
			s.state.ChannelsByServer[serverID] = channels
			return nil
		}
	}

	s.state.ChannelsByServer[serverID] = append(channels, ChannelSummary{ServerID: serverID, ID: channelID, Name: name, Type: normalizedType})
	return nil
}

// SelectServer updates active server state and clears channel selection when changed.
func (s *Shell) SelectServer(serverID string) error {
	if s == nil {
		return ErrShellRequired
	}
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		return ErrServerSelectionRequired
	}
	for _, srv := range s.state.Servers {
		if srv.ID == serverID {
			if s.state.SelectedServerID != serverID {
				s.state.SelectedChannelID = ""
				s.clearVoiceSession()
				if s.state.CurrentRoute == RouteChannelView {
					s.state.CurrentRoute = RouteServerList
				}
			}
			s.state.SelectedServerID = serverID
			s.state.Subscription = SubscriptionState{ServerID: serverID, ChannelID: s.state.SelectedChannelID}
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrServerSelectionRequired, serverID)
}

// SelectChannel updates active channel state.
func (s *Shell) SelectChannel(channelID string) error {
	if s == nil {
		return ErrShellRequired
	}
	if s.state.SelectedServerID == "" {
		return ErrServerSelectionRequired
	}
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return ErrChannelSelectionMissing
	}
	if !s.hasChannel(s.state.SelectedServerID, channelID) {
		return fmt.Errorf("%w: %s", ErrChannelSelectionMissing, channelID)
	}
	s.state.SelectedChannelID = channelID
	s.state.Subscription = SubscriptionState{ServerID: s.state.SelectedServerID, ChannelID: channelID}
	s.state.CurrentRoute = RouteChannelView
	if s.state.VoiceSession.Active && (s.state.VoiceSession.ServerID != s.state.SelectedServerID || s.state.VoiceSession.ChannelID != channelID) {
		s.clearVoiceSession()
	}
	return nil
}

// ServerRail returns a deterministic server rail model from current shell state.
func (s *Shell) ServerRail() []ServerRailItem {
	if s == nil {
		return nil
	}
	rail := make([]ServerRailItem, 0, len(s.state.Servers))
	for _, server := range s.state.Servers {
		rail = append(rail, ServerRailItem{
			ID:       server.ID,
			Name:     server.Name,
			Selected: server.ID == s.state.SelectedServerID,
		})
	}
	return rail
}

// ChannelSidebar returns channel entries for the selected server.
func (s *Shell) ChannelSidebar() []ChannelSidebarItem {
	if s == nil || s.state.SelectedServerID == "" {
		return nil
	}
	channels := s.state.ChannelsByServer[s.state.SelectedServerID]
	sidebar := make([]ChannelSidebarItem, 0, len(channels))
	for _, channel := range channels {
		normalizedType, err := normalizeChannelType(channel.Type)
		if err != nil {
			normalizedType = ChannelTypeText
		}
		sidebar = append(sidebar, ChannelSidebarItem{
			ID:        channel.ID,
			Name:      channel.Name,
			Type:      normalizedType,
			Indicator: indicatorForChannelType(normalizedType),
			Selected:  channel.ID == s.state.SelectedChannelID,
		})
	}
	return sidebar
}

func normalizeChannelType(channelType ChannelType) (ChannelType, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(channelType)))
	switch normalized {
	case "", string(ChannelTypeText):
		return ChannelTypeText, nil
	case string(ChannelTypeVoice):
		return ChannelTypeVoice, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrChannelTypeInvalid, channelType)
	}
}

func indicatorForChannelType(channelType ChannelType) string {
	switch channelType {
	case ChannelTypeVoice:
		return "icon-voice"
	default:
		return "icon-text"
	}
}

func normalizeVoiceConnectionStatus(status VoiceConnectionStatus) (VoiceConnectionStatus, error) {
	normalized := VoiceConnectionStatus(strings.ToLower(strings.TrimSpace(string(status))))
	switch normalized {
	case "", VoiceConnectionConnected:
		return VoiceConnectionConnected, nil
	case VoiceConnectionConnecting, VoiceConnectionDisconnected:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrVoiceConnectionInvalid, status)
	}
}

func normalizeVoicePushToTalkMode(mode VoicePushToTalkMode) (VoicePushToTalkMode, error) {
	normalized := VoicePushToTalkMode(strings.ToLower(strings.TrimSpace(string(mode))))
	switch normalized {
	case "", VoicePushToTalkDisabled:
		return VoicePushToTalkDisabled, nil
	case VoicePushToTalkHold:
		return VoicePushToTalkHold, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrVoicePushToTalkModeInvalid, mode)
	}
}

func defaultVoiceControlState() VoiceControlState {
	return VoiceControlState{
		PushToTalkMode:         VoicePushToTalkDisabled,
		InputDeviceID:          DefaultVoiceInputDeviceID,
		OutputDeviceID:         DefaultVoiceOutputDeviceID,
		AvailableInputDevices:  []string{DefaultVoiceInputDeviceID},
		AvailableOutputDevices: []string{DefaultVoiceOutputDeviceID},
	}
}

func normalizeVoiceControlState(state VoiceControlState) VoiceControlState {
	mode, err := normalizeVoicePushToTalkMode(state.PushToTalkMode)
	if err != nil {
		mode = VoicePushToTalkDisabled
	}
	state.PushToTalkMode = mode
	state.AvailableInputDevices = normalizeDeviceList(state.AvailableInputDevices)
	state.AvailableOutputDevices = normalizeDeviceList(state.AvailableOutputDevices)
	state.InputDeviceID = strings.TrimSpace(state.InputDeviceID)
	state.OutputDeviceID = strings.TrimSpace(state.OutputDeviceID)
	if resolved, err := resolveDeviceSelection(state.AvailableInputDevices, state.InputDeviceID, ErrVoiceInputDeviceUnknown); err == nil {
		state.InputDeviceID = resolved
	}
	if resolved, err := resolveDeviceSelection(state.AvailableOutputDevices, state.OutputDeviceID, ErrVoiceOutputDeviceUnknown); err == nil {
		state.OutputDeviceID = resolved
	}
	if state.PushToTalkMode == VoicePushToTalkDisabled {
		state.PushToTalkPressed = false
	}
	return state
}

func normalizeDeviceList(deviceIDs []string) []string {
	if len(deviceIDs) == 0 {
		return []string{}
	}
	uniq := map[string]struct{}{}
	for _, id := range deviceIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		uniq[trimmed] = struct{}{}
	}
	normalized := make([]string, 0, len(uniq))
	for id := range uniq {
		normalized = append(normalized, id)
	}
	sort.Strings(normalized)
	return normalized
}

func resolveDeviceSelection(devices []string, preferred string, baseErr error) (string, error) {
	preferred = strings.TrimSpace(preferred)
	if preferred == "" {
		if len(devices) == 0 {
			return "", baseErr
		}
		return devices[0], nil
	}
	for _, id := range devices {
		if id == preferred {
			return preferred, nil
		}
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("%w: %s", baseErr, preferred)
	}
	return devices[0], nil
}

func defaultSettingsState() SettingsState {
	return SettingsState{
		Profile: SettingsProfileState{},
		Audio: SettingsAudioState{
			InputVolume:      DefaultAudioVolume,
			OutputVolume:     DefaultAudioVolume,
			NoiseSuppression: true,
			EchoCancellation: true,
		},
	}
}

func normalizeSettingsState(state SettingsState) SettingsState {
	state.Profile.DisplayName = strings.TrimSpace(state.Profile.DisplayName)
	state.Profile.StatusMessage = strings.TrimSpace(state.Profile.StatusMessage)
	state.Audio.InputVolume = clampPercent(state.Audio.InputVolume)
	state.Audio.OutputVolume = clampPercent(state.Audio.OutputVolume)
	return state
}

func clampPercent(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func defaultNetworkDiagnosticsState() NetworkDiagnosticsState {
	return NetworkDiagnosticsState{
		DirectPathStatus: NetworkPathStatusUnavailable,
		RelayPathStatus:  NetworkPathStatusUnavailable,
		ActivePath:       NetworkPathNone,
		ReasonClass:      NetworkReasonClassNone,
	}
}

func normalizeNetworkDiagnosticsState(state NetworkDiagnosticsState) NetworkDiagnosticsState {
	validated, err := validateNetworkDiagnosticsState(state)
	if err != nil {
		return defaultNetworkDiagnosticsState()
	}
	return validated
}

func validateNetworkDiagnosticsState(state NetworkDiagnosticsState) (NetworkDiagnosticsState, error) {
	direct, err := normalizeNetworkPathStatus(state.DirectPathStatus)
	if err != nil {
		return NetworkDiagnosticsState{}, err
	}
	relay, err := normalizeNetworkPathStatus(state.RelayPathStatus)
	if err != nil {
		return NetworkDiagnosticsState{}, err
	}
	active, err := normalizeNetworkPath(state.ActivePath)
	if err != nil {
		return NetworkDiagnosticsState{}, err
	}
	reason, err := normalizeNetworkReasonClass(state.ReasonClass)
	if err != nil {
		return NetworkDiagnosticsState{}, err
	}
	return NetworkDiagnosticsState{
		DirectPathStatus: direct,
		RelayPathStatus:  relay,
		ActivePath:       active,
		ReasonClass:      reason,
		Summary:          strings.TrimSpace(state.Summary),
	}, nil
}

func normalizeNetworkPathStatus(status NetworkPathStatus) (NetworkPathStatus, error) {
	normalized := NetworkPathStatus(strings.ToLower(strings.TrimSpace(string(status))))
	switch normalized {
	case "", NetworkPathStatusUnavailable:
		return NetworkPathStatusUnavailable, nil
	case NetworkPathStatusConnecting, NetworkPathStatusActive:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrNetworkPathStatusInvalid, status)
	}
}

func normalizeNetworkPath(path NetworkPath) (NetworkPath, error) {
	normalized := NetworkPath(strings.ToLower(strings.TrimSpace(string(path))))
	switch normalized {
	case "", NetworkPathNone:
		return NetworkPathNone, nil
	case NetworkPathDirect, NetworkPathRelay:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrNetworkPathInvalid, path)
	}
}

func normalizeNetworkReasonClass(reason NetworkReasonClass) (NetworkReasonClass, error) {
	normalized := NetworkReasonClass(strings.ToLower(strings.TrimSpace(string(reason))))
	switch normalized {
	case "", NetworkReasonClassNone:
		return NetworkReasonClassNone, nil
	case NetworkReasonClassNATTraversal, NetworkReasonClassFirewallPolicy, NetworkReasonClassTimeout, NetworkReasonClassConnectivityDegrade, NetworkReasonClassUnknown:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrNetworkReasonClassInvalid, reason)
	}
}

func defaultDiagnosticRedactionPolicy() DiagnosticRedactionPolicy {
	return DiagnosticRedactionPolicy{
		Mode:         "preserve_shape_redact_sensitive_values",
		RedactedKeys: []string{"token", "secret", "password", "private_key", "session_key", "auth", "authorization"},
		Token:        "[REDACTED]",
	}
}

func redactDiagnosticRecord(record DiagnosticRecord) DiagnosticRecord {
	policy := defaultDiagnosticRedactionPolicy()
	record.Message = redactText(record.Message, policy.Token)
	if len(record.Metadata) == 0 {
		return record
	}
	metadata := make(map[string]string, len(record.Metadata))
	for key, value := range record.Metadata {
		lower := strings.ToLower(strings.TrimSpace(key))
		if isSensitiveKey(lower, policy.RedactedKeys) {
			metadata[key] = policy.Token
			continue
		}
		metadata[key] = redactText(value, policy.Token)
	}
	record.Metadata = metadata
	return record
}

func isSensitiveKey(key string, sensitiveKeys []string) bool {
	for _, candidate := range sensitiveKeys {
		if key == candidate {
			return true
		}
	}
	return false
}

func redactText(text, token string) string {
	if text == "" {
		return ""
	}

	markers := []string{"token=", "secret=", "password=", "private_key=", "session_key=", "auth=", "authorization="}
	lower := strings.ToLower(text)

	var builder strings.Builder
	builder.Grow(len(text))

	for i := 0; i < len(text); {
		markerLen := 0
		for _, marker := range markers {
			if strings.HasPrefix(lower[i:], marker) {
				markerLen = len(marker)
				break
			}
		}

		if markerLen == 0 {
			builder.WriteByte(text[i])
			i++
			continue
		}

		builder.WriteString(text[i : i+markerLen])
		i += markerLen

		for i < len(text) {
			ch := text[i]
			if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' || ch == ';' || ch == ',' {
				break
			}
			i++
		}
		builder.WriteString(token)
	}

	return builder.String()
}

func normalizeVoiceParticipantTileState(participant VoiceParticipantTileState) (VoiceParticipantTileState, error) {
	participant.ID = strings.TrimSpace(participant.ID)
	if participant.ID == "" {
		return VoiceParticipantTileState{}, ErrVoiceParticipantMissing
	}
	participant.Display = strings.TrimSpace(participant.Display)
	if participant.Display == "" {
		participant.Display = participant.ID
	}
	connectionStatus, err := normalizeVoiceConnectionStatus(participant.ConnectionStatus)
	if err != nil {
		return VoiceParticipantTileState{}, err
	}
	participant.ConnectionStatus = connectionStatus
	return participant, nil
}

func upsertVoiceParticipant(participants []VoiceParticipantTileState, participant VoiceParticipantTileState) []VoiceParticipantTileState {
	for i := range participants {
		if participants[i].ID == participant.ID {
			participants[i] = participant
			sortVoiceParticipants(participants)
			return participants
		}
	}
	participants = append(participants, participant)
	sortVoiceParticipants(participants)
	return participants
}

func removeVoiceParticipant(participants []VoiceParticipantTileState, participantID string) []VoiceParticipantTileState {
	for i := range participants {
		if participants[i].ID == participantID {
			return append(participants[:i], participants[i+1:]...)
		}
	}
	return participants
}

func sortVoiceParticipants(participants []VoiceParticipantTileState) {
	sort.Slice(participants, func(i, j int) bool {
		if participants[i].ID == participants[j].ID {
			return participants[i].Display < participants[j].Display
		}
		return participants[i].ID < participants[j].ID
	})
}

func (s *Shell) clearVoiceSession() {
	s.state.VoiceSession = VoiceSessionState{}
	controls := normalizeVoiceControlState(s.state.VoiceControls)
	controls.PushToTalkPressed = false
	controls.SelfMuted = false
	controls.SelfDeafened = false
	s.state.VoiceControls = controls
}

func (s *Shell) hasServer(serverID string) bool {
	for _, server := range s.state.Servers {
		if server.ID == serverID {
			return true
		}
	}
	return false
}

func (s *Shell) hasChannel(serverID, channelID string) bool {
	channels := s.state.ChannelsByServer[serverID]
	if len(channels) == 0 {
		return true
	}
	for _, channel := range channels {
		if channel.ID == channelID {
			return true
		}
	}
	return false
}

// Navigate transitions between shell routes while preserving relevant state.
func (s *Shell) Navigate(next Route) error {
	if s == nil {
		return ErrShellRequired
	}

	switch next {
	case RouteIdentitySetup:
		s.state.CurrentRoute = next
		return nil
	case RouteServerList:
		if !s.state.hasIdentity() {
			return ErrIdentitySetupRequired
		}
		s.state.CurrentRoute = next
		return nil
	case RouteChannelView:
		if !s.state.hasIdentity() {
			return ErrIdentitySetupRequired
		}
		if s.state.SelectedServerID == "" {
			return ErrServerSelectionRequired
		}
		if s.state.SelectedChannelID == "" {
			return ErrChannelSelectionMissing
		}
		s.state.CurrentRoute = next
		return nil
	case RouteSettings:
		s.state.CurrentRoute = next
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrRouteInvalid, next)
	}
}

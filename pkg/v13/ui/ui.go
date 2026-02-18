package ui

import (
	"fmt"

	"github.com/aether/code_aether/pkg/v13/chat"
	"github.com/aether/code_aether/pkg/v13/joinpolicy"
)

// Panel describes the join controls in the UI.
type Panel struct {
	SpaceName string
	Mode      joinpolicy.Mode
	Member    bool
}

// Label returns a deterministic text for the join panel state.
func (p Panel) Label() string {
	if p.Member {
		return fmt.Sprintf("%s • Joined (%s)", p.SpaceName, p.Mode)
	}
	if p.Mode == joinpolicy.ModeOpen {
		return fmt.Sprintf("%s • Open to join", p.SpaceName)
	}
	return fmt.Sprintf("%s • %s required", p.SpaceName, p.Mode)
}

// ComposerState tracks whether the composer has unsent text.
type ComposerState struct {
	ChannelID string
	Draft     string
}

// HasDraft reports whether there is pending text.
func (c ComposerState) HasDraft() bool {
	return c.Draft != ""
}

// DeliveryLabel returns the UI-friendly label for a message status.
func DeliveryLabel(msg chat.Message) string {
	return fmt.Sprintf("%s/%s/%s", msg.ChannelID, msg.Sender, msg.State)
}

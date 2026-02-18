package channels

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aether/code_aether/pkg/v13/spaces"
)

var (
	ErrChannelIDRequired  = errors.New("channel id is required")
	ErrSpaceIDRequired    = errors.New("space id is required for channel")
	ErrChannelNameMissing = errors.New("channel name is required")
)

// Channel models a deterministic text channel bound to a space.
type Channel struct {
	ID       string
	SpaceID  string
	Name     string
	Archived bool
	Members  []string
}

// NewChannel constructs a channel tied to a space.
func NewChannel(id, spaceID, name string) (Channel, error) {
	if strings.TrimSpace(id) == "" {
		return Channel{}, ErrChannelIDRequired
	}
	if strings.TrimSpace(spaceID) == "" {
		return Channel{}, ErrSpaceIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return Channel{}, ErrChannelNameMissing
	}
	return Channel{ID: strings.TrimSpace(id), SpaceID: strings.TrimSpace(spaceID), Name: strings.TrimSpace(name)}, nil
}

// AddMember registers a space member to the channel.
func (c *Channel) AddMember(user string) error {
	user = strings.TrimSpace(user)
	if user == "" {
		return ErrChannelNameMissing
	}
	for _, member := range c.Members {
		if member == user {
			return nil
		}
	}
	c.Members = append(c.Members, user)
	return nil
}

// BelongsToSpace confirms the channel-space binding.
func (c Channel) BelongsToSpace(spaceID string) bool {
	return strings.TrimSpace(c.SpaceID) == strings.TrimSpace(spaceID)
}

// Validate ensures channel state matches its space context.
func (c Channel) Validate(sp spaces.Space) error {
	if !c.BelongsToSpace(sp.ID) {
		return fmt.Errorf("channel space mismatch: %w", ErrSpaceIDRequired)
	}
	for _, member := range c.Members {
		if !sp.IsMember(member) {
			return fmt.Errorf("member %q missing from space", member)
		}
	}
	return nil
}

// Archive marks the channel as read-only.
func (c *Channel) Archive() {
	c.Archived = true
}

package spaces

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aether/code_aether/pkg/v13/joinpolicy"
)

var (
	ErrSpaceIDRequired = errors.New("space id is required")
	ErrFounderRequired = errors.New("founder is required")
	ErrMemberRequired  = errors.New("member id is required")
	ErrNotMember       = errors.New("user is not a member")
)

// Space captures deterministic state for a v13 workgroup.
type Space struct {
	ID      string
	Name    string
	Founder string
	Admins  []string
	Members []string
	Policy  joinpolicy.Mode
	Visible bool
}

// NewSpace constructs a baseline Space with founder auto-admin.
func NewSpace(id, name, founder string, policy joinpolicy.Mode) (Space, error) {
	if strings.TrimSpace(id) == "" {
		return Space{}, ErrSpaceIDRequired
	}
	if strings.TrimSpace(founder) == "" {
		return Space{}, ErrFounderRequired
	}
	if policy == "" {
		policy = joinpolicy.Default()
	}
	space := Space{
		ID:      strings.TrimSpace(id),
		Name:    strings.TrimSpace(name),
		Founder: strings.TrimSpace(founder),
		Policy:  policy,
		Visible: true,
	}
	space.Admins = append(space.Admins, space.Founder)
	space.Members = append(space.Members, space.Founder)
	return space, nil
}

// AddMember registers a user on the membership roster.
func (s *Space) AddMember(user string) error {
	if strings.TrimSpace(user) == "" {
		return ErrMemberRequired
	}
	user = strings.TrimSpace(user)
	if !s.IsMember(user) {
		s.Members = append(s.Members, user)
	}
	return nil
}

// PromoteToAdmin gives an existing member admin rights.
func (s *Space) PromoteToAdmin(user string) error {
	if strings.TrimSpace(user) == "" {
		return ErrMemberRequired
	}
	if !s.IsMember(user) {
		return ErrNotMember
	}
	user = strings.TrimSpace(user)
	for _, admin := range s.Admins {
		if admin == user {
			return nil
		}
	}
	s.Admins = append(s.Admins, user)
	return nil
}

// TransferFounder reassigns founder ownership to another admin.
func (s *Space) TransferFounder(target string) error {
	if strings.TrimSpace(target) == "" {
		return ErrFounderRequired
	}
	if !s.IsMember(target) {
		return ErrNotMember
	}
	s.Founder = target
	return s.PromoteToAdmin(target)
}

// IsMember checks membership deterministically.
func (s Space) IsMember(user string) bool {
	user = strings.TrimSpace(user)
	if user == "" {
		return false
	}
	for _, member := range s.Members {
		if member == user {
			return true
		}
	}
	return false
}

// Validate ensures the space contract stays coherent.
func (s Space) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return ErrSpaceIDRequired
	}
	if strings.TrimSpace(s.Founder) == "" {
		return ErrFounderRequired
	}
	if !s.IsMember(s.Founder) {
		return fmt.Errorf("%w: founder %q not a member", ErrNotMember, s.Founder)
	}
	return nil
}

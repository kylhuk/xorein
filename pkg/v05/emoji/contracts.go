package emoji

import "strings"

const MaxEmojiQuota = 50

func NormalizeShortcode(input string) string {
	trimmed := strings.TrimSpace(input)
	trimmed = strings.TrimPrefix(trimmed, ":")
	trimmed = strings.TrimSuffix(trimmed, ":")
	return strings.ToLower(trimmed)
}

func ValidShortcode(input string) bool {
	normalized := NormalizeShortcode(input)
	if normalized == "" || len(normalized) > 32 {
		return false
	}
	for _, r := range normalized {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

func CanAddReaction(currentCount int) bool {
	return currentCount < MaxEmojiQuota
}

type ReactionState string

const (
	ReactionAdded   ReactionState = "reaction.added"
	ReactionRemoved ReactionState = "reaction.removed"
	ReactionPending ReactionState = "reaction.pending"
)

type ReactionAction string

const (
	ReactionActionAdd    ReactionAction = "add"
	ReactionActionRemove ReactionAction = "remove"
)

func TransitionState(current ReactionState, action ReactionAction) ReactionState {
	switch action {
	case ReactionActionAdd:
		if current == ReactionRemoved {
			return ReactionPending
		}
		return ReactionAdded
	case ReactionActionRemove:
		return ReactionRemoved
	default:
		return current
	}
}

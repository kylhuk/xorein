package storage

// Canonical bucket names per docs/spec/v0.1/70-storage-and-key-derivation.md §6.3.
const (
	BucketIdentity      = "identity"
	BucketPeers         = "peers"
	BucketServers       = "servers"
	BucketChannels      = "channels"
	BucketMessages      = "messages"
	BucketDMs           = "dms"
	BucketFriends       = "friends"
	BucketRelayQueue    = "relay_queue"
	BucketVoiceSessions = "voice_sessions"
	BucketNotifications = "notifications"
	BucketSettings      = "settings"
)

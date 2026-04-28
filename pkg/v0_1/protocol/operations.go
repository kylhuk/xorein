package protocol

// Operation names and their required capabilities per spec 03 §6.
// Any operation not in this table is rejected with CodeUnsupportedOperation.
var OperationRequiredCaps = map[string][]FeatureFlag{
	// peer family
	"peer.info":              {"cap.peer.transport"},
	"peer.exchange":          {"cap.peer.transport"},
	"peer.bootstrap.register": {"cap.peer.bootstrap"},
	"peer.bootstrap.peers":   {"cap.peer.bootstrap"},
	"peer.manifest.publish":  {"cap.peer.manifest"},
	"peer.manifest.fetch":    {"cap.peer.manifest"},
	"peer.join":              {"cap.peer.join"},
	"peer.deliver":           {"cap.peer.delivery"},
	"peer.relay.store":       {"cap.peer.relay"},
	"peer.relay.drain":       {"cap.peer.relay"},

	// identity family
	"identity.publish_bundle": {"cap.identity"},
	"identity.fetch_bundle":   {"cap.identity"},

	// dm family (Seal mode)
	"dm.send":         {"cap.dm", "mode.seal"},
	"dm.session_init": {"cap.dm", "mode.seal"},

	// chat family
	"chat.send":    {"cap.chat"},
	"chat.join":    {"cap.chat"},
	"chat.history": {"cap.chat"},

	// manifest family
	"manifest.publish": {"cap.manifest"},
	"manifest.fetch":   {"cap.manifest"},

	// friends family
	"friends.request": {"cap.friends"},
	"friends.accept":  {"cap.friends"},
	"friends.remove":  {"cap.friends"},
	"friends.block":   {"cap.friends"},
	"friends.list":    {"cap.friends"},

	// presence family
	"presence.update": {"cap.presence"},
	"presence.query":  {"cap.presence"},

	// notify family
	"notify.push":  {"cap.notify"},
	"notify.drain": {"cap.notify"},

	// groupdm family (Tree mode)
	"groupdm.create": {"cap.group-dm", "mode.tree"},
	"groupdm.send":   {"cap.group-dm", "mode.tree"},
	"groupdm.add":    {"cap.group-dm", "mode.tree"},
	"groupdm.remove": {"cap.group-dm", "mode.tree"},

	// sync family — cap.sync universal; cap.archivist only required for push
	"sync.coverage": {"cap.sync"},
	"sync.fetch":    {"cap.sync"},
	"sync.push":     {"cap.sync", "cap.archivist"},

	// voice family (MediaShield mode)
	"voice.join":          {"cap.voice", "mode.mediashield"},
	"voice.leave":         {"cap.voice"},
	"voice.mute":          {"cap.voice"},
	"voice.offer":         {"cap.voice"},
	"voice.answer":        {"cap.voice"},
	"voice.ice":           {"cap.voice"},
	"voice.ice_complete":  {"cap.voice"},
	"voice.restart":       {"cap.voice"},
	"voice.terminate":     {"cap.voice"},
	"voice.frame":         {"cap.voice", "mode.mediashield"},

	// moderation family (spec 50)
	"moderation.kick":           {"cap.moderation"},
	"moderation.ban":            {"cap.moderation"},
	"moderation.unban":          {"cap.moderation"},
	"moderation.mute":           {"cap.moderation", "cap.slow-mode"},
	"moderation.slow_mode":      {"cap.moderation", "cap.slow-mode"},
	"moderation.delete_message": {"cap.moderation"},

	// governance family (spec 51) — cap.rbac per spec §2
	"governance.assign_role": {"cap.rbac"},
	"governance.revoke_role": {"cap.rbac"},
	"governance.create_role": {"cap.rbac"},
	"governance.delete_role": {"cap.rbac"},
	"governance.sync":        {"cap.rbac"},
}

// RequiredCapsFor returns the required capabilities for an operation.
// Returns nil and false if the operation is unknown.
func RequiredCapsFor(op string) ([]FeatureFlag, bool) {
	caps, ok := OperationRequiredCaps[op]
	return caps, ok
}

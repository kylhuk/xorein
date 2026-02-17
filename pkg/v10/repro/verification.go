package repro

// DeterministicBuildCommand returns the canonical build command string.
func DeterministicBuildCommand() string {
	return "go build ./... && go test ./... && pkg/v10/repro verify"
}

// BuildPins returns placeholder checksum pins for the release manifest.
func BuildPins() []string {
	return []string{
		"sha256:92b54a25d7ffbce4c4261130ff9769a7dd394284751a6ab4ccc2344078f30d26",
		"sha256:864f7f323a4526e799a6bfe38f912243ad6d620cf0690521f641bf0bf1ad94ce",
	}
}

// VerificationSteps outlines independent rebuild verification steps.
func VerificationSteps() []string {
	return []string{"checkout v1.0 tag", "run deterministic build command", "compare checksums"}
}

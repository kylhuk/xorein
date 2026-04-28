package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

// MajorVersion extracts the major component from a protocol ID like "/aether/chat/0.1.0".
// The version string is expected to be the last slash-delimited segment in the form
// "MAJOR.MINOR.PATCH". Returns an error if the ID is malformed or the major component
// is not a non-negative integer.
func MajorVersion(protocolID string) (int, error) {
	if protocolID == "" {
		return 0, fmt.Errorf("protocol: empty protocol ID")
	}
	parts := strings.Split(protocolID, "/")
	// Last non-empty segment is the version string.
	ver := ""
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			ver = parts[i]
			break
		}
	}
	if ver == "" {
		return 0, fmt.Errorf("protocol: cannot determine version from %q", protocolID)
	}
	vparts := strings.Split(ver, ".")
	if len(vparts) < 1 {
		return 0, fmt.Errorf("protocol: malformed version segment %q in %q", ver, protocolID)
	}
	major, err := strconv.Atoi(vparts[0])
	if err != nil {
		return 0, fmt.Errorf("protocol: non-integer major version in %q: %w", protocolID, err)
	}
	if major < 0 {
		return 0, fmt.Errorf("protocol: negative major version in %q", protocolID)
	}
	return major, nil
}

// RejectMajorDowngrade returns an error if remoteID has a strictly lower major version
// than localID. Per spec 03 §5, a major version downgrade MUST be rejected to prevent
// negotiation attacks that strip security properties.
func RejectMajorDowngrade(localID, remoteID string) error {
	localMajor, err := MajorVersion(localID)
	if err != nil {
		return fmt.Errorf("protocol: local ID parse error: %w", err)
	}
	remoteMajor, err := MajorVersion(remoteID)
	if err != nil {
		return fmt.Errorf("protocol: remote ID parse error: %w", err)
	}
	if remoteMajor < localMajor {
		return &PeerStreamError{
			Code:               CodeUnsupportedVersion,
			Message:            fmt.Sprintf("major version downgrade rejected: local=%d remote=%d", localMajor, remoteMajor),
			UnsupportedVersion: remoteID,
		}
	}
	return nil
}

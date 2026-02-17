package security

// AssetInventory returns deterministic audit assets.
func AssetInventory() map[string]string {
	return map[string]string{
		"E2EE":           "TREASURE",
		"P2P networking": "MeshSession",
		"Key management": "X3DH+MLS",
	}
}

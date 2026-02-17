package network

// Topology maps node IDs to regions for the bootstrap topology.
func Topology(nodes []string) map[string]string {
	regions := []string{"us-east", "us-west", "eu-central", "ap-south"}
	result := make(map[string]string)
	for idx, node := range nodes {
		result[node] = regions[idx%len(regions)]
	}
	return result
}

// ContinuityPlan records deterministic successor nodes for each region.
func ContinuityPlan() map[string]string {
	return map[string]string{
		"us-east":    "bootstrap-node-a",
		"us-west":    "bootstrap-node-b",
		"eu-central": "bootstrap-node-c",
		"ap-south":   "bootstrap-node-d",
	}
}

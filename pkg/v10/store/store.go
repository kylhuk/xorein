package store

import "strconv"

// BootstrapNodes returns deterministic node identifiers for the requested count.
func BootstrapNodes(count int) []string {
	if count <= 0 {
		count = 10
	}
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = "bootstrap-node-" + strconv.Itoa(i+1)
	}
	return result
}

// NodeLabels returns anchor labels for each bootstrap node.
func NodeLabels() map[string]string {
	labels := make(map[string]string)
	for _, node := range BootstrapNodes(10) {
		labels[node] = "VA-N1"
	}
	return labels
}

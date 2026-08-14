package router

import (
	"fmt"
	"testing"
)

func TestRouter(t *testing.T) {
	nodes := []string{
		"localhost:8001",
		"localhost:8002",
		"localhost:8003",
	}

	r := NewRouter(nodes)

	keys := []string{
		"apple",
		"banana",
		"cat",
		"dog",
		"hello",
	}

	for _, key := range keys {
		node := r.GetNode(key)

		if node == "" {
			t.Errorf("no node found for %s", key)
		}

		t.Logf("%s → %s", key, node)
	}
}

func TestKeyDistribution(t *testing.T) {
	nodes := []string{
		"localhost:8001",
		"localhost:8002",
		"localhost:8003",
	}

	r := NewRouter(nodes)

	counts := make(map[string]int)

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)

		node := r.GetNode(key)
		counts[node]++
	}

	for node, count := range counts {
		t.Logf("%s → %d keys", node, count)
	}
}

func TestNodeAdditionMovement(t *testing.T) {
	r := NewRouter([]string{
		"node-A",
		"node-B",
		"node-C",
	})

	keys := make([]string, 1000)

	for i := range keys {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	before := make(map[string]string)

	for _, key := range keys {
		before[key] = r.GetNode(key)
	}

	// Add a fourth node.
	r.ring.AddNode("node-D")

	moved := 0

	for _, key := range keys {
		after := r.GetNode(key)

		if before[key] != after {
			moved++
		}
	}

	t.Logf("keys moved: %d / %d", moved, len(keys))
}

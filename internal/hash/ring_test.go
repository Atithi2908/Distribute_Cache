package hash

import "testing"

func TestHashRing(t *testing.T) {
	ring := NewHashRing(100)

	ring.AddNode("node-A")
	ring.AddNode("node-B")
	ring.AddNode("node-C")

	tests := []string{
		"apple",
		"banana",
		"cat",
		"dog",
		"hello",
	}

	for _, key := range tests {
		node := ring.GetNode(key)

		if node == "" {
			t.Errorf("no node found for key %s", key)
		}

		t.Logf("%s → %s", key, node)
	}
}

func TestHashRingRemoveNode(t *testing.T) {
	ring := NewHashRing(100)

	ring.AddNode("node-A")
	ring.AddNode("node-B")
	ring.AddNode("node-C")

	ring.RemoveNode("node-B")

	for _, key := range []string{"apple", "banana", "cat", "dog"} {
		node := ring.GetNode(key)

		if node == "node-B" {
			t.Errorf("removed node was returned for key %s", key)
		}

		t.Logf("%s → %s", key, node)
	}
}

func TestNodeAddition(t *testing.T) {
	ring := NewHashRing(100)

	ring.AddNode("node-A")
	ring.AddNode("node-B")
	ring.AddNode("node-C")

	keys := []string{
		"apple",
		"banana",
		"cat",
		"dog",
		"elephant",
		"fish",
		"grape",
		"hello",
		"india",
		"jack",
	}

	before := make(map[string]string)

	for _, key := range keys {
		before[key] = ring.GetNode(key)
	}

	ring.AddNode("node-D")

	for _, key := range keys {
		after := ring.GetNode(key)

		if before[key] != after {
			t.Logf("%s moved: %s → %s", key, before[key], after)
		}
	}
}

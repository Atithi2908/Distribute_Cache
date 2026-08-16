package hash

import (
	"fmt"
	"hash/crc32"
	"sort"
)

type HashRing struct {
	nodes        map[uint32]string
	hashKeys     []uint32
	virtualNodes int
}

func NewHashRing(virtualNodes int) *HashRing {
	return &HashRing{
		nodes:        make(map[uint32]string),
		virtualNodes: virtualNodes,
	}
}

func hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (h *HashRing) AddNode(node string) {
	for i := 0; i < h.virtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s#%d", node, i)

		hash := hashKey(virtualNode)

		h.nodes[hash] = node
		h.hashKeys = append(h.hashKeys, hash)
	}

	sort.Slice(h.hashKeys, func(i, j int) bool {
		return h.hashKeys[i] < h.hashKeys[j]
	})
}

func (h *HashRing) GetNodes(key string, count int) []string {
	if len(h.hashKeys) == 0 || count <= 0 {
		return nil
	}

	if count > len(h.nodes) {
		count = len(h.nodes)
	}

	hash := hashKey(key)

	index := sort.Search(len(h.hashKeys), func(i int) bool {
		return h.hashKeys[i] >= hash
	})

	if index == len(h.hashKeys) {
		index = 0
	}

	result := make([]string, 0, count)
	seen := make(map[string]bool)

	for len(result) < count {
		node := h.nodes[h.hashKeys[index]]

		if !seen[node] {
			result = append(result, node)
			seen[node] = true
		}

		index = (index + 1) % len(h.hashKeys)
	}

	return result
}

func (h *HashRing) GetNode(key string) string {
	nodes := h.GetNodes(key, 1)

	if len(nodes) == 0 {
		return ""
	}

	return nodes[0]
}

func (h *HashRing) RemoveNode(node string) {
	for i := 0; i < h.virtualNodes; i++ {
		virtualNode := fmt.Sprintf("%s#%d", node, i)

		hash := hashKey(virtualNode)

		delete(h.nodes, hash)

		for j, key := range h.hashKeys {
			if key == hash {
				h.hashKeys = append(h.hashKeys[:j], h.hashKeys[j+1:]...)
				break
			}
		}
	}
}

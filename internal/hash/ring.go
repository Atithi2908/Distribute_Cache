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

func (h *HashRing) GetNode(key string) string {
	if len(h.hashKeys) == 0 {
		return ""
	}

	hash := hashKey(key)

	index := sort.Search(len(h.hashKeys), func(i int) bool {
		return h.hashKeys[i] >= hash
	})

	if index == len(h.hashKeys) {
		index = 0
	}

	return h.nodes[h.hashKeys[index]]
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

package cache

import "sync"

type MemoryCache struct {
	store map[string][]byte
	mu    sync.RWMutex
}

var _ Cache = (*MemoryCache)(nil)

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string][]byte),
	}
}

func (c *MemoryCache) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
	return nil
}

func (c *MemoryCache) Get(key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.store[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}

func (c *MemoryCache) All() map[string][]byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]byte)

	for key, value := range c.store {
		result[key] = append([]byte(nil), value...)
	}

	return result
}

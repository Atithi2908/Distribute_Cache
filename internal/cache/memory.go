package cache

type MemoryCache struct {
	store map[string][]byte
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string][]byte),
	}
}

func (c *MemoryCache) Set(key string, value []byte) error {
	c.store[key] = value
	return nil
}

func (c *MemoryCache) Get(key string) ([]byte, error) {
	val, ok := c.store[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (c *MemoryCache) Delete(key string) error {
	delete(c.store, key)
	return nil
}

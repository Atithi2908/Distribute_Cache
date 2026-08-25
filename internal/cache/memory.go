package cache

import (
	"container/list"
	"sync"
	"time"
)

type cacheItem struct {
	key        string
	value      []byte
	expiration time.Time // zero value means no expiration
}

type MemoryCache struct {
	store    map[string]*list.Element
	lruList  *list.List
	capacity int // 0 means no capacity limit
	mu       sync.RWMutex

	cleanupStop chan struct{}
	cleanupWg   sync.WaitGroup
}

var _ Cache = (*MemoryCache)(nil)

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store:   make(map[string]*list.Element),
		lruList: list.New(),
	}
}

func NewMemoryCacheWithCapacity(capacity int) *MemoryCache {
	return &MemoryCache{
		store:    make(map[string]*list.Element),
		lruList:  list.New(),
		capacity: capacity,
	}
}

func (c *MemoryCache) SetCapacity(capacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.capacity = capacity
	c.evict()
}

func (c *MemoryCache) Set(key string, value []byte) error {
	return c.SetWithTTL(key, value, 0)
}

func (c *MemoryCache) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	valCopy := append([]byte(nil), value...)

	if elem, ok := c.store[key]; ok {
		item := elem.Value.(*cacheItem)
		item.value = valCopy
		item.expiration = exp
		c.lruList.MoveToFront(elem)
		return nil
	}

	item := &cacheItem{
		key:        key,
		value:      valCopy,
		expiration: exp,
	}
	elem := c.lruList.PushFront(item)
	c.store[key] = elem

	c.evict()

	return nil
}

func (c *MemoryCache) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.store[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	item := elem.Value.(*cacheItem)

	// Lazy expiration check
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		c.lruList.Remove(elem)
		delete(c.store, key)
		return nil, ErrKeyNotFound
	}

	c.lruList.MoveToFront(elem)
	valCopy := append([]byte(nil), item.value...)
	return valCopy, nil
}

func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.store[key]; ok {
		c.lruList.Remove(elem)
		delete(c.store, key)
	}
	return nil
}

func (c *MemoryCache) All() map[string][]byte {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	result := make(map[string][]byte)

	for key, elem := range c.store {
		item := elem.Value.(*cacheItem)
		if !item.expiration.IsZero() && now.After(item.expiration) {
			c.lruList.Remove(elem)
			delete(c.store, key)
			continue
		}
		result[key] = append([]byte(nil), item.value...)
	}

	return result
}

func (c *MemoryCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

func (c *MemoryCache) evict() {
	if c.capacity <= 0 {
		return
	}

	for len(c.store) > c.capacity {
		backElem := c.lruList.Back()
		if backElem == nil {
			break
		}
		backItem := backElem.Value.(*cacheItem)
		c.lruList.Remove(backElem)
		delete(c.store, backItem.key)
	}
}

func (c *MemoryCache) StartCleanup(interval time.Duration) {
	c.mu.Lock()
	if c.cleanupStop != nil {
		c.mu.Unlock()
		return
	}
	stopChan := make(chan struct{})
	c.cleanupStop = stopChan
	c.cleanupWg.Add(1)
	c.mu.Unlock()

	go func() {
		defer c.cleanupWg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.DeleteExpired()
			case <-stopChan:
				return
			}
		}
	}()
}

func (c *MemoryCache) StopCleanup() {
	c.mu.Lock()
	if c.cleanupStop == nil {
		c.mu.Unlock()
		return
	}
	close(c.cleanupStop)
	c.cleanupStop = nil
	c.mu.Unlock()

	c.cleanupWg.Wait()
}

func (c *MemoryCache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, elem := range c.store {
		item := elem.Value.(*cacheItem)
		if !item.expiration.IsZero() && now.After(item.expiration) {
			c.lruList.Remove(elem)
			delete(c.store, key)
		}
	}
}

package cache

import "time"

type Cache interface {
	Set(key string, value []byte) error
	SetWithTTL(key string, value []byte, ttl time.Duration) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}

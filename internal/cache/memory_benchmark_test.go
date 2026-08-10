package cache

import "testing"

func BenchmarkMemoryCacheGet(b *testing.B) {
	cache := NewMemoryCache()

	cache.Set("name", []byte("Atithi"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Get("name")
	}
}
func BenchmarkMemoryCacheSet(b *testing.B) {
	cache := NewMemoryCache()

	value := []byte("Atithi")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set("name", value)
	}
}

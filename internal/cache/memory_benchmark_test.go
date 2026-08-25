package cache

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkMemoryCacheGet(b *testing.B) {
	c := NewMemoryCache()
	c.Set("name", []byte("Atithi"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Get("name")
	}
}

func BenchmarkMemoryCacheSet(b *testing.B) {
	c := NewMemoryCache()
	val := []byte("Atithi")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Set("name", val)
	}
}

func BenchmarkMemoryCacheDelete(b *testing.B) {
	c := NewMemoryCache()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Delete("name")
	}
}

func BenchmarkMemoryCacheGetWithTTL(b *testing.B) {
	c := NewMemoryCache()
	c.SetWithTTL("name", []byte("Atithi"), 10*time.Hour)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Get("name")
	}
}

func BenchmarkMemoryCacheSetWithTTL(b *testing.B) {
	c := NewMemoryCache()
	val := []byte("Atithi")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.SetWithTTL("name", val, 10*time.Hour)
	}
}

func BenchmarkMemoryCacheLRUEviction(b *testing.B) {
	c := NewMemoryCacheWithCapacity(100)
	val := []byte("Atithi")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%200)
		c.Set(key, val)
	}
}

func BenchmarkMemoryCacheParallelGet(b *testing.B) {
	c := NewMemoryCache()
	c.Set("name", []byte("Atithi"))

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Get("name")
		}
	})
}

func BenchmarkMemoryCacheParallelSet(b *testing.B) {
	c := NewMemoryCache()
	val := []byte("Atithi")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			c.Set(key, val)
			i++
		}
	})
}

func BenchmarkMemoryCacheParallelMixed(b *testing.B) {
	c := NewMemoryCacheWithCapacity(1000)
	c.Set("name", []byte("Atithi"))
	val := []byte("Atithi")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%5 == 0 {
				key := fmt.Sprintf("key-%d", i%500)
				c.SetWithTTL(key, val, 5*time.Minute)
			} else {
				c.Get("name")
			}
			i++
		}
	})
}

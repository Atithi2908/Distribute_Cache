package cache

import "testing"

func TestMemoryCacheSetGet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple value", "name", "Atithi"},
		{"another value", "city", "Dharwad"},
		{"empty value", "empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewMemoryCache()

			err := cache.Set(tt.key, []byte(tt.value))
			if err != nil {
				t.Fatal(err)
			}

			value, err := cache.Get(tt.key)
			if err != nil {
				t.Fatal(err)
			}

			if string(value) != tt.value {
				t.Fatalf("expected %q, got %q", tt.value, string(value))
			}
		})
	}
}

func TestMemoryCacheMissingKey(t *testing.T) {
	cache := NewMemoryCache()

	_, err := cache.Get("does-not-exist")

	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestMemoryCacheDelete(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set("name", []byte("Atithi"))
	if err != nil {
		t.Fatal(err)
	}

	err = cache.Delete("name")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.Get("name")

	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestMemoryCacheOverwrite(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set("name", []byte("Atithi"))
	if err != nil {
		t.Fatal(err)
	}

	err = cache.Set("name", []byte("Rahul"))
	if err != nil {
		t.Fatal(err)
	}

	value, err := cache.Get("name")
	if err != nil {
		t.Fatal(err)
	}

	if string(value) != "Rahul" {
		t.Fatalf("expected Rahul, got %s", value)
	}
}

func TestMemoryCacheDeleteMissingKey(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Delete("does-not-exist")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

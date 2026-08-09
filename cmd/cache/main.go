package main

import (
	"fmt"

	"distributed-cache/internal/cache"
)

func main() {
	c := cache.NewMemoryCache()

	err := c.Set("name", []byte("Atithi"))
	if err != nil {
		fmt.Println(err)
		return
	}

	value, err := c.Get("name")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(value))

	err = c.Delete("name")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = c.Get("name")
	if err == cache.ErrKeyNotFound {
		fmt.Println("key deleted successfully")
	}
}

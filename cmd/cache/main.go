package main

import (
	"log"

	"distributed-cache/internal/cache"
	"distributed-cache/internal/server"
)

func main() {
	c := cache.NewMemoryCache()

	if err := server.Start(c); err != nil {
		log.Fatal(err)
	}
}

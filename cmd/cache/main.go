package main

import (
	"log"
	"os"

	"distributed-cache/internal/cache"
	"distributed-cache/internal/server"
)

func main() {
	address := ":8000"

	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	c := cache.NewMemoryCache()

	if err := server.Start(c, address); err != nil {
		log.Fatal(err)
	}
}

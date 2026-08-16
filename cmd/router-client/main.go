package main

import (
	"fmt"
	"log"

	"distributed-cache/internal/router"
)

func main() {
	nodes := []string{
		"localhost:8001",
		"localhost:8002",
		"localhost:8003",
	}

	r := router.NewRouter(nodes)

	// Write data
	err := r.Set("user123", []byte("Atithi"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("SET user123 = Atithi successful")

	// Read data
	value, err := r.Get("user123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("GET user123:", value)

	// Another key
	err = r.Set("city", []byte("Delhi"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("SET city = Delhi successful")

	value, err = r.Get("city")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("GET city:", value)
}

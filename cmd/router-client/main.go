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

	response, err := r.Send("SET apple red", "apple")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("SET:", response)

	response, err = r.Send("GET apple", "apple")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("GET:", response)
}

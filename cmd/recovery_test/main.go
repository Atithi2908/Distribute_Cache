package main

import (
	"fmt"
	"time"

	"distributed-cache/internal/router"
)

func main() {
	nodes := []string{
		"localhost:8001",
		"localhost:8002",
		"localhost:8003",
	}

	r := router.NewRouter(nodes)

	fmt.Println("Checking node health...")

	r.HealthCheck()

	fmt.Println("Health check completed.")

	time.Sleep(1 * time.Second)

	fmt.Println("Checking again...")

	r.HealthCheck()

	fmt.Println("Recovery check completed.")
}

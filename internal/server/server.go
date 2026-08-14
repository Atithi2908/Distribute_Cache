package server

import (
	"bufio"
	"fmt"
	"net"

	"distributed-cache/internal/cache"
)

func Start(c *cache.MemoryCache, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	defer listener.Close()

	fmt.Println("Server listening on", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go handleConnection(conn, c)
	}
}

func handleConnection(conn net.Conn, c *cache.MemoryCache) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		req, err := ParseRequest(scanner.Text())

		if err != nil {
			fmt.Fprintln(conn, "ERROR", err)
			continue
		}

		handleRequest(conn, c, req)
	}
}

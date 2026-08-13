package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}

	defer conn.Close()

	reader := bufio.NewReader(conn)

	commands := []string{
		"SET name Atithi",
		"GET name",
		"DELETE name",
		"GET name",
	}

	for _, command := range commands {

		_, err := fmt.Fprintf(conn, "%s\n", command)
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}

		fmt.Printf("%s → %s", command, response)
	}
}

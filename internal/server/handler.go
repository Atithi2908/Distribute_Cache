package server

import (
	"fmt"
	"net"

	"distributed-cache/internal/cache"
)

func handleRequest(conn net.Conn, c *cache.MemoryCache, req Request) {
	var response Response

	switch req.Command {
	case "SET":
		err := c.Set(req.Key, []byte(req.Value))
		if err != nil {
			response = Response{Status: "ERROR"}
			break
		}

		response = Response{Status: "OK"}

	case "GET":
		value, err := c.Get(req.Key)
		if err != nil {
			response = Response{Status: "NOT_FOUND"}
			break
		}

		response = Response{
			Status: "OK",
			Value:  string(value),
		}

	case "DELETE":
		err := c.Delete(req.Key)
		if err != nil {
			response = Response{Status: "ERROR"}
			break
		}

		response = Response{Status: "OK"}
	}

	fmt.Fprintln(conn, response.Status, response.Value)
}

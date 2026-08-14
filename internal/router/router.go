package router

import (
	"bufio"
	"fmt"
	"net"

	"distributed-cache/internal/hash"
)

type Router struct {
	ring *hash.HashRing
}

func NewRouter(nodes []string) *Router {
	ring := hash.NewHashRing(100)

	for _, node := range nodes {
		ring.AddNode(node)
	}

	return &Router{
		ring: ring,
	}
}

func (r *Router) GetNode(key string) string {
	return r.ring.GetNode(key)
}

func (r *Router) Send(command string, key string) (string, error) {
	node := r.GetNode(key)

	if node == "" {
		return "", fmt.Errorf("no cache node available")
	}

	conn, err := net.Dial("tcp", node)
	if err != nil {
		return "", err
	}

	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%s\n", command)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return response, nil
}

func (r *Router) Set(key string, value []byte) (string, error) {
	return r.Send(
		fmt.Sprintf("SET %s %s", key, string(value)),
		key,
	)
}

func (r *Router) Get(key string) (string, error) {
	return r.Send("GET "+key, key)
}

func (r *Router) Delete(key string) (string, error) {
	return r.Send("DELETE "+key, key)
}

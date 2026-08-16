package router

import (
	"bufio"
	"distributed-cache/internal/hash"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Router struct {
	ring   *hash.HashRing
	nodes  []string
	health map[string]bool
	mu     sync.RWMutex
}

func NewRouter(nodes []string) *Router {
	ring := hash.NewHashRing(100)

	health := make(map[string]bool)

	for _, node := range nodes {
		ring.AddNode(node)
		health[node] = true
	}

	return &Router{
		ring:   ring,
		nodes:  nodes,
		health: health,
	}
}

func (r *Router) GetNode(key string) string {
	nodes := r.ring.GetNodes(key, len(r.nodes))

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, node := range nodes {
		if r.health[node] {
			return node
		}
	}

	return ""
}

func (r *Router) Send(command string, key string) (string, error) {
	node := r.GetNode(key)

	if node == "" {
		return "", fmt.Errorf("no healthy cache node available")
	}

	response, err := r.sendToNode(node, command)

	if err != nil {
		r.markUnhealthy(node)
		return "", err
	}

	return response, nil
}

func (r *Router) Set(key string, value []byte) error {
	nodes := r.ring.GetNodes(key, 3)

	var wg sync.WaitGroup
	successes := 0

	var mu sync.Mutex

	for _, node := range nodes {
		r.mu.RLock()
		healthy := r.health[node]
		r.mu.RUnlock()

		if !healthy {
			continue
		}

		wg.Add(1)

		go func(node string) {
			defer wg.Done()

			_, err := r.sendToNode(
				node,
				fmt.Sprintf("SET %s %s", key, string(value)),
			)

			if err != nil {
				r.markUnhealthy(node)
				return
			}

			mu.Lock()
			successes++
			mu.Unlock()
		}(node)
	}

	wg.Wait()

	const writeQuorum = 2

	if successes < writeQuorum {
		return fmt.Errorf(
			"write quorum not reached: %d/%d",
			successes,
			writeQuorum,
		)
	}

	return nil
}

func (r *Router) Get(key string) (string, error) {
	nodes := r.ring.GetNodes(key, 3)

	var wg sync.WaitGroup
	var mu sync.Mutex

	responses := make([]string, 0)
	successes := 0

	for _, node := range nodes {
		r.mu.RLock()
		healthy := r.health[node]
		r.mu.RUnlock()

		if !healthy {
			continue
		}

		wg.Add(1)

		go func(node string) {
			defer wg.Done()

			response, err := r.sendToNode(node, "GET "+key)

			if err != nil {
				r.markUnhealthy(node)
				return
			}

			mu.Lock()
			successes++
			responses = append(responses, response)
			mu.Unlock()
		}(node)
	}

	wg.Wait()

	const readQuorum = 2

	if successes < readQuorum {
		return "", fmt.Errorf(
			"read quorum not reached: %d/%d",
			successes,
			readQuorum,
		)
	}

	if len(responses) == 0 {
		return "", fmt.Errorf("key not found")
	}

	return responses[0], nil
}

func (r *Router) Delete(key string) (string, error) {
	return r.Send("DELETE "+key, key)
}

func (r *Router) sendToNode(node string, command string) (string, error) {
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

	return reader.ReadString('\n')
}

func (r *Router) markUnhealthy(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.health[node] = false

	fmt.Println("Node marked unhealthy:", node)
}

func (r *Router) HealthCheck() {
	for _, node := range r.nodes {
		conn, err := net.Dial("tcp", node)

		if err != nil {
			r.markUnhealthy(node)
			continue
		}

		conn.Close()

		r.mu.Lock()
		if !r.health[node] {
			fmt.Println("Node recovered:", node)
		}
		r.health[node] = true
		r.mu.Unlock()
	}
}

func (r *Router) RecoverNode(node string, sourceNodes []string) error {
	for _, source := range sourceNodes {
		if source == node {
			continue
		}

		conn, err := net.Dial("tcp", source)
		if err != nil {
			continue
		}

		conn.Close()

		return nil
	}

	return fmt.Errorf("no healthy source node available")
}

func (r *Router) DumpNode(node string) (map[string]string, error) {
	conn, err := net.Dial("tcp", node)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	_, err = fmt.Fprintln(conn, "DUMP")
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)

	data := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		if line == "END" {
			break
		}

		parts := strings.SplitN(line, " ", 2)

		if len(parts) == 2 {
			data[parts[0]] = parts[1]
		}
	}

	return data, nil
}

func (r *Router) RestoreNode(target string, data map[string]string) error {
	for key, value := range data {
		if !r.ShouldStore(key, target) {
			continue
		}

		_, err := r.sendToNode(
			target,
			fmt.Sprintf("SET %s %s", key, value),
		)

		if err != nil {
			return fmt.Errorf(
				"failed to restore key %s: %w",
				key,
				err,
			)
		}
	}

	return nil
}

func (r *Router) ShouldStore(key string, node string) bool {
	nodes := r.ring.GetNodes(key, 3)

	for _, n := range nodes {
		if n == node {
			return true
		}
	}

	return false
}

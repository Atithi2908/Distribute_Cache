package router

import (
	"bufio"
	"distributed-cache/internal/cache"
	"distributed-cache/internal/server"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestReplicationWithTTL(t *testing.T) {
	addrs := make([]string, 3)
	listeners := make([]net.Listener, 3)

	for i := 0; i < 3; i++ {
		c := cache.NewMemoryCache()

		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to listen: %v", err)
		}
		listeners[i] = l
		addrs[i] = l.Addr().String()

		go func(listener net.Listener, mc *cache.MemoryCache) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				go handleTestConn(conn, mc)
			}
		}(l, c)
	}

	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	r := NewRouter(addrs)
	defer r.Close()

	// Replicated SET with 1 second TTL
	err := r.SetWithTTL("user1", []byte("Atithi"), 1*time.Second)
	if err != nil {
		t.Fatalf("SetWithTTL failed: %v", err)
	}

	// Immediate GET
	resp, err := r.Get("user1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !strings.Contains(resp, "Atithi") {
		t.Fatalf("expected Atithi, got %s", resp)
	}

	// Sleep past TTL
	time.Sleep(1200 * time.Millisecond)

	// GET after expiration should fail quorum or return NOT_FOUND
	resp, err = r.Get("user1")
	if err == nil && strings.Contains(resp, "OK") && strings.Contains(resp, "Atithi") {
		t.Fatalf("expected key to expire, but got %s", resp)
	}
}

func handleTestConn(conn net.Conn, c *cache.MemoryCache) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		req, err := server.ParseRequest(scanner.Text())
		if err != nil {
			fmt.Fprintln(conn, "ERROR", err)
			continue
		}
		switch req.Command {
		case "SET":
			var setErr error
			if req.TTL > 0 {
				setErr = c.SetWithTTL(req.Key, []byte(req.Value), req.TTL)
			} else {
				setErr = c.Set(req.Key, []byte(req.Value))
			}
			if setErr != nil {
				fmt.Fprintln(conn, "ERROR")
			} else {
				fmt.Fprintln(conn, "OK")
			}
		case "GET":
			val, err := c.Get(req.Key)
			if err != nil {
				fmt.Fprintln(conn, "NOT_FOUND")
			} else {
				fmt.Fprintln(conn, "OK", string(val))
			}
		case "DELETE":
			c.Delete(req.Key)
			fmt.Fprintln(conn, "OK")
		}
	}
}

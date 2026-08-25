package router

import (
	"bufio"
	"distributed-cache/internal/cache"
	"distributed-cache/internal/server"
	"fmt"
	"net"
	"testing"
	"time"
)

func setupTestNodes(b *testing.B, count int) ([]string, []net.Listener, func()) {
	addrs := make([]string, count)
	listeners := make([]net.Listener, count)

	for i := 0; i < count; i++ {
		c := cache.NewMemoryCache()
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			b.Fatalf("failed to listen: %v", err)
		}
		listeners[i] = l
		addrs[i] = l.Addr().String()

		go func(listener net.Listener, mc *cache.MemoryCache) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				go func(c net.Conn) {
					defer c.Close()
					scanner := bufio.NewScanner(c)
					for scanner.Scan() {
						req, err := server.ParseRequest(scanner.Text())
						if err != nil {
							fmt.Fprintln(c, "ERROR")
							continue
						}
						switch req.Command {
						case "SET":
							if req.TTL > 0 {
								mc.SetWithTTL(req.Key, []byte(req.Value), req.TTL)
							} else {
								mc.Set(req.Key, []byte(req.Value))
							}
							fmt.Fprintln(c, "OK")
						case "GET":
							val, err := mc.Get(req.Key)
							if err != nil {
								fmt.Fprintln(c, "NOT_FOUND")
							} else {
								fmt.Fprintln(c, "OK", string(val))
							}
						}
					}
				}(conn)
			}
		}(l, c)
	}

	cleanup := func() {
		for _, l := range listeners {
			l.Close()
		}
	}

	return addrs, listeners, cleanup
}

func BenchmarkRouterSetReplicated(b *testing.B) {
	addrs, _, cleanup := setupTestNodes(b, 3)
	defer cleanup()

	r := NewRouter(addrs)
	defer r.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%d", i%100)
		_ = r.Set(key, []byte("bench-value"))
	}
}

func BenchmarkRouterGetQuorum(b *testing.B) {
	addrs, _, cleanup := setupTestNodes(b, 3)
	defer cleanup()

	r := NewRouter(addrs)
	defer r.Close()

	_ = r.Set("quorum-key", []byte("quorum-value"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Get("quorum-key")
	}
}

func BenchmarkRouterParallelReplicatedSet(b *testing.B) {
	addrs, _, cleanup := setupTestNodes(b, 3)
	defer cleanup()

	r := NewRouter(addrs)
	defer r.Close()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("pkey-%d", i%100)
			_ = r.SetWithTTL(key, []byte("pval"), 5*time.Minute)
			i++
		}
	})
}

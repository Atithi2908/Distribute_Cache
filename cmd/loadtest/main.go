package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"distributed-cache/internal/cache"
	"distributed-cache/internal/router"
	"distributed-cache/internal/server"
)

func main() {
	clientsFlag := flag.Int("clients", 50, "Number of concurrent clients")
	opsFlag := flag.Int("ops", 10000, "Total number of operations")
	readRatioFlag := flag.Float64("ratio", 0.8, "Read ratio (e.g. 0.8 for 80% GET, 20% SET)")
	ttlFlag := flag.Int("ttl", 0, "TTL in seconds for SET operations (0 = no TTL)")
	nodesFlag := flag.String("nodes", "", "Comma-separated cache node addresses (empty = auto-spawn 3 local nodes)")
	flag.Parse()

	var nodes []string
	var listeners []net.Listener

	if *nodesFlag == "" {
		fmt.Println("No nodes specified. Auto-spawning 3 local in-memory cache servers...")
		for i := 0; i < 3; i++ {
			l, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				fmt.Printf("Failed to listen: %v\n", err)
				return
			}
			listeners = append(listeners, l)
			nodes = append(nodes, l.Addr().String())

			c := cache.NewMemoryCache()
			go func(lis net.Listener, mc *cache.MemoryCache) {
				for {
					conn, err := lis.Accept()
					if err != nil {
						return
					}
					go handleConn(conn, mc)
				}
			}(l, c)
		}
		defer func() {
			for _, l := range listeners {
				l.Close()
			}
		}()
	} else {
		nodes = strings.Split(*nodesFlag, ",")
	}

	fmt.Printf("\n=== Distributed Cache Load Test ===\n")
	fmt.Printf("Nodes: %v\n", nodes)
	fmt.Printf("Concurrent Clients: %d\n", *clientsFlag)
	fmt.Printf("Total Operations: %d\n", *opsFlag)
	fmt.Printf("Read Ratio: %.0f%% GET, %.0f%% SET\n", *readRatioFlag*100, (1.0-*readRatioFlag)*100)
	fmt.Printf("SET TTL: %d seconds\n\n", *ttlFlag)

	r := router.NewRouter(nodes)
	defer r.Close()

	// Pre-populate some keys
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = r.Set(key, []byte(fmt.Sprintf("val-%d", i)))
	}

	var totalSuccess uint64
	var totalFail uint64

	latencies := make([]time.Duration, *opsFlag)
	var latencyIndex uint64

	opsPerClient := *opsFlag / *clientsFlag
	var wg sync.WaitGroup
	startTime := time.Now()

	for c := 0; c < *clientsFlag; c++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(clientID)))

			for i := 0; i < opsPerClient; i++ {
				key := fmt.Sprintf("key-%d", rng.Intn(200))
				opStart := time.Now()
				var err error

				if rng.Float64() < *readRatioFlag {
					// GET operation
					_, err = r.Get(key)
				} else {
					// SET operation
					val := fmt.Sprintf("value-%d", rng.Intn(10000))
					if *ttlFlag > 0 {
						err = r.SetWithTTL(key, []byte(val), time.Duration(*ttlFlag)*time.Second)
					} else {
						err = r.Set(key, []byte(val))
					}
				}

				elapsed := time.Since(opStart)
				idx := atomic.AddUint64(&latencyIndex, 1) - 1
				if idx < uint64(len(latencies)) {
					latencies[idx] = elapsed
				}

				if err == nil {
					atomic.AddUint64(&totalSuccess, 1)
				} else {
					atomic.AddUint64(&totalFail, 1)
				}
			}
		}(c)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	actualOps := atomic.LoadUint64(&latencyIndex)
	validLatencies := latencies[:actualOps]
	sort.Slice(validLatencies, func(i, j int) bool {
		return validLatencies[i] < validLatencies[j]
	})

	var totalLatency time.Duration
	for _, l := range validLatencies {
		totalLatency += l
	}

	avgLatency := time.Duration(0)
	p50 := time.Duration(0)
	p95 := time.Duration(0)
	p99 := time.Duration(0)

	if len(validLatencies) > 0 {
		avgLatency = totalLatency / time.Duration(len(validLatencies))
		p50 = validLatencies[len(validLatencies)*50/100]
		p95 = validLatencies[len(validLatencies)*95/100]
		p99 = validLatencies[len(validLatencies)*99/100]
	}

	opsPerSec := float64(totalSuccess+totalFail) / totalDuration.Seconds()

	fmt.Println("=== Load Test Results ===")
	fmt.Printf("Total Duration:      %v\n", totalDuration)
	fmt.Printf("Total Operations:    %d\n", totalSuccess+totalFail)
	fmt.Printf("Successful Ops:      %d\n", totalSuccess)
	fmt.Printf("Failed Ops:          %d\n", totalFail)
	fmt.Printf("Throughput:          %.2f ops/sec\n", opsPerSec)
	fmt.Printf("Average Latency:     %v\n", avgLatency)
	fmt.Printf("p50 Latency:         %v\n", p50)
	fmt.Printf("p95 Latency:         %v\n", p95)
	fmt.Printf("p99 Latency:         %v\n", p99)
}

func handleConn(conn net.Conn, c *cache.MemoryCache) {
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

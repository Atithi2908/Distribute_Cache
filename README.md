# Distributed Cache System in Go

A Redis-like distributed in-memory cache system built from scratch in idiomatic Go without third-party frameworks.

---

## 🌟 Architecture Overview

```
                          Client / Load Tester
                                   |
                                   v
                             Router Layer
                                   |
                       Consistent Hash Ring
                 (CRC32 + Virtual Nodes: 100 per node)
                                   |
            +----------------------+----------------------+
            |                      |                      |
            v                      v                      v
         Node A                 Node B                 Node C
      (Primary / Replica)    (Primary / Replica)    (Primary / Replica)
            |                      |                      |
            v                      v                      v
       MemoryCache            MemoryCache            MemoryCache
    (TTL + LRU + Mutex)    (TTL + LRU + Mutex)    (TTL + LRU + Mutex)
```

---

## 🔥 Key Features (Phases 1–7)

### Phase 1 & 2: Thread-Safe In-Memory Storage
- Synchronized with `sync.RWMutex` for high-throughput concurrent reads and exclusive writes.
- Clean standard `Cache` interface: `Set`, `SetWithTTL`, `Get`, `Delete`.

### Phase 3 & 4: TCP Server & Consistent Hashing
- Custom TCP server with command parser (`SET`, `GET`, `DELETE`, `DUMP`).
- Hash ring using CRC32 checksums and virtual nodes (default: 100 virtual nodes per physical node) for uniform distribution and minimal remapping on topology changes.

### Phase 5: Masterless Replication & Quorum Consensus
- **Replication Factor**: $N = 3$.
- **Write Quorum**: $W = 2$ (requires at least 2 successful node writes before returning success).
- **Read Quorum**: $R = 2$ (queries multiple nodes concurrently to achieve consensus).
- **Health Tracking & Node Recovery**: Automatic health checks, marking failing nodes unhealthy, and `DUMP` / `RestoreNode` support for node bootstrapping.

### Phase 6: Advanced Cache Features (TTL & LRU)
- **TTL / Expiration**:
  - Optional per-key TTL (e.g. `SET key value 60` for 60 seconds TTL).
  - **Lazy Expiration**: Checked on `Get` access; expired keys are deleted on access.
  - **Background Cleanup**: Goroutine ticker (`StartCleanup(interval)`) periodically removes expired entries without holding locks longer than necessary.
- **LRU Eviction**:
  - $O(1)$ operations using `map[string]*list.Element` combined with Go standard library `container/list`.
  - Configurable capacity via `NewMemoryCacheWithCapacity(cap)`.
  - When capacity is reached, least recently used items are evicted.
- **Protocol Updates**: Support for `SET key value [ttl_seconds]`.
- **Replication + TTL**: Router preserves relative TTL across all $N=3$ replicas.

### Phase 7: Benchmarking, Connection Pooling & Load Testing
- **Connection Pooling**: Bounded TCP connection pool per node (`chan net.Conn`) in the router to eliminate socket dial overhead per request.
- **Comprehensive Benchmarks**: Measuring ns/op, B/op, and allocs/op for GET/SET/DELETE/TTL/LRU and parallel operations.
- **Load Testing CLI**: Command-line tool `cmd/loadtest/main.go` supporting configurable concurrency, read/write ratio, operation count, TTL, and reporting throughput + p50/p95/p99 latency.

---

## 🛠️ TCP Protocol Specification

| Command | Format | Example | Description |
|---|---|---|---|
| **SET** | `SET <key> <value> [ttl]` | `SET user123 Atithi 60` | Sets key to value with optional TTL (in seconds) |
| **GET** | `GET <key>` | `GET user123` | Retrieves value for key (returns `NOT_FOUND` if expired/missing) |
| **DELETE** | `DELETE <key>` | `DELETE user123` | Removes key from cache |
| **DUMP** | `DUMP` | `DUMP` | Dumps all non-expired key-value pairs (ends with `END`) |

---

## 🚀 Running Tests & Benchmarks

### Unit Tests
```bash
go test -v ./internal/...
```

### Benchmarks
```bash
# Benchmark memory cache
go test -bench=Benchmark ./internal/cache

# Benchmark distributed router & replication
go test -bench=Benchmark -benchtime=1s ./internal/router
```

### Load Testing Tool
```bash
# Run load test with 50 clients, 10,000 operations (80% GET, 20% SET)
go run ./cmd/loadtest/main.go -clients 50 -ops 10000 -ratio 0.8
```

---

## 📊 Performance Benchmark Summary

### In-Memory Cache Performance (Intel Core i5-12450H)
| Operation | Latency (ns/op) | Memory (B/op) | Allocations |
|---|---|---|---|
| `GET` | 40.16 ns | 8 B | 1 alloc |
| `SET` | 39.27 ns | 8 B | 1 alloc |
| `DELETE` | 23.76 ns | 0 B | 0 allocs |
| `GET` with TTL | 43.35 ns | 8 B | 1 alloc |
| `SET` with TTL | 43.60 ns | 8 B | 1 alloc |
| `LRU Eviction` | 206.70 ns | 128 B | 4 allocs |
| `Parallel GET` | 84.39 ns | 8 B | 1 alloc |
| `Parallel SET` | 167.80 ns | 22 B | 2 allocs |
| `Parallel Mixed` | 101.50 ns | 10 B | 1 alloc |

### Distributed Load Test Results (3 Replicated Local Nodes, Connection Pooled)
- **Throughput**: ~28,640 operations / sec
- **Average Latency**: 633 µs
- **p50 Latency**: < 100 µs
- **p95 Latency**: 2.25 ms
- **p99 Latency**: 5.18 ms

---

## 📌 Architecture & Design Decisions

1. **Idiomatic Go Standard Library**: Built without heavy external dependencies. Uses standard `sync.RWMutex`, `container/list`, `net`, `bufio`, `time`, and `sort`.
2. **Double Expiration Strategy**:
   - *Lazy expiration* ensures immediate miss on access without waiting for cleanup cycles.
   - *Background ticker cleanup* guarantees abandoned keys are purged over time without blocking active GET requests.
3. **Double Linked List + Map LRU**: Combines $O(1)$ lookup speed of Go map with $O(1)$ eviction/ordering updates of doubly linked lists.
4. **Connection Pool per Node**: Avoids TCP handshake overhead for every GET/SET operation while protecting against socket corruption via automatic conn replacement on I/O failure.

---

## ⚠️ Known Limitations & Future Work

- **In-Memory Volatility**: Nodes do not persist state to disk on restart (snapshots/AOF append-only logs can be added).
- **Cluster Membership Protocol**: Router health check uses periodic polling instead of a gossip protocol (e.g. SWIM/Gossip).
- **Network Boundaries**: TCP commands use space-separated strings; binary protocol (e.g. RESP or protobuf) could be implemented for arbitrary binary values containing spaces or newlines.

package server

import (
	"bufio"
	"distributed-cache/internal/cache"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestServerSetAndGet(t *testing.T) {
	c := cache.NewMemoryCache()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		handleConnection(conn, c)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// SET
	_, err = conn.Write([]byte("SET name Atithi\n"))
	if err != nil {
		t.Fatal(err)
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	if response != "OK \n" {
		t.Fatalf("expected OK, got %q", response)
	}

	// GET
	_, err = conn.Write([]byte("GET name\n"))
	if err != nil {
		t.Fatal(err)
	}

	response, err = reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	if response != "OK Atithi\n" {
		t.Fatalf("expected OK Atithi, got %q", response)
	}

	time.Sleep(10 * time.Millisecond)
}

func TestServerConcurrentClients(t *testing.T) {
	c := cache.NewMemoryCache()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go handleConnection(conn, c)
		}
	}()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", listener.Addr().String())
			if err != nil {
				t.Errorf("connection failed: %v", err)
				return
			}
			defer conn.Close()

			reader := bufio.NewReader(conn)

			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)

			// SET
			fmt.Fprintf(conn, "SET %s %s\n", key, value)

			response, err := reader.ReadString('\n')
			if err != nil {
				t.Errorf("read failed: %v", err)
				return
			}

			if response != "OK \n" {
				t.Errorf("expected OK, got %q", response)
			}

			// GET
			fmt.Fprintf(conn, "GET %s\n", key)

			response, err = reader.ReadString('\n')
			if err != nil {
				t.Errorf("read failed: %v", err)
				return
			}

			expected := fmt.Sprintf("OK %s\n", value)

			if response != expected {
				t.Errorf("expected %q, got %q", expected, response)
			}
		}(i)
	}

	wg.Wait()
}

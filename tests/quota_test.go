package main

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestClientQueueQuota(t *testing.T) {
	// Adjust address to match your server's listening port
	serverAddr := "127.0.0.1:1080"

	// We will simulate 100 concurrent clients hitting a server with no peers
	clientCount := 100
	var wg sync.WaitGroup

	startSignal := make(chan struct{})

	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Wait for start signal to hit the server simultaneously
			<-startSignal

			conn, err := net.DialTimeout("tcp", serverAddr, 2*time.Second)
			if err != nil {
				// We don't t.Fatal here because we want to see
				// how many successfully made it through
				return
			}
			defer conn.Close()

			// 2. Send some initial payload data
			payload := fmt.Appendf(nil, "Client-Payload-%d", id)
			_, err = conn.Write(payload)
			if err != nil {
				return
			}

			// 3. Keep the connection open to simulate a persistent client
			// waiting for a tunnel assignment
			time.Sleep(10 * time.Second)
		}(i)
	}

	fmt.Printf("Launching %d clients against %s...\n", clientCount, serverAddr)
	close(startSignal) // Release all goroutines at once
	wg.Wait()

	fmt.Println("Test finished. Check server logs to ensure all clients were added to the pool.")
}

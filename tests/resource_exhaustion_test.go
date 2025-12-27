package main

import (
	"01proxy/model"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func TestResourceExhaustion(t *testing.T) {
	target := "127.0.0.1:1080"
	concurrency := 50_000
	var wg sync.WaitGroup
	connections := make(map[net.Conn]struct{}, concurrency)
	mx := sync.Mutex{}

	t.Logf("Testing with %d connection", concurrency)
	start := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Dial the server
			conn, err := net.DialTimeout("tcp", target, 2*time.Second)
			if err != nil {
				log.Fatal(err)
				return
			}

			// 1. Send Protocol Prefix [cite: 14]
			conn.Write(model.PREFIX())

			// 2. Send maximum uint16 length (65535)
			// This triggers: commandBuf := make([]byte, length) in ReadCommand
			lenBuf := make([]byte, 2)
			binary.BigEndian.PutUint16(lenBuf, 65535)

			// 3. Flood the server with these requests
			conn.Write(lenBuf)
			mx.Lock()
			connections[conn] = struct{}{}
			fmt.Printf("%s - Performed %d connection\n", time.Since(start), len(connections))
			mx.Unlock()
		}(i)
	}

	wg.Wait()
	time.Sleep(30 * time.Second)
}

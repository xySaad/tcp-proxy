package main

import (
	"01proxy/model"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	PEER_REQUEST    = []byte("==PEER==REQUEST")
	PEER_ACCEPTED   = []byte("==PEER==ACCEPTED")
	START_BRIDGE    = []byte("==START==BRIDGE")
	BRIDGE_REJECTED = []byte("==BRIDGE==REJECTED")
	BRIDGE_ACCEPTED = []byte("==BRIDGE==ACCEPTED")
	TUNNEL_REQUEST  = []byte("==TUNNEL==REQUEST")
	TUNNEL_ACCEPTED = []byte("==TUNNEL==ACCEPTED")
)

const serverAddr = "localhost:1080"

// Test 1: Race condition in Pool.Find() - multiple goroutines accessing pool during iteration
func TestRace_ConcurrentPeerAccessDuringIteration(t *testing.T) {
	log.Println("Test 1: Concurrent peer access during pool iteration")

	var wg sync.WaitGroup

	// Connect multiple peers concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				log.Printf("Peer %d failed to connect: %v", id, err)
				return
			}
			defer conn.Close()

			model.WriteHeader(conn, PEER_REQUEST)
			buffer := make([]byte, len(PEER_ACCEPTED))
			io.ReadFull(conn, buffer)

			log.Printf("Peer %d connected", id)
			time.Sleep(2 * time.Second)
		}(i)
		time.Sleep(50 * time.Millisecond)
	}

	// While peers are connecting, send clients concurrently
	time.Sleep(200 * time.Millisecond)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				log.Printf("Client %d failed: %v", id, err)
				return
			}
			defer conn.Close()

			// Send regular client data
			conn.Write([]byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"))
			log.Printf("Client %d sent data", id)
		}(i)
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	log.Println("Test 1 completed")
}

// Test 2: Race on Pool.RemoveBy during iteration in ForEach
func TestRace_RemoveDuringForEach(t *testing.T) {
	log.Println("Test 2: Remove peer during client queue processing")

	var wg sync.WaitGroup

	// Add a peer
	peerConn, _ := net.Dial("tcp", serverAddr)
	defer peerConn.Close()
	model.WriteHeader(peerConn, PEER_REQUEST)
	buffer := make([]byte, len(PEER_ACCEPTED))
	io.ReadFull(peerConn, buffer)

	// Queue multiple clients
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, _ := net.Dial("tcp", serverAddr)
			defer conn.Close()
			conn.Write([]byte(fmt.Sprintf("CLIENT-%d-DATA-PAYLOAD", id)))
		}(i)
		time.Sleep(50 * time.Millisecond)
	}

	// Disconnect peer to trigger RemoveBy during handleClient
	time.Sleep(200 * time.Millisecond)
	peerConn.Close()

	// Add another peer to trigger ForEach on queued clients
	time.Sleep(100 * time.Millisecond)
	newPeer, _ := net.Dial("tcp", serverAddr)
	defer newPeer.Close()
	model.WriteHeader(newPeer, PEER_REQUEST)
	io.ReadFull(newPeer, buffer)

	wg.Wait()
	log.Println("Test 2 completed")
}

// Test 3: Race on Peer.StartBridge mutex vs Pool.Find accessing Peer.Quota
func TestRace_PeerQuotaAccessDuringStartBridge(t *testing.T) {
	log.Println("Test 3: Concurrent quota checks during StartBridge")

	// Connect peer
	peerConn, _ := net.Dial("tcp", serverAddr)
	defer peerConn.Close()
	model.WriteHeader(peerConn, PEER_REQUEST)
	buffer := make([]byte, len(PEER_ACCEPTED))
	io.ReadFull(peerConn, buffer)

	// Simulate slow bridge acceptance
	go func() {
		for {
			buf := make([]byte, len(START_BRIDGE))
			_, err := io.ReadFull(peerConn, buf)
			if err != nil {
				return
			}
			if bytes.Equal(buf, START_BRIDGE) {
				time.Sleep(100 * time.Millisecond) // Delay response
				peerConn.Write(BRIDGE_ACCEPTED)
			}
		}
	}()

	var wg sync.WaitGroup
	// Send multiple clients rapidly to trigger concurrent NextPeer calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, _ := net.Dial("tcp", serverAddr)
			defer conn.Close()
			conn.Write([]byte(fmt.Sprintf("CLIENT-%d", id)))
			time.Sleep(500 * time.Millisecond)
		}(i)
		time.Sleep(20 * time.Millisecond)
	}

	wg.Wait()
	log.Println("Test 3 completed")
}

// Test 4: Race on Client.buffer field (not protected by mutex)
func TestRace_ClientBufferAccess(t *testing.T) {
	log.Println("Test 4: Client buffer accessed from multiple places")

	// This tests the race where Client.buffer is read in handleClient
	// but the Client struct itself isn't protected

	var wg sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, _ := net.Dial("tcp", serverAddr)
			defer conn.Close()

			// Send partial data that becomes client buffer
			data := []byte(fmt.Sprintf("PARTIAL-DATA-%d-XXXXXXXXXX", id))
			conn.Write(data)
			time.Sleep(100 * time.Millisecond)
		}(i)
		time.Sleep(30 * time.Millisecond)
	}

	wg.Wait()
	log.Println("Test 4 completed")
}

// Test 5: Race on Tunnels channel with concurrent sends
func TestRace_TunnelChannelAccess(t *testing.T) {
	log.Println("Test 5: Concurrent tunnel registration")

	var wg sync.WaitGroup

	// Connect peer first
	peerConn, _ := net.Dial("tcp", serverAddr)
	defer peerConn.Close()
	model.WriteHeader(peerConn, PEER_REQUEST)
	buffer := make([]byte, len(PEER_ACCEPTED))
	io.ReadFull(peerConn, buffer)

	// Handle bridge requests
	go func() {
		for {
			buf := make([]byte, len(START_BRIDGE))
			_, err := io.ReadFull(peerConn, buf)
			if err != nil {
				return
			}
			if bytes.Equal(buf, START_BRIDGE) {
				peerConn.Write(BRIDGE_ACCEPTED)
				// Send tunnel ID after accepting bridge
				idBuf := make([]byte, 8)
				binary.BigEndian.PutUint64(idBuf, uint64(time.Now().UnixNano()))
				peerConn.Write(idBuf)
			}
		}
	}()

	// Send client
	go func() {
		conn, _ := net.Dial("tcp", serverAddr)
		defer conn.Close()
		conn.Write([]byte("CLIENT-DATA"))
		time.Sleep(2 * time.Second)
	}()

	time.Sleep(100 * time.Millisecond)

	// Send multiple tunnels concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, _ := net.Dial("tcp", serverAddr)
			defer conn.Close()

			// Send TUNNEL_REQUEST header
			model.WriteHeader(conn, TUNNEL_REQUEST)

			// Send tunnel ID (8 bytes)
			idBuf := make([]byte, 8)
			binary.BigEndian.PutUint64(idBuf, uint64(id))
			conn.Write(idBuf)

			// Wait for TUNNEL_ACCEPTED
			buf := make([]byte, len(TUNNEL_ACCEPTED))
			io.ReadFull(conn, buf)
			log.Printf("Tunnel %d registered", id)
			time.Sleep(1 * time.Second)
		}(i)
		time.Sleep(20 * time.Millisecond)
	}

	wg.Wait()
	log.Println("Test 5 completed")
}

// Test 6: Race in Pool.RemoveBy - modifying slice during iteration
func TestRace_RemoveBySliceModification(t *testing.T) {
	log.Println("Test 6: Slice modification during RemoveBy")

	var wg sync.WaitGroup

	// Add multiple peers
	peers := make([]net.Conn, 5)
	for i := 0; i < 5; i++ {
		conn, _ := net.Dial("tcp", serverAddr)
		peers[i] = conn
		model.WriteHeader(conn, PEER_REQUEST)
		buffer := make([]byte, len(PEER_ACCEPTED))
		io.ReadFull(conn, buffer)
		time.Sleep(50 * time.Millisecond)
	}

	// Close peers concurrently to trigger RemoveBy
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			time.Sleep(time.Duration(idx*30) * time.Millisecond)
			peers[idx].Close()
		}(i)
	}

	// Send clients to trigger peer iteration
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _ := net.Dial("tcp", serverAddr)
			defer conn.Close()
			conn.Write([]byte("CLIENT"))
		}()
		time.Sleep(25 * time.Millisecond)
	}

	wg.Wait()
	log.Println("Test 6 completed")
}

func main() {
	log.Println("Starting race condition tests...")
	log.Println("Make sure the server is running on", serverAddr)
	log.Println("Run with: go run -race race_test.go")
	log.Println("")

	t := &testing.T{}

	TestRace_ConcurrentPeerAccessDuringIteration(t)
	time.Sleep(1 * time.Second)

	TestRace_RemoveDuringForEach(t)
	time.Sleep(1 * time.Second)

	TestRace_PeerQuotaAccessDuringStartBridge(t)
	time.Sleep(1 * time.Second)

	TestRace_ClientBufferAccess(t)
	time.Sleep(1 * time.Second)

	TestRace_TunnelChannelAccess(t)
	time.Sleep(1 * time.Second)

	TestRace_RemoveBySliceModification(t)

	log.Println("\nAll tests completed!")
	log.Println("Check for race detector warnings")
}

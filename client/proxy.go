package main

import (
	"01proxy/model"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type ProxyHandler struct{}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle CONNECT method for HTTPS tunneling
	if r.Method != http.MethodConnect {
		http.Error(w, "Only HTTPS (CONNECT) is supported", http.StatusMethodNotAllowed)
		log.Printf("Rejected non-CONNECT request: %s %s", r.Method, r.URL)
		return
	}

	log.Printf("CONNECT request: %s", r.Host)

	// Dial the target server
	targetConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, "Failed to connect to target", http.StatusServiceUnavailable)
		log.Printf("Failed to connect to %s: %v", r.Host, err)
		return
	}
	defer targetConn.Close()

	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		log.Println("Hijacking not supported")
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		log.Printf("Failed to hijack: %v", err)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
		return
	}

	// Bidirectional copy
	go transfer(targetConn, clientConn)
	transfer(clientConn, targetConn)

	log.Printf("Connection closed: %s", r.Host)
}

func transfer(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func Proxy() *net.TCPAddr {
	proxy := &ProxyHandler{}
	server := &http.Server{
		Addr:         ":0",
		Handler:      proxy,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	log.Println("This proxy only supports HTTPS (CONNECT method)")
	ln, tcpAddr, err := model.ServerListen(server)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	go func() {
		err = server.Serve(ln)
		if err != nil {
			log.Fatal(err)
			return
		}
	}()
	log.Println("Starting HTTPS proxy on :", tcpAddr)

	return tcpAddr
}

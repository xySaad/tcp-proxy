//go:build !test
// +build !test

package main

import (
	"01proxy/client"
	"01proxy/server"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/xySaad/snapshot"
)

const (
	RETRY_LIMIT = 10
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", "tcprp <client|server> <?address>")
		return
	}

	address := ":1080"

	if len(os.Args) > 2 {
		address = os.Args[2]
		if !strings.Contains(address, ":") {
			address += ":1080"
		}
	}

	if os.Args[1] == "server" {
		srv, err := server.New(address)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Listening", address)
		srv.Run()
		return
	}

	if os.Args[1] != "client" {
		fmt.Println("Usage:", "tcprp <client|server> <?address>")
		return
	}

	proxyAddr := client.Proxy()

	var delay time.Duration = 1
	snapshot.Retry(RETRY_LIMIT, func(s *snapshot.Snapshot) {
		snapshot.Freeze(s, &delay)
		client.Client(address, proxyAddr, s)
		s.BreakPoint(func() {
			log.Println("Retrying after", time.Second*delay)
			<-time.Tick(time.Second * delay)
			delay *= 2
		})
	})

}

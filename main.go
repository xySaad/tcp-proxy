//go:build !test
// +build !test

package main

import (
	"01proxy/client"
	"01proxy/server"
	"log"
	"os"
	"time"

	"github.com/xySaad/snapshot"
)

const (
	RETRY_LIMIT = 10
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		adress := "0.0.0.0:1080"
		if len(os.Args) > 2 && adress == "" {
			adress = os.Args[2]
		}
		srv, err := server.New(adress)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Listening", adress)
		srv.Run()
	} else {
		address := client.Proxy()

		var delay time.Duration = 1
		snapshot.Retry(RETRY_LIMIT, func(s *snapshot.Snapshot) {
			snapshot.Freeze(s, &delay)
			client.Client(address, s)
			s.BreakPoint(func() {
				log.Println("Retrying after", time.Second*delay)
				<-time.Tick(time.Second * delay)
				delay *= 2
			})
		})
	}
}

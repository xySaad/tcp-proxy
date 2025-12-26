package main

import (
	"01proxy/client"
	"01proxy/server"
	"log"
	"os"
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
		client.Client(client.Proxy())
	}
}

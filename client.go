package main

import (
	"01proxy/model"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func Client(proxyAdress *net.TCPAddr) {
	remoteServer, err := net.Dial("tcp", "0.0.0.0:1080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer remoteServer.Close()
	fmt.Println("Connected to the remote server")

	_, err = remoteServer.Write(model.PEER_REQUEST)
	if err != nil {
		fmt.Println(err)
		return
	}

	buffer := make([]byte, len(model.PEER_ACCEPTED))
	_, err = io.ReadFull(remoteServer, buffer)
	if err != nil || !bytes.Equal(buffer, model.PEER_ACCEPTED) {
		fmt.Println("server proxy rejected - ", err)
		return
	}

	buffer = make([]byte, len(model.START_BRIDGE))
	for {
		_, err = io.ReadFull(remoteServer, buffer)
		if err == io.EOF {
			log.Fatal("remote server has been stoped", err)
			return
		}
		if err != nil || !bytes.Equal(buffer, model.START_BRIDGE) {
			fmt.Println("bridge rejected - ", err, string(buffer))
			continue
		}

		log.Println("Bridgging...")
		localProxy, err := net.DialTCP("tcp", nil, proxyAdress)
		if err != nil {
			fmt.Println(err)
			remoteServer.Write(model.BRIDGE_REJECTED)
			continue
		}
		remoteServer.Write(model.BRIDGE_ACCEPTED)
		model.BiCopy(remoteServer, localProxy)
		log.Println("Bridgging DONE!")
	}
}

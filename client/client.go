package client

import (
	"01proxy/model"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func PeerHandshake() (net.Conn, error) {
	remoteServer, err := net.Dial("tcp", "0.0.0.0:1080")
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to the remote server")

	_, err = remoteServer.Write(model.PEER_REQUEST)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, len(model.PEER_ACCEPTED))
	_, err = io.ReadFull(remoteServer, buffer)
	if err != nil || !bytes.Equal(buffer, model.PEER_ACCEPTED) {
		return nil, fmt.Errorf("server proxy rejected - %s", err)
	}

	return remoteServer, nil
}

func TunnelHandshake() (net.Conn, error) {
	remoteServer, err := net.Dial("tcp", "0.0.0.0:1080")
	if err != nil {
		return nil, err
	}

	_, err = remoteServer.Write(model.TUNNEL_REQUEST)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, len(model.TUNNEL_ACCEPTED))
	_, err = io.ReadFull(remoteServer, buffer)
	if err != nil || !bytes.Equal(buffer, model.TUNNEL_ACCEPTED) {
		return nil, fmt.Errorf("tunnel rejected - %s", err)
	}

	return remoteServer, nil
}

func Client(proxyAdress *net.TCPAddr) {
	remoteServer, err := PeerHandshake()
	if err != nil {
		log.Println(err)
		return
	}

	buffer := make([]byte, len(model.START_BRIDGE))
	for {
		_, err = io.ReadFull(remoteServer, buffer)
		if err == io.EOF {
			log.Fatal("remote server has been stoped", err)
			return
		}
		if err != nil || !bytes.Equal(buffer, model.START_BRIDGE) {
			fmt.Println(err, "ignore buffer:", string(buffer))
			continue
		}

		log.Println("Bridgging...")
		localProxy, err := net.DialTCP("tcp", nil, proxyAdress)
		if err != nil {
			fmt.Println("bridge rejected - ", err)
			remoteServer.Write(model.BRIDGE_REJECTED)
			continue
		}
		remoteServer.Write(model.BRIDGE_ACCEPTED)

		log.Println("Bridgging DONE!")
		go func() {
			tunnel, err := TunnelHandshake()
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Println("Connected as tunnel")
			model.BiCopy(tunnel, localProxy)
			log.Println("Tunnel Closed", tunnel.LocalAddr())
		}()
	}
}

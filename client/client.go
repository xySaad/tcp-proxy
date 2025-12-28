package client

import (
	"01proxy/model"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/xySaad/snapshot"
)

func PeerHandshake() (net.Conn, error) {
	remoteServer, err := net.Dial("tcp", "0.0.0.0:1080")
	if err != nil {
		return nil, err
	}
	defer func(err error) {
		if err != nil {
			remoteServer.Close()
		}
	}(err)

	_, err = model.WriteHeader(remoteServer, model.PEER_REQUEST())
	if err != nil {
		return nil, err
	}

	buffer := model.PEER_ACCEPTED()
	_, err = io.ReadFull(remoteServer, buffer)
	if err != nil || !bytes.Equal(buffer, model.PEER_ACCEPTED()) {
		return nil, fmt.Errorf("server proxy rejected - %s", err)
	}

	return remoteServer, nil
}

func TunnelHandshakeWithID(id uint64) (net.Conn, error) {
	remoteServer, err := net.Dial("tcp", "0.0.0.0:1080")
	if err != nil {
		return nil, err
	}

	defer func(err error) {
		if err != nil {
			remoteServer.Close()
		}
	}(err)

	_, err = model.WriteHeader(remoteServer, model.TUNNEL_REQUEST())
	if err != nil {
		return nil, err
	}

	idBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(idBuf, id)
	if _, err := remoteServer.Write(idBuf); err != nil {
		return nil, err
	}

	buffer := model.TUNNEL_ACCEPTED()
	_, err = io.ReadFull(remoteServer, buffer)
	if err != nil || !bytes.Equal(buffer, model.TUNNEL_ACCEPTED()) {
		return nil, fmt.Errorf("tunnel rejected - %s - %s", err, string(buffer))
	}

	return remoteServer, nil
}

func Client(proxyAdress *net.TCPAddr, reset *snapshot.Snapshot) {
	remoteServer, err := PeerHandshake()
	if err != nil {
		log.Printf("[CLIENT] Peer handshake failed: %v", err)
		return
	}
	defer remoteServer.Close()
	log.Printf("[CLIENT] Connected to remote server")
	reset.Reset()
	for {
		command, err := model.ReadCommand(remoteServer)
		if err != nil {
			log.Printf("[CLIENT] Remote server disconnected")
			break
		}

		if bytes.Equal(command, model.KEEP_ALIVE()) {
			remoteServer.Write(model.KEEP_ALIVE())
			continue
		}

		if !bytes.Equal(command, model.START_BRIDGE()) {
			continue
		}

		localProxy, err := net.DialTCP("tcp", nil, proxyAdress)
		if err != nil {
			remoteServer.Write(model.BRIDGE_REJECTED())
			continue
		}
		remoteServer.Write(model.BRIDGE_ACCEPTED())

		idBuf := make([]byte, 8)
		if _, err := io.ReadFull(remoteServer, idBuf); err != nil {
			log.Printf("[CLIENT] Failed to read tunnel ID: %v", err)
			break
		}
		id := binary.BigEndian.Uint64(idBuf)

		go func() {
			tunnel, err := TunnelHandshakeWithID(id)
			if err != nil {
				log.Printf("[TUNNEL] Handshake failed: %v", err)
				return
			}
			defer tunnel.Close()
			log.Printf("[TUNNEL] Established connection (ID: %d)", id)
			model.BiCopy(tunnel, localProxy)
			log.Printf("[TUNNEL] Connection closed (ID: %d)", id)
		}()
	}
}

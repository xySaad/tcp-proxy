package server

import (
	"01proxy/model"
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
)

type Peer struct {
	Quota   int
	Conn    net.Conn
	Tunnels map[string]net.Conn
	mx      sync.Mutex
}

func (p *Pool) NextPeer() (peer *Peer) {
	return p.Peers.Find(func(p *Peer) bool {
		if p.Quota < model.MAX_PEER_QUOTA {
			return true
		}
		return false
	})
}

func (p *Peer) StartBridge() error {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.Conn.Write(model.START_BRIDGE)
	buffer := make([]byte, len(model.BRIDGE_ACCEPTED))
	_, err := io.ReadFull(p.Conn, buffer)
	if err != nil || !bytes.Equal(buffer, model.BRIDGE_ACCEPTED) {
		return fmt.Errorf("%s BRIDGE_REJECTED - %s", err, string(buffer))
	}
	return nil
}

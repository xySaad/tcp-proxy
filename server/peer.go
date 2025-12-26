package server

import (
	"01proxy/model"
	"01proxy/model/mutex"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Peer struct {
	Quota int
	Conn  net.Conn
	Mx    mutex.Mutex
}

func (p *Pool) NextPeer() (peer *Peer) {
	return p.Peers.Find(func(p *Peer) bool {
		if p.Quota < model.MAX_PEER_QUOTA {
			return true
		}
		return false
	})
}

func (s *Server) handlePeer(p *Peer) {
	s.pool.Peers.Add(p)
	go func() {
		clients := s.pool.Clients.Clear()
		if len(clients) < 1 {
			return
		}

		fmt.Printf("processing %d queue clients\n", len(clients))
		for _, c := range clients {
			go s.handleClient(c)
		}
	}()
}
func (p *Peer) StartBridge(id uint64) error {
	if _, err := p.Conn.Write(model.START_BRIDGE); err != nil {
		return err
	}

	buffer := make([]byte, len(model.BRIDGE_ACCEPTED))
	_, err := io.ReadFull(p.Conn, buffer)
	if err != nil {
		return err
	}

	if !bytes.Equal(buffer, model.BRIDGE_ACCEPTED) {
		return fmt.Errorf("buffer didn't match %s: %s", model.BRIDGE_ACCEPTED, buffer)
	}
	idBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(idBuf, id)
	if _, err := p.Conn.Write(idBuf); err != nil {
		return err
	}
	return nil
}

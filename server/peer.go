package server

import (
	"01proxy/model"
	"01proxy/model/constants"
	"01proxy/model/mutex"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

type Peer struct {
	Quota int
	Conn  net.Conn
	Mx    mutex.Mutex
}

func (pool *Pool) LockedNextPeer() *Peer {
	return pool.Peers.Find(func(p *Peer) bool {
		if !p.Mx.TryLock() {
			return false
		}

		if p.Quota < constants.MAX_PEER_QUOTA {
			p.Quota++
			return true
		}

		p.Mx.Unlock()
		return false
	})
}

func (s *Server) handlePeer(p *Peer) {
	s.pool.Peers.Add(p)
	clients := s.pool.Clients.Clear()
	if len(clients) < 1 {
		return
	}

	log.Printf("processing %d queue clients\n", len(clients))
	for _, c := range clients {
		go s.handleClient(c)
	}
}
func (p *Peer) StartBridge(id uint64) error {
	if err := model.WriteCommand(p.Conn, constants.START_BRIDGE()); err != nil {
		return err
	}

	buffer := constants.BRIDGE_ACCEPTED()
	_, err := io.ReadFull(p.Conn, buffer)
	if err != nil {
		return err
	}

	if !bytes.Equal(buffer, constants.BRIDGE_ACCEPTED()) {
		return fmt.Errorf("buffer didn't match %s: %s", constants.BRIDGE_ACCEPTED(), buffer)
	}
	idBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(idBuf, id)
	if _, err := p.Conn.Write(idBuf); err != nil {
		return err
	}
	return nil
}

func (p *Peer) KeepAlive() bool {
	if !p.Mx.TryLock() {
		return true
	}
	defer p.Mx.Unlock()

	if nil != model.WriteCommand(p.Conn, constants.KEEP_ALIVE()) {
		return false
	}

	_, err := io.ReadFull(p.Conn, constants.KEEP_ALIVE())
	if err != nil {
		return false
	}

	return false
}

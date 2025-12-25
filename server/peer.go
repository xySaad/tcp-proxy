package server

import (
	"01proxy/model"
	"net"
)

type Peer struct {
	Quota int
	Conn  net.Conn
}

func (p *Pool) NextPeer() (peer *Peer) {
	return p.Peers.Find(func(p Peer) bool {
		if p.Quota < model.MAX_PEER_QUOTA {
			return true
		}
		return false
	})
}

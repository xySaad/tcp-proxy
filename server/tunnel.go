package server

import (
	"01proxy/model/constants"
	"log"
	"net"
)

type Tunnel struct {
	ID   uint64
	Conn net.Conn
}

func (s *Server) handleTunnel(conn Tunnel) {
	ch, ok := s.pool.TunnelMap.Get(conn.ID)
	if !ok {
		log.Println("Tunnel connected with invalid id", conn.ID)
		conn.Conn.Close()
		return
	}

	_, err := conn.Conn.Write(constants.TUNNEL_ACCEPTED())
	if err != nil {
		s.pool.TunnelMap.Delete(conn.ID)
		ch <- nil
		return
	}
	log.Println("Sending Tunnel with id", conn.ID, "to channel")
	ch <- conn.Conn
}

package server

import (
	"01proxy/model"
	"bytes"
	"io"
	"log"
	"net"
	"os"
)

type Pool struct {
	Clients model.Pool[Client]
	Peers   model.Pool[Peer]
	Tunnels chan net.Conn
}

type Server struct {
	pool Pool
	ln   net.Listener
}

func New(adress string) (*Server, error) {
	if len(os.Args) > 2 && adress == "" {
		adress = os.Args[2]
	}

	ln, err := net.Listen("tcp", adress)
	if err != nil {
		return nil, err
	}

	return &Server{
		pool: Pool{
			Tunnels: make(chan net.Conn),
		},
		ln: ln,
	}, nil
}

func (s *Server) handleClient(client Client) {
	peer := s.pool.NextPeer()
	// queue the conn until a peer is available
	if peer == nil {
		s.pool.Clients.Add(client)
		log.Println("Client queued", client.Conn.RemoteAddr(), "- Total:", s.pool.Clients.Size())
		return
	}

	peer.Conn.Write(model.START_BRIDGE)
	buffer := make([]byte, len(model.BRIDGE_ACCEPTED))
	_, err := io.ReadFull(peer.Conn, buffer)
	if err != nil || !bytes.Equal(buffer, model.BRIDGE_ACCEPTED) {
		log.Println(err, "BRIDGE_REJECTED - ", string(buffer))
		return
	}

	tunnel := <-s.pool.Tunnels
	log.Println("copy tunnel and client")
	tunnel.Write(client.buffer)
	model.BiCopy(tunnel, client.Conn)
	log.Println("DONE: client-peer copy")
}

func (s *Server) Run() {
	for {
		conn, err := s.nextConn()
		if err != nil {
			log.Println(err)
			continue
		}

		switch conn := conn.(type) {
		case Peer:
			s.pool.Peers.Add(Peer{Quota: 0, Conn: conn.Conn})
			log.Println("Peer added", conn.Conn.RemoteAddr(), "- Total:", s.pool.Peers.Size())
			go s.pool.Clients.ForEach(s.handleClient)
		case net.Conn:
			s.pool.Tunnels <- conn
			log.Println("Tunnel added", conn.RemoteAddr())
		case Client:
			go s.handleClient(conn)
		}
	}
}

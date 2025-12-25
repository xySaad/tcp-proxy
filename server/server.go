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
		pool: Pool{},
		ln:   ln,
	}, nil
}

func (s *Server) handleClient(client *Client) {
	peer := s.pool.NextPeer()
	// queue the conn until a peer is available
	if peer == nil {
		s.pool.Clients.Add(*client)
		log.Println("Client queued", client.conn.RemoteAddr(), "- Total:", s.pool.Clients.Size())
		return
	}

	peer.Conn.Write(model.START_BRIDGE)
	buffer := make([]byte, len(model.BRIDGE_ACCEPTED))
	_, err := io.ReadFull(peer.Conn, buffer)
	if err != nil || !bytes.Equal(buffer, model.BRIDGE_ACCEPTED) {
		log.Println("BRIDGE_REJECTED - ", err)
		return
	}

	peer.Conn.Write(client.buffer)
	log.Println("copy peer and client")
	model.BiCopy(peer.Conn, client.conn)
	log.Println("DONE: client-peer copy")
}

func (s *Server) Run() {
	for {
		conn, isPeer, err := s.nextConn()
		if err != nil {
			log.Println(err)
			continue
		}

		if isPeer {
			s.pool.Peers.Add(Peer{Quota: 0, Conn: conn.conn})
			log.Println("Peer added", conn.conn.RemoteAddr(), "- Total:", s.pool.Peers.Size())
			log.Println("DEBUG: clients:", s.pool.Clients.Size())
			s.pool.Clients.ForEach(s.handleClient)
			continue
		}

		go s.handleClient(conn)
	}
}

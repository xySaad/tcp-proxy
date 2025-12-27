package server

import (
	"01proxy/model"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
)

type Pool struct {
	Clients   model.Pool[Client]
	Peers     model.Pool[*Peer]
	TunnelMap model.Map[uint64, chan net.Conn]
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
		pool: Pool{TunnelMap: model.NewTunnelMap[uint64, chan net.Conn]()},
		ln:   ln,
	}, nil
}

func (s *Server) handleClient(client Client) {
	peer := s.pool.NextPeer()
	if peer == nil {
		s.pool.Clients.Add(client)
		return
	}
	log.Println("using peer", peer.Conn.RemoteAddr())
	peer.Mx.Lock()
	defer peer.Mx.Unlock()

	var id uint64 = uint64(rand.Int63())
	ch := make(chan net.Conn)
	s.pool.TunnelMap.Set(id, ch)
	go func() {
		tunnel := <-ch
		if tunnel == nil {
			log.Println("tunnel with id", id, "is null")
			return
		}
		s.pool.TunnelMap.Delete(id)
		fmt.Printf("writing buffer to client %s: %s\n", client.Conn.RemoteAddr(), string(client.buffer))
		tunnel.Write(client.buffer)
		client.buffer = nil
		fmt.Printf("copying %s <=> %s\n", tunnel.RemoteAddr(), client.Conn.RemoteAddr())

		err := model.BiCopy(tunnel, client.Conn)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Successfully Copied %s <=> %s\n", tunnel.RemoteAddr(), client.Conn.RemoteAddr())
	}()

	err := peer.StartBridge(id)
	if err == nil {
		return
	}
	log.Println("Bridge rejected, Reason:", err)
	log.Println("cleanup peer and tunnel", id, "...")
	peer.Conn.Close()
	s.pool.Peers.RemoveBy(func(p *Peer) bool {
		return p.Conn == peer.Conn
	})
	ch <- nil
	s.handleClient(client)
}

func (s *Server) Run() {
	for {
		rawConn, err := s.ln.Accept()
		if err != nil {
			log.Println("ERROR at ln.Accept:", err)
			continue
		}

		go func() {
			conn, err := s.nextConn(rawConn)
			if err != nil {
				log.Println("Error processing conn header", err)
				return
			}

			switch conn := conn.(type) {
			case Peer:
				s.handlePeer(&conn)
			case Tunnel:
				s.handleTunnel(conn)
			case Client:
				s.handleClient(conn)
			}
		}()
	}
}

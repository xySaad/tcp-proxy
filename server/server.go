package server

import (
	"01proxy/model"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"slices"
	"time"
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
	peer := s.pool.LockedNextPeer()
	if peer == nil {
		s.pool.Clients.Add(client)
		return
	}
	log.Println("using peer", peer.Conn.RemoteAddr())
	defer func() {
		peer.Mx.Unlock()
	}()

	var id uint64 = uint64(rand.Int63())
	ch := make(chan net.Conn)
	s.pool.TunnelMap.Set(id, ch)
	go func() {
		defer func() {
			peer.Mx.Lock()
			peer.Quota--
			peer.Mx.Unlock()
			if waitingClient, ok := s.pool.Clients.Pop(); ok {
				log.Println("Quota freed, handling queued client")
				go s.handleClient(waitingClient)
			}
		}()

		tunnel := <-ch
		if tunnel == nil {
			log.Println("tunnel with id", id, "is null")
			return
		}
		s.pool.TunnelMap.Delete(id)
		defer client.Conn.Close()
		log.Printf("writing buffer to client %s: %s\n", client.Conn.RemoteAddr(), string(client.buffer))
		tunnel.Write(client.buffer)
		client.buffer = nil

		log.Printf("copying %s <=> %s\n", tunnel.RemoteAddr(), client.Conn.RemoteAddr())
		err := model.BiCopy(tunnel, client.Conn)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Successfully Copied %s <=> %s\n", tunnel.RemoteAddr(), client.Conn.RemoteAddr())
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
	go s.Dispenser()
	for {
		rawConn, err := s.ln.Accept()
		if err != nil {
			log.Println("ERROR at ln.Accept:", err)
			continue
		}
		timeoutConn := TimeoutConn{Conn: rawConn, Timeout: time.Second * 5}

		go func() {
			conn, err := s.nextConn(&timeoutConn)
			if err != nil {
				rawConn.Close()
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

func (s *Server) Dispenser() {
	for {
		<-time.Tick(5 * time.Second)
		deadPeers := []*Peer{}

		for _, p := range s.pool.Peers.Value() {
			if !p.Mx.TryLock() {
				continue
			}

			err := model.WriteCommand(p.Conn, model.KEEP_ALIVE())
			if err != nil {
				deadPeers = append(deadPeers, p)
				p.Conn.Close()
				p.Mx.Unlock()
				continue
			}
			_, err = io.ReadFull(p.Conn, model.KEEP_ALIVE())
			if err != nil {
				deadPeers = append(deadPeers, p)
				p.Conn.Close()
				p.Mx.Unlock()
				continue
			}
			p.Mx.Unlock()
		}

		if len(deadPeers) > 0 {
			s.pool.Peers.RemoveBy(func(p *Peer) bool {
				log.Println("dead peer removed", p.Conn.RemoteAddr())
				return slices.Contains(deadPeers, p)
			})
		}
	}
}

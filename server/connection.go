package server

import (
	"01proxy/model"
	"bytes"
	"io"
	"net"
)

type Client struct {
	buffer []byte
	Conn   net.Conn
}

func (s *Server) nextConn() (any, error) {
	conn, err := s.ln.Accept()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, len(model.PEER_REQUEST))
	_, err = io.ReadFull(conn, buffer)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(buffer, model.PEER_REQUEST) {
		_, err = conn.Write(model.PEER_ACCEPTED)
		if err != nil {
			return nil, err
		}

		return Peer{
			Conn: conn,
		}, nil
	}

	buffer2 := make([]byte, len(model.TUNNEL_REQUEST)-len(buffer))
	_, err = io.ReadFull(conn, buffer2)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(append(buffer, buffer2...), model.TUNNEL_REQUEST) {
		_, err = conn.Write(model.TUNNEL_ACCEPTED)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	return Client{
		buffer: append(buffer, buffer2...),
		Conn:   conn,
	}, nil
}

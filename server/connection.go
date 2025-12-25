package server

import (
	"01proxy/model"
	"bytes"
	"io"
	"net"
)

type Client struct {
	buffer []byte
	conn   net.Conn
}

func (s *Server) nextConn() (*Client, bool, error) {
	conn, err := s.ln.Accept()
	if err != nil {
		return nil, false, err
	}

	buffer := make([]byte, len(model.PEER_REQUEST))
	_, err = io.ReadFull(conn, buffer)
	if err != nil {
		return nil, false, err
	}

	if !bytes.Equal(buffer, model.PEER_REQUEST) {
		return &Client{
			buffer: buffer,
			conn:   conn,
		}, false, nil
	}

	_, err = conn.Write(model.PEER_ACCEPTED)
	if err != nil {
		return nil, false, err
	}
	return &Client{
		buffer: nil,
		conn:   conn,
	}, true, nil
}

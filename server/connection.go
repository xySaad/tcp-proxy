package server

import (
	"01proxy/model"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"
)

type TimeoutConn struct {
	net.Conn
	Timeout time.Duration
}

func (c *TimeoutConn) Read(b []byte) (n int, err error) {
	err = c.Conn.SetReadDeadline(time.Now().Add(c.Timeout))
	defer c.Conn.SetReadDeadline(time.Time{})
	if err != nil {
		return
	}
	return c.Conn.Read(b)
}
func (c *TimeoutConn) Write(b []byte) (n int, err error) {
	err = c.Conn.SetWriteDeadline(time.Now().Add(c.Timeout))
	defer c.Conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return
	}
	return c.Conn.Write(b)
}

func (s *Server) nextConn(conn net.Conn) (any, error) {
	log.Println("connection accepted", conn.RemoteAddr())
	buffer, err := model.ReadExact(conn, model.PREFIX())
	if err != nil {
		return nil, err
	}
	if buffer != nil {
		log.Println(conn.RemoteAddr(), "connected as client with buffer", string(buffer))
		return Client{
			buffer: buffer,
			Conn:   conn,
		}, nil
	}

	command, err := model.ReadCommand(conn)
	if bytes.Equal(command, model.PEER_REQUEST()) {
		_, err = conn.Write(model.PEER_ACCEPTED())
		if err != nil {
			return nil, err
		}
		log.Println(conn.RemoteAddr(), "connected as peer")
		return Peer{
			Conn: conn,
		}, nil
	}

	if bytes.Equal(command, model.TUNNEL_REQUEST()) {
		log.Println(conn.RemoteAddr(), "connected as tunnel")
		idBuf := make([]byte, 8)
		if _, err := io.ReadFull(conn, idBuf); err != nil {
			return nil, err
		}
		id := binary.BigEndian.Uint64(idBuf)

		log.Println(conn.RemoteAddr(), "registred with id", id)
		return Tunnel{ID: id, Conn: conn}, nil
	}

	log.Println(conn.RemoteAddr(), "connected as client with buffer", string(append(buffer, command...)))
	return Client{
		buffer: append(buffer, command...),
		Conn:   conn,
	}, nil
}

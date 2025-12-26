package server

import (
	"net"
)

type Client struct {
	buffer []byte
	Conn   net.Conn
}

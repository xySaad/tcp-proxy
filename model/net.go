package model

import (
	"io"
	"net"
	"net/http"
	"time"
)

func BiCopy(live net.Conn, killable net.Conn) {
	done := make(chan struct{})
	go func() {
		io.Copy(killable, live)
		live.SetReadDeadline(time.Time{})
		close(done)
	}()

	io.Copy(live, killable)
	killable.Close()
	live.SetReadDeadline(time.Now())
	<-done
}

func ServerListen(s *http.Server) (net.Listener, *net.TCPAddr, error) {
	addr := s.Addr

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	return ln, ln.Addr().(*net.TCPAddr), nil
}

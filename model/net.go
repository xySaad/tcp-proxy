package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func BiCopy(live net.Conn, killable net.Conn) error {
	var err1 error
	go func() {
		_, err1 = io.Copy(killable, live)
		live.SetDeadline(time.Time{})
		if err1 != nil {
			err1 = fmt.Errorf("Error on copy %s <=> %s: %s", killable.RemoteAddr(), live.RemoteAddr(), err1)
		}

		if !os.IsTimeout(err1) {
			killable.SetDeadline(time.Now())
		}
	}()

	_, err2 := io.Copy(live, killable)
	killable.SetDeadline(time.Time{})
	if !os.IsTimeout(err2) {
		live.SetDeadline(time.Now())
	} else if err1 != nil {
		return err1
	}

	if err2 != nil {
		return fmt.Errorf("Error on copy %s <=> %s: %s", live.RemoteAddr(), killable.RemoteAddr(), err2)
	}

	return nil
}

func ServerListen(s *http.Server) (net.Listener, *net.TCPAddr, error) {
	addr := s.Addr

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	return ln, ln.Addr().(*net.TCPAddr), nil
}

func WriteHeader(conn io.Writer, command []byte) (int, error) {
	totalLen := len(PREFIX) + 2 + len(command)

	buf := make([]byte, totalLen)

	copy(buf[:], PREFIX)
	binary.BigEndian.PutUint16(buf[len(PREFIX):], uint16(len(command)))
	copy(buf[len(PREFIX)+2:], command)
	return conn.Write(buf)
}

func ReadCommand(conn io.Reader) ([]byte, error) {
	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint16(lenBuf)

	commandBuf := make([]byte, length)
	if _, err := io.ReadFull(conn, commandBuf); err != nil {
		return nil, err
	}

	return commandBuf, nil
}

func ReadExact(conn io.Reader, expected []byte) ([]byte, error) {
	expectedBuf := make([]byte, len(expected))
	for total := 0; total < len(expectedBuf); {
		n, err := conn.Read(expectedBuf[total:])
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(expectedBuf[:n], expected[total:]) {
			return expectedBuf, nil
		}
		total += n
	}

	return nil, nil
}

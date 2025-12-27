package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type copyResult struct {
	err      error
	dst, src net.Conn
}

func BiCopy(a, b net.Conn) error {
	erchan := make(chan copyResult, 2)
	go func() {
		_, err := io.Copy(b, a)
		erchan <- copyResult{err: err, dst: b, src: a}
		b.SetDeadline(time.Now())
	}()

	go func() {
		_, err := io.Copy(a, b)
		erchan <- copyResult{err: err, dst: a, src: b}
		a.SetDeadline(time.Now())
	}()

	defer a.SetDeadline(time.Time{})
	defer b.SetDeadline(time.Time{})

	err := <-erchan
	<-erchan
	if err.err != nil {
		return fmt.Errorf("Error on copy %s => %s: %s", err.src.RemoteAddr(), err.dst.RemoteAddr(), err.err)
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
	totalLen := len(PREFIX()) + 2 + len(command)

	buf := make([]byte, totalLen)

	copy(buf[:], PREFIX())
	binary.BigEndian.PutUint16(buf[len(PREFIX()):], uint16(len(command)))
	copy(buf[len(PREFIX())+2:], command)
	return conn.Write(buf)
}

func WriteCommand(conn io.Writer, command []byte) error {
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(command)))

	if _, err := conn.Write(lenBuf); err != nil {
		return fmt.Errorf("failed to write length header: %w", err)
	}

	if _, err := conn.Write(command); err != nil {
		return fmt.Errorf("failed to write command body: %w", err)
	}

	return nil
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

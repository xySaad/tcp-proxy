package main

import (
	"01proxy/model"
	"bytes"
	"log"
	"net"
	"testing"
	"time"
)

func getCommand(command []byte) []byte {
	buf := bytes.NewBuffer(nil)
	err := model.WriteCommand(buf, command)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return buf.Bytes()
}

func TestReadExact(t *testing.T) {
	tests := [][][]byte{
		{model.PREFIX()}, // valid prefix but invalid command; should result in error while parsing command
		{model.PREFIX()[:5], model.PREFIX()[5:], getCommand(model.PEER_REQUEST())}, // valid peer; should connect as peer normally
		{[]byte("no idea")}} // connect as client

	for _, test := range tests {
		conn, err := net.Dial("tcp", ":1080")
		if err != nil {
			t.Fatal(err)
			return
		}

		for _, msg := range test {
			conn.Write([]byte(msg))
			time.Sleep(time.Second * 2)
		}
		conn.Close()
	}
}

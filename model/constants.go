package model

import "fmt"

const (
	_PEER_REQUEST    = "==PEER==REQUEST"
	_PEER_ACCEPTED   = "==PEER==ACCEPTED"
	_START_BRIDGE    = "==START==BRIDGE"
	_BRIDGE_REJECTED = "==BRIDGE==REJECTED"
	_BRIDGE_ACCEPTED = "==BRIDGE==ACCEPTED"
	_TUNNEL_REQUEST  = "==TUNNEL==REQUEST"
	_TUNNEL_ACCEPTED = "==TUNNEL==ACCEPTED"
	_TUNNEL_REJECTED = "==TUNNEL==REJECTED"
	_KEEP_ALIVE      = "==KEEP==ALIVE"
)

func KEEP_ALIVE() []byte {
	return []byte(_KEEP_ALIVE)
}
func PEER_REQUEST() []byte {
	return []byte(_PEER_REQUEST)
}
func PEER_ACCEPTED() []byte {
	return []byte(_PEER_ACCEPTED)
}
func START_BRIDGE() []byte {
	return []byte(_START_BRIDGE)
}
func BRIDGE_REJECTED() []byte {
	return []byte(_BRIDGE_REJECTED)
}
func BRIDGE_ACCEPTED() []byte {
	return []byte(_BRIDGE_ACCEPTED)
}
func TUNNEL_REQUEST() []byte {
	return []byte(_TUNNEL_REQUEST)
}
func TUNNEL_ACCEPTED() []byte {
	return []byte(_TUNNEL_ACCEPTED)
}
func TUNNEL_REJECTED() []byte {
	return []byte(_TUNNEL_REJECTED)
}

const VERSION_MAJOR byte = 0
const VERSION_MINOR byte = 1

var prefix = fmt.Sprintln("TCPRP", VERSION_MAJOR, VERSION_MINOR)

func PREFIX() []byte {
	return []byte(prefix)
}

const MAX_PEER_QUOTA = 5

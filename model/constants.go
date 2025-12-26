package model

var (
	PEER_REQUEST    = []byte("==PEER==REQUEST")
	PEER_ACCEPTED   = []byte("==PEER==ACCEPTED")
	START_BRIDGE    = []byte("==START==BRIDGE")
	BRIDGE_REJECTED = []byte("==BRIDGE==REJECTED")
	BRIDGE_ACCEPTED = []byte("==BRIDGE==ACCEPTED")

	TUNNEL_REQUEST  = []byte("==TUNNEL==REQUEST")
	TUNNEL_ACCEPTED = []byte("==TUNNEL==ACCEPTED")
	TUNNEL_REJECTED = []byte("==TUNNEL==REJECTED")
)

var PROTOCOL = [...]byte{'T', 'C', 'P', 'R', 'P'}

const VERSION_MAJOR byte = 0
const VERSION_MINOR byte = 1

var PREFIX = append(PROTOCOL[:], ' ', VERSION_MAJOR, VERSION_MINOR, '\n')

const MAX_PEER_QUOTA = 5

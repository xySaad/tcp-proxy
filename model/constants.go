package model

var (
	PEER_REQUEST    = []byte("==PEER==")
	PEER_ACCEPTED   = []byte("==PEER==ACCEPTED==")
	START_BRIDGE    = []byte("==START==BRIDGE")
	BRIDGE_REJECTED = []byte("==BRIDGE==REJECTED")
	BRIDGE_ACCEPTED = []byte("==BRIDGE==ACCEPTED")
)

const (
	MAX_PEER_QUOTA = 5
)

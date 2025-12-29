package constants

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

func PREFIX() []byte {
	return []byte(prefix)
}

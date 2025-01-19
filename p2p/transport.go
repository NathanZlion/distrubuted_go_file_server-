package p2p

import "net"

// represents a remote node
type Peer interface {
	net.Conn
	Send([]byte) error
}

// represents a means to handle communication between two nodes
// this can be TCP, UDP, websockets ...
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	Dial(string) error
	ListenAddr() net.Addr
}

package p2p

// represents a remote node
type Peer interface{}

// represents a means to handle communication between two nodes
// this can be TCP, UDP, websockets ...
type Transport interface {
	ListenAndAccept() error
}

package p2p

import (
	"fmt"
	"net"
	"sync"
)

// a peer connected to us with TCP
type TCPPeer struct {
	// underlying connection of peer
	conn net.Conn
	// If we initiated the connection it's outbound
	// if accepting an incoming connection it's inbound
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddress string
	ShakeHands    HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	mu       sync.RWMutex // protects the peer communication
	peers    map[net.Addr]Peer
	/**
	type Addr interface
	{
		Network() string // name of the network (for example, "tcp", "udp")
		String() string  // string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
	}
	*/
}

func NewTcpTransport(otps TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: otps,
	}
}

// listen start funciton
func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)

	if err != nil {
		return err
	}

	fmt.Printf("Started Listening %v\n\n", t.ListenAddress)

	go t.startAcceptLoop()

	return nil
}

// the connection accept loop
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if err != nil {
			fmt.Println("TCP: accept listener error")
		} else {
			go t.handleConn(conn)
		}
	}
}

// handle incoming connection requests
func (t *TCPTransport) handleConn(conn net.Conn) {
	// add it to the peers map
	peer := NewTCPPeer(conn, false)

	// try the handshake, if not successfull say goodbye to the peer
	if err := t.ShakeHands(peer); err != nil {
		peer.conn.Close()
		fmt.Printf("TCP Handshake error : %s\n", err)
		return
	}

	fmt.Printf("Incoming connection %+v\n", peer)

	msg := &Message{}
	for {
		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP decoding error %s\n", err)
			continue
		}

		msg.From = conn.RemoteAddr()
		fmt.Printf("Message: %+v \n", msg)
	}
}

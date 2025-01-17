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

func (peer *TCPPeer) Close() error {
	err := peer.conn.Close()
	return err
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddress string
	HandShakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcChan  chan RPC

	mu    sync.RWMutex // protects the peer communication
	peers map[net.Addr]Peer
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
		rpcChan:          make(chan RPC),
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

// the <- here makes the channel read only form this method
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
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
	var err error

	defer func() {
		fmt.Printf("Peer connection dropped: %s \n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, false)

	// try the handshake, if not successfull say goodbye to the peer
	if err = t.HandShakeFunc(peer); err != nil {
		return
	}

	// add it to the peers map
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	fmt.Printf("Incoming connection %+v\n", peer)

	rpc := RPC{}
	for {
		err := t.Decoder.Decode(conn, &rpc)

		if err == net.ErrClosed {
			fmt.Printf("TCP Network Closed %s\n", err)
			return
		}

		if err != nil {
			// fmt.Println(reflect.TypeOf(err))
			// fmt.Printf("TCP Read error %s\n", err)
			// continue
			return
		}

		rpc.From = conn.RemoteAddr()
		t.rpcChan <- rpc
	}
}

package p2p

import (
	"fmt"
	"net"
	"sync"
)

// a peer connected to us with TCP
type TCPPeer struct {
	// underlying connection of peer
	net.Conn
	// If we initiated the connection it's outbound
	// if accepting an incoming connection it's inbound
	outbound bool

	Wg *sync.WaitGroup
}

func DefaultOnPeer(peer Peer) error {
	return nil
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

type TCPTransportOpts struct {
	ListenAddress string
	HandShakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(peer Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcChan  chan RPC
	mu       sync.RWMutex
	peers    map[net.Addr]Peer
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

	go t.startAcceptLoop()

	fmt.Printf("Started Listening on port %v\n", t.ListenAddress)
	return nil
}

func (t *TCPTransport) ListenAddr() net.Addr {
	return t.listener.Addr()
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
			return
		} else {
			go t.handleConn(conn, false)
		}
	}
}

// handle incoming connection requests
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("Peer connection dropped\n")
		if err != nil {
			fmt.Printf("Error: %s \n", err)
		}
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	// try the handshake, if not successfull say goodbye to the peer
	if err = t.HandShakeFunc(peer); err != nil {
		err = err
		return
	}

	// add it to the peers map
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			err = err
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
			return
		}

		rpc.From = conn.RemoteAddr().String()

		// read key and start the wait group
		peer.Wg.Add(1)
		fmt.Println("Waitgroup blocked waiting for streaming of file to be done")
		t.rpcChan <- rpc
		peer.Wg.Wait()
		fmt.Println("Done waiting")
	}
}

func (t *TCPTransport) Close() error {
	err := t.listener.Close()
	// for peer := range t.peers {
	// 	pee := t.peers[peer]
	// 	pee.Close()
	// }
	return err
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Write(b)
	return err
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

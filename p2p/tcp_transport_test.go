package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTcpTransport(t *testing.T) {
	listenAddress := ":4000"

	tcpTransportOpts := TCPTransportOpts{
		ListenAddress: listenAddress,
		HandShakeFunc: NopHandShakeFunc,
		Decoder:       DefaultDecoder{},
	}

	tcpTransport := NewTcpTransport(tcpTransportOpts)

	assert.Equal(t, tcpTransport.ListenAddress, listenAddress)
}

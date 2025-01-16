package main

import (
	"log"

	"github.com/NathanZlion/distruted_go_file_server-/p2p"
)

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: ":4000",
		ShakeHands:    p2p.NopHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTcpTransport(tcpTransportOpts)

	if err := tcpTransport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}

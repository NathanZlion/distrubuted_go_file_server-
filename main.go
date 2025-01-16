package main

import (
	"errors"
	"fmt"
	"github.com/NathanZlion/distruted_go_file_server-/p2p"
	"log"
	"math/rand"
)

func OnPeer(peer p2p.Peer) error {
	if randomNum := rand.Intn(3); randomNum >= 1 {
		return errors.New("Peer Blocked")
	}
	return nil
}

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: ":4000",
		HandShakeFunc: p2p.NopHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}

	tcpTransport := p2p.NewTcpTransport(tcpTransportOpts)

	if err := tcpTransport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			message := <-tcpTransport.Consume()
			fmt.Printf("Message: %+v \n", message)
		}
	}()

	select {}
}

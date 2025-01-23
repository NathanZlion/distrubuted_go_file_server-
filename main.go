package main

import (
	// "bytes"
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/NathanZlion/distruted_go_file_server-/p2p"
)

func makeServer(listenAddr, storageRoot string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandShakeFunc: p2p.NopHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTcpTransport(tcpTransportOpts)

	fileServerOpts := FileServeropts{
		ListenAddr:           listenAddr,
		StorageRoot:          storageRoot,
		PathTransformFunc:    CASPathTransformFunc,
		Transport:            tcpTransport,
		BootstrapServersList: nodes,
	}

	fs := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = fs.OnPeer
	return fs
}

func main() {
	server1 := makeServer(":3000", "3000_network_storage")
	server2 := makeServer(":4000", "4000_network_storage", ":3000")

	go func() {
		if err := server1.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(time.Second * 1)

	go server2.Start()

	time.Sleep(time.Second * 1)

	data := bytes.NewReader([]byte("Gugu Gaga Iglabo ulala"))
	server1.Store("My Data", data)

	time.Sleep(time.Second * 2)

	r, err := server2.Get("My Data")
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Read Data, %s Error %+v \n", string(b))

	select {}
}

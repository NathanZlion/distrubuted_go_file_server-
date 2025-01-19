package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/NathanZlion/distruted_go_file_server-/p2p"
)

type FileServeropts struct {
	ListenAddr           string
	StorageRoot          string
	PathTransformFunc    PathTransformFunc
	Transport            p2p.Transport
	BootstrapServersList []string
}

type FileServer struct {
	FileServeropts
	store  *Store
	quitch chan struct{}

	peerlock sync.Mutex
	peers    map[string]p2p.Peer
}

func NewFileServer(opts FileServeropts) *FileServer {
	storeOpts := StoreOpts{
		opts.StorageRoot,
		opts.PathTransformFunc,
	}

	return &FileServer{
		FileServeropts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key      string
	ByteSize int
}

func (s *FileServer) Broadcast(message *Message) error {
	s.peerlock.Lock()
	defer s.peerlock.Unlock()

	peers := make([]io.Writer, 0, len(s.peers))

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	fmt.Printf("Broadcasting file to %d peers: %+v, Message: %+v\n", len(peers), peers, message)

	return gob.NewEncoder(io.MultiWriter(peers...)).Encode(message)
}

func (s *FileServer) StoreFile(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	size, err := s.store.WriteStream(key, tee)

	if err != nil {
		return err
	}

	msgBuf := new(bytes.Buffer)

	keyMsg := Message{
		Payload: MessageStoreFile{
			Key:      key,
			ByteSize: size,
		},
	}

	if err := gob.NewEncoder(msgBuf).Encode(keyMsg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 1)

	// send a large file
	for _, peer := range s.peers {
		n, err := io.Copy(peer, buf)
		if err != nil {
			fmt.Printf("Written %d bytes \n", n)
		}
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.BootstrapServer()
	s.Loop()

	return nil
}

func (s *FileServer) Loop() {
	defer func() {
		s.Transport.Close()
		err := recover()
		fmt.Printf("Stopped Server loop due to %v \n", err)
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var message Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&message); err != nil {
				log.Printf("Error while decoding Incoming message.\nBytes: %+v Error: %+v.\n", rpc.Payload, err)
				continue
			}

			if err := s.HandleMessage(rpc.From, &message); err != nil {
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) HandleMessage(from string, msg *Message) error {
	switch t := msg.Payload.(type) {
	case MessageStoreFile:
		s.HandleMessageFileStore(from, &t)
	default:
		fmt.Printf("Data type: ")
	}

	return nil
}

func (s *FileServer) HandleMessageFileStore(from string, msg *MessageStoreFile) error {
	peer, ok := s.peers[from]
	defer peer.(*p2p.TCPPeer).Wg.Done()

	if !ok {
		return fmt.Errorf("Peer %s not found in peers map", peer)
	}

	_, err := s.store.WriteStream(msg.Key, io.LimitReader(peer, int64(msg.ByteSize)))
	return err
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(peer p2p.Peer) error {
	s.peerlock.Lock()
	defer s.peerlock.Unlock()

	s.peers[peer.RemoteAddr().String()] = peer

	return nil
}

func (s *FileServer) BootstrapServer() error {
	for _, serverAddr := range s.BootstrapServersList {
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				fmt.Println(err)
			}
		}(serverAddr)
	}

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}

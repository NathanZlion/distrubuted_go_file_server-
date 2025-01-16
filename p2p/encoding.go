package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

// for plain decoding with no GOB
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *Message) error {
	buff := make([]byte, 2000)
	n, err := r.Read(buff)

	if err != nil {
		return err
	}

	// write it to the message payload
	msg.Payload = buff[:n]

	return nil
}

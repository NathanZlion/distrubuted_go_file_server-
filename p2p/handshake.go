package p2p


type HandshakeFunc func(any) error

func NopHandShakeFunc(any) error { return nil }

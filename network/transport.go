package network

import "io"

type NetAddr string

// Network implements net.Addr.
func (n NetAddr) Network() string {
	panic("unimplemented")
}

// String implements net.Addr.
func (n NetAddr) String() string {
	panic("unimplemented")
}

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type Transport interface {
	Consume() <-chan RPC
	Connect(transport Transport) error
	SendMessage(to NetAddr, payload []byte) error
	Addr() NetAddr
}

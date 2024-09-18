package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
)


type TCPPeer struct {
	conn net.Conn
	Outgoing bool
}

// 接收连接对象
type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	peerCh chan *TCPPeer
}

func (p *TCPPeer) Send(payload []byte) error {
	_, err := p.conn.Write(payload)
	return err
}

func (p *TCPPeer) ReceiveLoop(rpcCh chan<- RPC) {
	defer p.conn.Close()

	for {
		buf := make([]byte, 4096)
		n, err := p.conn.Read(buf)
		if err == io.EOF {
			continue
		}
		if err != nil {
			continue
		}

		rpc := RPC{
			From:    NetAddr(p.conn.RemoteAddr().String()),
			Payload: bytes.NewReader(buf[:n]),
		}

		rpcCh <- rpc
	}
}



func NewTCPPeer(conn net.Conn, outgoing bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		Outgoing: outgoing,
	}
}

func NewTCPTransport(listenAddr string) (*TCPTransport, error) {
	return &TCPTransport{
		listenAddr: listenAddr,
		peerCh:     make(chan *TCPPeer),
	}, nil
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln
	go t.acceptLoop()

	return nil
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("接受连接时发生错误: %s\n", err)
			continue
		}

		peer := NewTCPPeer(conn, false)
		t.peerCh <- peer
	}
}



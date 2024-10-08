package network

import (
	"bytes"
	"encoding/gob"
	"go-chain/core"
	"go-chain/inter"
	"go-chain/utils"
	"io"
)

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x1
	MessageTypeBlock     MessageType = 0x2
	MessageTypeGetBlocks MessageType = 0x3
	MessageTypeBlocks    MessageType = 0x4
	MessageTypeGetStatus MessageType = 0x5
	MessageTypeStatus    MessageType = 0x6
	MessageTypeGetPeers  MessageType = 0x7
	MessageTypePeers     MessageType = 0x8
)

type Message struct {
	Type MessageType
	Body []byte
}

type GetBlocksMessage struct {
	From uint32
	To   uint32
}

// BlocksMessage 表示区块消息
type BlocksMessage struct {
	Blocks []*core.Block
}

func (bm BlocksMessage) FirstBlock() *core.Block {
	if len(bm.Blocks) == 0 {
		return nil
	}
	return bm.Blocks[0]
}

// LastBlock 返回区块消息中的最后一个区块
func (bm BlocksMessage) LastBlock() *core.Block {
	if len(bm.Blocks) == 0 {
		return nil
	}
	return bm.Blocks[len(bm.Blocks)-1]
}


// GetStatusMessage 表示获取状态消息
type GetStatusMessage struct {
	// 获取状态消息可能不需要额外字段
}

// StatusMessage 表示状态消息
type StatusMessage struct {
	// id of server
	ID            string
	Version       uint32
	CurrentHeight uint32
}

type GetPeersMessage struct {
}

type PeersMessage struct {
	Peers []string
}


var _ inter.Codable = new(GetBlocksMessage)
var _ inter.Codable = new(BlocksMessage)
var _ inter.Codable = new(GetStatusMessage)
var _ inter.Codable = new(StatusMessage)
var _ inter.Codable = new(GetBlocksMessage)
var _ inter.Codable = new(GetPeersMessage)
var _ inter.Codable = new(PeersMessage)

// 为每种消息类型实现 Encode 和 Decode 方法
func (m *GetBlocksMessage) Encode(w io.Writer) error {
	return utils.EncodeMessage(m, w)
}

func (m *GetBlocksMessage) Decode(r io.Reader) error {
	return utils.DecodeMessage(m, r)
}

func (m *BlocksMessage) Encode(w io.Writer) error {
	return utils.EncodeMessage(m, w)
}

func (m *BlocksMessage) Decode(r io.Reader) error {
	return utils.DecodeMessage(m, r)
}

func (m *GetStatusMessage) Encode(w io.Writer) error {
	return utils.EncodeMessage(m, w)
}

func (m *GetStatusMessage) Decode(r io.Reader) error {
	return utils.DecodeMessage(m, r)
}

func (m *StatusMessage) Encode(w io.Writer) error {
	return utils.EncodeMessage(m, w)
}

func (m *StatusMessage) Decode(r io.Reader) error {
	return utils.DecodeMessage(m, r)
}

func (m *GetPeersMessage) Encode(w io.Writer) error {
	return utils.EncodeMessage(m, w)
}

func (m *GetPeersMessage) Decode(r io.Reader) error {
	return utils.DecodeMessage(m, r)
}

func (m *PeersMessage) Encode(w io.Writer) error {
	return utils.EncodeMessage(m, w)
}

func (m *PeersMessage) Decode(r io.Reader) error {
	return utils.DecodeMessage(m, r)
}

func EncodeMessage(t MessageType, c inter.Codable) ([]byte, error) {
	var b bytes.Buffer
	c.Encode(&b)
	msg := Message{
		Type: t,
		Body: b.Bytes(),
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
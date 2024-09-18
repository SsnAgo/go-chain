package core

import (
	"bytes"
	"go-chain/types"
	"go-chain/utils"
	"io"
	"time"
)

type BlockHeader struct {
	Version       uint32
	PrevBlockHash types.Hash
	DataHash      types.Hash
	Height        uint32
	Timestamp     int64
	// nonce表示的是这个块的工作量 即矿工挖到的nonce
	Nonce uint32
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
}

func (b *Block) Encode(w io.Writer) error {
	return utils.EncodeMessage(b, w)
}
func (b *Block) Decode(r io.Reader) error {
	return utils.DecodeMessage(b, r)
}

// NewBlock 创建一个新的区块
// 创建区块的时候 同时需要打包交易数据 计算DataHash
// 打包的交易从哪里来  从一个tx Pool里面获取
func NewBlock(prevBlockHash types.Hash, height uint32, transactions []*Transaction) *Block {
	header := &BlockHeader{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Timestamp:     time.Now().Unix(),
	}

	block := &Block{
		Header:       header,
		Transactions: transactions,
	}

	// todo 打包tx

	// 计算交易的DataHash
	dataHash, err := block.CalculateDataHash()
	if err != nil {
		// 处理错误，这里简单地返回nil
		return nil
	}
	header.DataHash = dataHash

	return block
}

// CalculateDataHash 计算区块中交易数据的哈希值
func (b *Block) CalculateDataHash() (types.Hash, error) {
	buf := &bytes.Buffer{}
	for _, tx := range b.Transactions {
		if err := tx.Encode(buf); err != nil {
			return types.Hash{}, err
		}
	}
	return types.HashFromBytes(utils.SHA256(buf.Bytes())), nil
}

func (b *Block) GetDataHash() types.Hash {
	return b.Header.DataHash
}

func (b *Block) GetPrevBlockHash() types.Hash {
	return b.Header.PrevBlockHash
}

// Verify 验证区块的有效性
func (b *Block) Verify() bool {
	// 验证区块头
	if !b.verifyHeader() {
		return false
	}

	// 验证交易
	if !b.verifyTransactions() {
		return false
	}

	// 验证DataHash
	if !b.verifyDataHash() {
		return false
	}

	return true
}

// verifyHeader 验证区块头
func (b *Block) verifyHeader() bool {
	// 验证版本号
	if b.Header.Version != 1 {
		return false
	}

	// 验证时间戳
	if b.Header.Timestamp > time.Now().Unix() {
		return false
	}

	// 可以添加更多的头部验证逻辑

	return true
}

// verifyTransactions 验证区块中的所有交易
func (b *Block) verifyTransactions() bool {
	for _, tx := range b.Transactions {
		if !tx.Verify() {
			return false
		}
	}
	return true
}

// verifyDataHash 验证DataHash是否正确
func (b *Block) verifyDataHash() bool {
	calculatedHash, err := b.CalculateDataHash()
	if err != nil {
		return false
	}
	return b.Header.DataHash == calculatedHash
}

func (b *Block) PreOf(nxt *Block) bool {
	return nxt.GetPrevBlockHash() == b.GetDataHash()
}

func (b *Block) Height() uint32 {
	return b.Header.Height
}

func (b *Block) Timestamp() int64 {
	return b.Header.Timestamp
}

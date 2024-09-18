package core

import (
	"bytes"
	"encoding/binary"
	"go-chain/cryptoo"
	"go-chain/types"
	"go-chain/utils"
	"io"
)

type Transaction struct {
	// 交易本身数据
	From cryptoo.PublicKey
	To   cryptoo.PublicKey
	Data []byte
	Value uint64
	Nonce int64

	// tx data hash
	Hash types.Hash
	// 交易签名
	Signature  *cryptoo.Signature
}

func (t *Transaction) setDigest() {
	t.Hash = t.CalHash()
}
func (t *Transaction) CalHash() types.Hash {
	b := &bytes.Buffer{}
	binary.Write(b, binary.LittleEndian, t.From)
	binary.Write(b, binary.LittleEndian, t.To)
	binary.Write(b, binary.LittleEndian, t.Data)
	binary.Write(b, binary.LittleEndian, t.Value)
	binary.Write(b, binary.LittleEndian, t.Nonce)
	return types.Hash(utils.SHA256(b.Bytes()))
}

func (t *Transaction) GetHash() types.Hash {
	return t.Hash
}

func (t *Transaction) sign(priv *cryptoo.PrivateKey) {
	// 计算交易数据的Hash值
	t.setDigest()

	// 使用cryptoo包中的结构体和方法进行签名
	signature, err := priv.Sign(t.Hash[:])
	if err != nil {
		// 处理签名错误
		return
	}
	
	// 将签名赋值给交易的Signature字段
	t.Signature = signature
}

func (t *Transaction) Encode(w io.Writer) error {
	return utils.EncodeMessage(t, w)
}

func (t *Transaction) Decode(r io.Reader) error {
	return utils.DecodeMessage(t, r)
}

func NewTransaction(signerPriv *cryptoo.PrivateKey, to cryptoo.PublicKey, data []byte, value uint64, nonce int64) *Transaction {
	tx := &Transaction{
		From: signerPriv.GetPublicKey(),
		To:   to,
		Data: data,
		Value: value,
		Nonce: nonce,
	}
	// 生成Hash
	tx.sign(signerPriv)
	return tx
}

func (t *Transaction) Verify() bool {
	// 验证交易签名
	if t.Signature == nil {
		return false
	}

	// 重新计算交易的摘要
	t.setDigest()

	// 使用公钥验证签名
	return t.Signature.Verify(t.From, t.Hash[:])
}

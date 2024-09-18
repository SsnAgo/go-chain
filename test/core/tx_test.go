package test

import (
	"go-chain/core"
	"go-chain/cryptoo"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	// 生成测试用的私钥和公钥
	fromPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPubKey := toPrivKey.GetPublicKey()

	// 创建一个新的交易
	data := []byte("测试数据")
	value := uint64(100)
	nonce := int64(1)
	tx := core.NewTransaction(fromPrivKey, toPubKey, data, value, nonce)

	// 验证交易的各个字段
	assert.Equal(t, fromPrivKey.GetPublicKey(), tx.From)
	assert.Equal(t, toPubKey, tx.To)
	assert.Equal(t, data, tx.Data)
	assert.Equal(t, value, tx.Value)
	assert.Equal(t, nonce, tx.Nonce)
	assert.NotNil(t, tx.Signature)
}

func TestTransaction_CalHash(t *testing.T) {
	// 创建两个相同的交易
	fromPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPubKey := toPrivKey.GetPublicKey()
	
	tx1 := core.NewTransaction(fromPrivKey, toPubKey, []byte("测试数据"), 100, 1)
	tx2 := core.NewTransaction(fromPrivKey, toPubKey, []byte("测试数据"), 100, 1)

	// 验证两个交易的哈希值相同
	assert.Equal(t, tx1.GetHash(), tx2.GetHash())

	// 修改交易的某个字段，验证哈希值变化
	tx2.Value = 200
	newHahs := tx2.CalHash()
	assert.NotEqual(t, tx1.CalHash(), newHahs)
}

func TestTransaction_Sign(t *testing.T) {
	fromPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPubKey := toPrivKey.GetPublicKey()

	tx := core.NewTransaction(fromPrivKey, toPubKey, []byte("测试数据"), 100, 1)

	// 验证签名已经生成
	assert.NotNil(t, tx.Signature)

	// 验证签名是有效的
	
	isValid := tx.Verify()
	assert.True(t, isValid)

	// 修改交易数据，验证签名变为无效
	tx.Value = 200
	isValid = tx.Verify()
	assert.False(t, isValid)
}

func TestTransaction_VerifyTransaction(t *testing.T) {
	fromPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPrivKey, _ := cryptoo.GeneratePrivateKey()
	toPubKey := toPrivKey.GetPublicKey()

	tx := core.NewTransaction(fromPrivKey, toPubKey, []byte("测试数据"), 100, 1)

	// 验证有效交易
	assert.True(t, tx.Verify())

	// 修改交易数据，验证交易变为无效
	tx.Value = 200
	assert.False(t, tx.Verify())

	// 使用错误的私钥签名，验证交易无效
	wrongPrivKey, _ := cryptoo.GeneratePrivateKey()
	tx = core.NewTransaction(wrongPrivKey, toPubKey, []byte("测试数据"), 100, 1)
	tx.From = fromPrivKey.GetPublicKey() // 使用正确的发送方公钥
	assert.False(t, tx.Verify())
}


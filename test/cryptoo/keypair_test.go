package test

import (
	"fmt"
	"go-chain/cryptoo"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey, err := cryptoo.GeneratePrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, privateKey)
	fmt.Println(privateKey.String())
	assert.Equal(t, 64, len(privateKey.String()))
}

func TestPrivateKey_GetPublicKey(t *testing.T) {
	privateKey, _ := cryptoo.GeneratePrivateKey()
	publicKey := privateKey.GetPublicKey()
	assert.NotNil(t, publicKey)
	fmt.Println(publicKey.String())
	assert.Equal(t, 66, len(publicKey.String()))
}

func TestPrivateKey_Sign(t *testing.T) {
	privateKey, _ := cryptoo.GeneratePrivateKey()
	message := []byte("测试消息")
	signature, err := privateKey.Sign(message)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, 128, len(signature.String()))
}

func TestPublicKey_Verify(t *testing.T) {
	privateKey, _ := cryptoo.GeneratePrivateKey()
	publicKey := privateKey.GetPublicKey()
	message := []byte("测试消息")
	signature, _ := privateKey.Sign(message)
	isValid := signature.Verify(publicKey, message)
	assert.True(t, isValid)
	
	// 测试无效签名
	otherPrivateKey, _ := cryptoo.GeneratePrivateKey()
	invalidSignature, _ := otherPrivateKey.Sign(message)
	isValid = invalidSignature.Verify(publicKey, message)
	assert.False(t, isValid)
}

func TestIntegration_KeyPairSignAndVerify(t *testing.T) {
	// 生成密钥对
	privateKey, _ := cryptoo.GeneratePrivateKey()
	publicKey := privateKey.GetPublicKey()
	
	// 签名消息
	message := []byte("集成测试消息")
	signature, err := privateKey.Sign(message)
	assert.NoError(t, err)
	
	// 验证签名
	isValid := signature.Verify(publicKey, message)
	assert.True(t, isValid)
	
	// 修改消息后验证
	modifiedMessage := []byte("修改后的消息")
	isValid = signature.Verify(publicKey, modifiedMessage)
	assert.False(t, isValid)
}

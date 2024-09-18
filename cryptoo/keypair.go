package cryptoo

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go-chain/types"
	"go-chain/utils"
	"io"
	"math/big"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

type PublicKey []byte

type Signature struct {
	R *big.Int
	S *big.Int
}

// Sign 使用私钥对消息进行签名
func (pk *PrivateKey) Sign(data []byte) (*Signature, error) {
	// 对data生成固定长度的hash值 一般称作digest
	hash := utils.SHA256(data)
	r, s, err := ecdsa.Sign(rand.Reader, pk.key, hash)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %v", err)
	}
	return &Signature{
		R: r,
		S: s,
	}, nil
}

func (pk *PrivateKey) String() string {
	return hex.EncodeToString(pk.key.D.Bytes())
}

// GetPublicKey 获取对应的公钥
func (pk *PrivateKey) GetPublicKey() PublicKey {
	return elliptic.MarshalCompressed(pk.key.PublicKey.Curve, pk.key.PublicKey.X, pk.key.PublicKey.Y)
}


// NewPrivateKey 生成一个新的私钥
func newPrivateKeyFromReader(r io.Reader) (*PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}
	// 将私钥转换为16进制字符串
	privKeyHex := hex.EncodeToString(key.D.Bytes())
	fmt.Printf("私钥请保存好，不要泄露给任何人。\n私钥: %s\n", privKeyHex)
	return &PrivateKey{key: key}, nil
}

func GeneratePrivateKey() (*PrivateKey, error) {
	return newPrivateKeyFromReader(rand.Reader)
}

// GenerateAddress 根据公钥生成地址
func (pub PublicKey) Address() types.Address {
	hash := utils.SHA256(pub)
	address := hash[:20]
	return types.Address(address)
}

// String 返回公钥的十六进制字符串表示
func (pub PublicKey) String() string {
	return hex.EncodeToString(pub)
}


// String 返回签名的十六进制字符串表示
func (s *Signature) String() string {
	b := append(s.R.Bytes(), s.S.Bytes()...)
	return hex.EncodeToString(b)
}

// Verify 验证签名是否有效
func (s *Signature) Verify(publicKey PublicKey, data []byte) bool {
	// 解析公钥
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), publicKey)
	if x == nil {
		return false
	}
	
	// 重建公钥对象
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	
	// 计算消息的哈希值
	hash := utils.SHA256(data)
	
	// 验证签名
	return ecdsa.Verify(pubKey, hash, s.R, s.S)
}

func UseOrGenPrivateKey(privKeyStr string) (*PrivateKey, error) {
	if privKeyStr != "" {
		// 如果提供了私钥字符串，尝试解析它
		privKeyBytes, err := hex.DecodeString(privKeyStr)
		if err != nil {
			return nil, fmt.Errorf("解析私钥字符串失败: %v", err)
		}
		key, err := ecdsa.GenerateKey(elliptic.P256(), bytes.NewReader(privKeyBytes))
		if err != nil {
			return nil, fmt.Errorf("从私钥字符串生成密钥失败: %v", err)
		}
		return &PrivateKey{key: key}, nil
	}
	
	// 如果没有提供私钥字符串，生成新的私钥
	return GeneratePrivateKey()
}


// 




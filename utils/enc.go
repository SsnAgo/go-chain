package utils

import (
	"encoding/gob"
	"go-chain/inter"
	"io"
)

// 使用泛型实现通用的 Encode 和 Decode 函数
func EncodeMessage[T inter.Codable](m T, w io.Writer) error {
	return gob.NewEncoder(w).Encode(m)
}

func DecodeMessage[T inter.Codable](m T, r io.Reader) error {
	return gob.NewDecoder(r).Decode(m)
}
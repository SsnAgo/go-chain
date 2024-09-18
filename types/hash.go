package types

import (
	"fmt"
	"math/rand"
)

type Hash [32]uint8

func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		msg := fmt.Sprintf("invalid hash length: %d", len(b))
		panic(msg)
	}
	return Hash(b)
}

func (h Hash) IsZero() bool {
	for _, b := range h {
		if b!= 0 {
			return false
		}
	}
	return true
}

func RandomHash() Hash {
	var h Hash
	for i := range h {
		h[i] = uint8(rand.Intn(256))
	}
	return h
}

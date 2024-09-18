package inter

import (
	"io"
)

// 定义一个通用的接口，包含 Encode 和 Decode 方法
type Codable interface {
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

type Validator interface {
	Validate(any) bool
}

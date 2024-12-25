package encoder

import (
	"bytes"
	"io"

	"github.com/andybalholm/brotli"
)

func init() {
	encoderManger.Set("br", &Br{})
}

type Br struct {
}

func (b *Br) ToUTF8(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(reader)
}

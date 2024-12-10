package openAI

import (
	"bytes"
	"fmt"
	"io"

	"github.com/eolinker/eosc"

	"github.com/andybalholm/brotli"
)

type IEncoder interface {
	ToUTF8([]byte) ([]byte, error)
}

type EncoderManger struct {
	encoders eosc.Untyped[string, IEncoder]
}

func NewEncoderManger() *EncoderManger {
	return &EncoderManger{encoders: eosc.BuildUntyped[string, IEncoder]()}
}

func (e *EncoderManger) Set(name string, encoder IEncoder) {
	e.encoders.Set(name, encoder)
}

func (e *EncoderManger) ToUTF8(name string, data []byte) ([]byte, error) {
	encoder, ok := e.encoders.Get(name)
	if !ok {
		return nil, fmt.Errorf("encoder %s not found", name)
	}
	return encoder.ToUTF8(data)
}

var encoderManger = NewEncoderManger()

func init() {
	encoderManger.Set("br", &Br{})
}

type Br struct {
}

func (b *Br) ToUTF8(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(reader)
}

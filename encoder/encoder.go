package encoder

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type IEncoder interface {
	ToUTF8([]byte) ([]byte, error)
}

type encoder struct {
	encoders eosc.Untyped[string, IEncoder]
}

func newEncoder() *encoder {
	return &encoder{encoders: eosc.BuildUntyped[string, IEncoder]()}
}

func (e *encoder) Set(name string, encoder IEncoder) {
	e.encoders.Set(name, encoder)
}

func (e *encoder) ToUTF8(name string, data []byte) ([]byte, error) {
	enc, ok := e.encoders.Get(name)
	if !ok {
		return nil, fmt.Errorf("encoder %s not found", name)
	}
	return enc.ToUTF8(data)
}

var encoderManger = newEncoder()

func ToUTF8(name string, data []byte) ([]byte, error) {
	return encoderManger.ToUTF8(name, data)
}

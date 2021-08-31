package reader_yaml

import (
	"github.com/ghodss/yaml"
)

type BytesData []byte

func (b BytesData) Marshal() ([]byte, error) {
	return b, nil
}

func (b BytesData) UnMarshal(v interface{}) error {
	return yaml.Unmarshal(b, v)
}

func MarshalBytes(v interface{}) (BytesData, error) {
	return yaml.Marshal(v)
}

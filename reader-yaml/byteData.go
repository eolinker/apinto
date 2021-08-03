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
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	return BytesData(data), nil
}

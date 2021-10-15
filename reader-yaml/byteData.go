package reader_yaml

import (
	"github.com/ghodss/yaml"
)

//BytesData yaml编码数据
type BytesData []byte

//Marshal 获取编码后的yaml数据
func (b BytesData) Marshal() ([]byte, error) {
	return b, nil
}

//UnMarshal 获取解码后的yaml数据
func (b BytesData) UnMarshal(v interface{}) error {
	return yaml.Unmarshal(b, v)
}

//MarshalBytes 对数据进行编码并返回yaml编码数据类型
func MarshalBytes(v interface{}) (BytesData, error) {
	return yaml.Marshal(v)
}

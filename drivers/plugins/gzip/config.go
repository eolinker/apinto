package gzip

import "errors"

var (
	ErrorMinLengthError = errors.New("the min_length field should > 0")
)

type Config struct {
	Types     []string `json:"types" label:"content-type列表" description:"需要压缩的响应content-type类型列表"`
	MinLength int      `json:"min_length" label:"长度" description:"待压缩内容的最小长度" `
	Vary      bool     `json:"vary" label:"是否加上Vary头部"`
}

func (c *Config) doCheck() error {
	if c.MinLength < 0 {
		return ErrorMinLengthError
	}
	if c.MinLength == 0 {
		// 设置默认值
		c.MinLength = 1
	}
	return nil
}

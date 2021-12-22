package gzip

import "errors"

var (
	ErrorMinLengthError = errors.New("the min_length field should > 0")
)

type Config struct {
	Types       []string `json:"types"`
	MinLength   int      `json:"min_length"`
	Vary        bool     `json:"vary"`
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

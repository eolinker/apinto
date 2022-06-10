package syslog

import (
	"errors"
	"github.com/eolinker/eosc"
)

var (
	errFormatterType = errors.New("type is illegal. ")
	errAddress       = errors.New("address is illegal. ")
	errNetwork       = errors.New("network is illegal. ")
	errLevelType     = errors.New("level is illegal. ")
	errFormatterConf = errors.New("formatter config can not be null. ")
)

type Config struct {
	// 所使用的网络协议, 如:tcp,udp,unix
	Network string `json:"network" yaml:"network"`
	Address string `json:"address" yaml:"address"`
	Level   string `json:"level" yaml:"level"`

	Type      string               `json:"type" yaml:"type" description:"格式类型" enum:"line,json"`
	Formatter eosc.FormatterConfig `json:"formatter" description:"输出格式" yaml:"formatter"`
}

func (c *Config) doCheck() error {
	if c.Network == "" {
		return errNetwork
	}
	if c.Address == "" {
		return errAddress
	}
	if c.Level == "" {
		return errLevelType
	}
	if len(c.Formatter) == 0 {
		return errFormatterConf
	}
	if c.Type == "" {
		c.Type = "line"
	}
	return nil
}

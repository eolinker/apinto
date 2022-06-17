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
	Network string `json:"network" yaml:"network" enum:"tcp,udp,unix" label:"网络协议"`
	Address string `json:"address" yaml:"address" label:"请求地址"`
	Level   string `json:"level" yaml:"level" enum:"error,warn,info,debug,trace" label:"日志等级"`

	Type      string               `json:"type" yaml:"type" enum:"line,json" label:"输出格式"`
	Formatter eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
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

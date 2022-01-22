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
	Config *SysConfig `json:"config" yaml:"config"`
}

type SysConfig struct {
	Network string `json:"network" yaml:"network"`
	Address string `json:"address" yaml:"address"`
	Level   string `json:"level" yaml:"level"`

	Type      string               `json:"type" yaml:"type"`
	Formatter eosc.FormatterConfig `json:"formatter" yaml:"formatter"`
}

func (c *Config) doCheck() error {
	if c.Config.Network == "" {
		return errNetwork
	}
	if c.Config.Address == "" {
		return errAddress
	}
	if c.Config.Level == "" {
		return errLevelType
	}
	if len(c.Config.Formatter) == 0 {
		return errFormatterConf
	}
	if c.Config.Type == "" {
		c.Config.Type = "line"
	}
	return nil
}

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
	Scopes []string `json:"scopes" label:"作用域"`
	// 所使用的网络协议, 如:tcp,udp,unix
	Network       string               `json:"network" yaml:"network" enum:"tcp,udp,unix" label:"网络协议"`
	Address       string               `json:"address" yaml:"address" label:"请求地址"`
	Level         string               `json:"level" yaml:"level" enum:"error,warn,info,debug,trace" label:"日志等级"`
	Type          string               `json:"type" yaml:"type" enum:"line,json" label:"输出格式"`
	ContentResize []ContentResize      `json:"content_resize" yaml:"content_resize" label:"内容截断配置" switch:"type===json"`
	Formatter     eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}

type ContentResize struct {
	Size   int    `json:"size" label:"内容截断大小" description:"单位：M" default:"10" minimum:"0"`
	Suffix string `json:"suffix" label:"匹配标签后缀"`
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

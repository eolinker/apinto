package syslog

import (
	"errors"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/log/common/syslog-transporter"
)

type DriverConfig struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	Network       string `json:"network"`
	URL           string `json:"url"`
	Level         string `json:"level"`
	FormatterName string `json:"formatter"`
}

func toConfig(c *DriverConfig) (*syslog_transporter.Config, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.InfoLevel
	}

	config := &syslog_transporter.Config{
		Network: c.Network,
		RAddr:   c.URL,
		Level:   level,
	}

	return config, nil

}

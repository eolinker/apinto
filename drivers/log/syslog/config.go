package syslog

import (
	"errors"

	"github.com/eolinker/eosc/log"
	syslog_transporter "github.com/eolinker/goku/log-transport/syslog"
)

type DriverConfig struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	Network       string `json:"network"`
	RAddr         string `json:"raddr"`
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
		RAddr:   c.RAddr,
		Level:   level,
	}

	return config, nil

}

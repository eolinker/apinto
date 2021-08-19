package syslog

import (
	"errors"
	"github.com/eolinker/eosc/log"
)

type DriverConfig struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	Network       string `json:"network"`
	URL           string `json:"url"`
	Level         string `json:"level"`
	FormatterName string `json:"formatter"`
}

type Config struct {
	Network string
	RAddr   string
	Level   log.Level
}

func toConfig(c *DriverConfig) (*Config, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.InfoLevel
	}

	config := &Config{
		Network: c.Network,
		RAddr:   c.URL,
		Level:   level,
	}

	return config, nil

}

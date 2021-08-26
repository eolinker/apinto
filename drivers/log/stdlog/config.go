package stdlog

import (
	"errors"

	"github.com/eolinker/eosc/log"
	stdlog_transporter "github.com/eolinker/goku/log-transport/stdlog"
)

type DriverConfig struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	Level         string `json:"level"`
	FormatterName string `json:"formatter"`
}

func toConfig(c *DriverConfig) (*stdlog_transporter.Config, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.InfoLevel
	}

	config := &stdlog_transporter.Config{
		Level: level,
	}

	return config, nil

}

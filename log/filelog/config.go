package filelog

import (
	"errors"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/log/common/filelog-transporter"
)

//DriverConfig 普通log驱动配置
type DriverConfig struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	Dir           string `json:"dir"`
	File          string `json:"file"`
	Level         string `json:"level"`
	Period        string `json:"period"`
	Expire        int    `json:"expire"`
	FormatterName string `json:"formatter"`
}

func toConfig(c *DriverConfig) (*filelog_transporter.Config, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	period, err := filelog_transporter.ParsePeriod(c.Period)
	if err != nil {
		return nil, err
	}
	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.InfoLevel
	}

	config := filelog_transporter.Config{
		Dir:    c.Dir,
		File:   c.File,
		Expire: c.Expire,
		Period: period,
		Level:  level,
	}

	return &config, nil
}

package filelog

import (
	"errors"
	"github.com/eolinker/eosc/log"
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

//DriverConfig access-log驱动配置
type DriverConfigAccess struct {
	Name          string   `json:"name"`
	Driver        string   `json:"driver"`
	Dir           string   `json:"dir"`
	File          string   `json:"file"`
	Level         string   `json:"level"`
	Period        string   `json:"period"`
	Expire        int      `json:"expire"`
	FormatterName string   `json:"formatter"`
	Fields        []string `json:"fields"`
}

type Config struct {
	Dir    string
	File   string
	Expire int
	Period LogPeriod
	Level  log.Level
}

func ToConfig(c *DriverConfig) (*Config, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	period, err := ParsePeriod(c.Period)
	if err != nil {
		return nil, err
	}
	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.InfoLevel
	}

	config := Config{
		Dir:    c.Dir,
		File:   c.File,
		Expire: c.Expire,
		Period: period,
		Level:  level,
	}

	return &config, nil
}

func ToConfigAccess(c *DriverConfigAccess) (*Config, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	period, err := ParsePeriod(c.Period)
	if err != nil {
		return nil, err
	}
	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.InfoLevel
	}

	config := Config{
		Dir:    c.Dir,
		File:   c.File,
		Expire: c.Expire,
		Period: period,
		Level:  level,
	}

	return &config, nil
}

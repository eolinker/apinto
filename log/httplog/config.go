package httplog

import (
	"errors"
	"github.com/eolinker/eosc/log"
	"net/http"
)

type DriverConfig struct {
	Name          string            `json:"name"`
	Driver        string            `json:"driver"`
	Method        string            `json:"method"`
	Url           string            `json:"url"`
	Headers       map[string]string `json:"headers"`
	Level         string            `json:"level"`
	FormatterName string            `json:"formatter"`
}

type Config struct {
	Method  string
	Url     string
	Headers http.Header
	Level   log.Level

	HandlerCount int
}

func toHeader(items map[string]string) http.Header {
	header := make(http.Header)
	for k, v := range items {
		header.Set(k, v)
	}
	return header
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
		Method:       c.Method,
		Url:          c.Url,
		Headers:      toHeader(c.Headers),
		Level:        level,
		HandlerCount: 5, // 默认值， 以后可能会改成配置
	}

	return config, nil

}

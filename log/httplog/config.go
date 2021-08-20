package httplog

import (
	"errors"
	"github.com/eolinker/eosc/log"
	httplog_transporter "github.com/eolinker/goku/log/common/httplog-transporter"
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

func toHeader(items map[string]string) http.Header {
	header := make(http.Header)
	for k, v := range items {
		header.Set(k, v)
	}
	return header
}

func toConfig(c *DriverConfig) (*httplog_transporter.Config, error) {
	if c == nil {
		return nil, errors.New("config is nil")
	}

	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.InfoLevel
	}

	config := &httplog_transporter.Config{
		Method:       c.Method,
		Url:          c.Url,
		Headers:      toHeader(c.Headers),
		Level:        level,
		HandlerCount: 5, // 默认值， 以后可能会改成配置
	}

	return config, nil

}

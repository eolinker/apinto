package params_transformer

import (
	"fmt"
	"strings"
)

type Config struct {
	Params    []*TransParam `json:"params"`
	Remove    bool          `json:"remove"`
	ErrorType string        `json:"error_type"`
}

type TransParam struct {
	Name          string `json:"name"`
	Position      string `json:"position"`
	ProxyName     string `json:"proxy_name"`
	ProxyPosition string `json:"proxy_position"`
	Required      bool   `json:"required"`
	Conflict      string `json:"conflict"`
}

func (c *Config) doCheck() error {
	c.ErrorType = strings.ToLower(c.ErrorType)
	if c.ErrorType != "text" && c.ErrorType != "json" {
		c.ErrorType = "text"
	}

	for _, param := range c.Params {
		param.Position = strings.ToLower(param.Position)
		if param.Position != "query" && param.Position != "header" && param.Position != "body" {
			return fmt.Errorf(paramPositionErrInfo, param.Position)
		}

		param.ProxyPosition = strings.ToLower(param.ProxyPosition)
		if param.ProxyPosition != "query" && param.ProxyPosition != "header" && param.ProxyPosition != "body" {
			return fmt.Errorf(paramPositionErrInfo, param.ProxyPosition)
		}

		if param.Name == "" {
			return fmt.Errorf(paramNameErrInfo)
		}

		if param.ProxyName == "" {
			return fmt.Errorf(paramProxyNameErrInfo)
		}
	}

	return nil
}

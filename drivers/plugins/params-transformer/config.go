package params_transformer

import (
	"fmt"
	"strings"
)

type Config struct {
	Params    []*TransParam `json:"params" label:"参数列表"`
	Remove    bool          `json:"remove" label:"映射后删除原参数"`
	ErrorType string        `json:"error_type" enum:"text,json" label:"报错输出格式" `
}

type TransParam struct {
	Name          string `json:"name" label:"待映射参数名称"`
	Position      string `json:"position" label:"待映射参数所在位置" enum:"header,query,body"`
	ProxyName     string `json:"proxy_name" label:"目标参数名称"`
	ProxyPosition string `json:"proxy_position" label:"目标参数所在位置" enum:"header,query,body"`
	Required      bool   `json:"required" label:"待映射参数是否必含"`
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

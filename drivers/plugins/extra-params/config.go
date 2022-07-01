package extra_params

import (
	"fmt"
	"strings"
)

type Config struct {
	Params    []*ExtraParam `json:"params" label:"参数列表"`
	ErrorType string        `json:"error_type" enum:"text,json" label:"报错输出格式" `
}

type ExtraParam struct {
	Name     string `json:"name" label:"参数名"`
	Position string `json:"position" enum:"header,query,body" label:"参数位置"`
	Value    string `json:"value" label:"参数值"`
	Conflict string `json:"conflict" label:"参数冲突时的处理方式" enum:"origin,convert,error"`
}

func (c *Config) doCheck() error {
	c.ErrorType = strings.ToLower(c.ErrorType)
	if c.ErrorType != "text" && c.ErrorType != "json" {
		c.ErrorType = "text"
	}

	for _, param := range c.Params {
		if param.Name == "" {
			return fmt.Errorf(paramNameErrInfo)
		}

		param.Position = strings.ToLower(param.Position)
		if param.Position != "query" && param.Position != "header" && param.Position != "body" {
			return fmt.Errorf(paramPositionErrInfo, param.Position)
		}

		param.Conflict = strings.ToLower(param.Conflict)
		if param.Conflict != paramOrigin && param.Conflict != paramConvert && param.Conflict != paramError {
			param.Conflict = paramConvert
		}
	}

	return nil
}

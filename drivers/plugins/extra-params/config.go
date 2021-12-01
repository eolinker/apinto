package extra_params

import (
	"fmt"
	"strings"
)

type Config struct {
	Params    []*ExtraParam `json:"params"`
	ErrorType string        `json:"error_type"`
}

type ExtraParam struct {
	Name     string      `json:"name"`
	Position string      `json:"position"`
	Value    interface{} `json:"value"`
	Conflict string      `json:"conflict"`
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

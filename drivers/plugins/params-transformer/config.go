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
	errType := strings.ToLower(c.ErrorType)
	c.ErrorType = errType
	if errType != "" && errType != "text" && errType != "json" {
		return fmt.Errorf(respTypeErrInfo, errType)
	}

	for _, param := range c.Params {
		requestPosition := strings.ToLower(param.Position)
		param.Position = requestPosition
		if requestPosition != "query" && requestPosition != "header" && requestPosition != "body" {
			return fmt.Errorf(paramPositionErrInfo, requestPosition)
		}

		proxyPosition := strings.ToLower(param.Position)
		param.Position = proxyPosition
		if proxyPosition != "query" && proxyPosition != "header" && proxyPosition != "body" {
			return fmt.Errorf(paramPositionErrInfo, proxyPosition)
		}

		conflictSolution := strings.ToLower(param.Conflict)
		param.Conflict = conflictSolution
		if conflictSolution != paramOrigin && conflictSolution != paramConvert && conflictSolution != paramError {
			return fmt.Errorf(conflictSolutionErrInfo, conflictSolution)
		}
	}

	return nil
}

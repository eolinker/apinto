package params_transformer

import (
	"fmt"
	"strings"
)

type Config struct {
	Params                 []*TransParam `json:"params"`
	RemoveAfterTransformed bool          `json:"removeAfterTransformed"`
	ResponseType           string        `json:"responseType"`
}

type TransParam struct {
	ParamName             string `json:"name"`
	ParamPosition         string `json:"position"`
	ProxyParamName        string `json:"proxy_name"`
	ProxyParamPosition    string `json:"proxy_position"`
	Required              bool   `json:"required"`
	ParamConflictSolution string `json:"conflict"`
}

func (c *Config) doCheck() error {
	respType := strings.ToLower(c.ResponseType)
	c.ResponseType = respType
	if respType != "" && respType != "text" && respType != "json" {
		return fmt.Errorf(respTypeErrInfo, respType)
	}

	for _, param := range c.Params {
		requestPosition := strings.ToLower(param.ParamPosition)
		param.ParamPosition = requestPosition
		if requestPosition != "query" && requestPosition != "header" && requestPosition != "body" {
			return fmt.Errorf(paramPositionErrInfo, requestPosition)
		}

		proxyPosition := strings.ToLower(param.ParamPosition)
		param.ParamPosition = proxyPosition
		if proxyPosition != "query" && proxyPosition != "header" && proxyPosition != "body" {
			return fmt.Errorf(paramPositionErrInfo, proxyPosition)
		}

		conflictSolution := strings.ToLower(param.ParamConflictSolution)
		param.ParamConflictSolution = conflictSolution
		if conflictSolution != paramOrigin && conflictSolution != paramConvert && conflictSolution != paramError {
			return fmt.Errorf(conflictSolutionErrInfo, conflictSolution)
		}
	}

	return nil
}

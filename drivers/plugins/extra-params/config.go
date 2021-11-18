package extra_params

import (
	"fmt"
	"strings"
)

type Config struct {
	Params       []*ExtraParam `json:"params"`
	ResponseType string        `json:"responseType"`
}

type ExtraParam struct {
	ParamName             string      `json:"paramName"`
	ParamPosition         string      `json:"paramPosition"`
	ParamValue            interface{} `json:"paramValue"`
	ParamConflictSolution string      `json:"paramConflictSolution"`
}

func (c *Config) doCheck() error {
	respType := strings.ToLower(c.ResponseType)
	c.ResponseType = respType
	if respType != "" && respType != "text" && respType != "json" {
		return fmt.Errorf(respTypeErrInfo, respType)
	}

	for _, param := range c.Params {
		position := strings.ToLower(param.ParamPosition)
		param.ParamPosition = position
		if position != "query" && position != "header" && position != "body" {
			return fmt.Errorf(paramPositionErrInfo, position)
		}

		conflictSolution := strings.ToLower(param.ParamConflictSolution)
		param.ParamConflictSolution = conflictSolution
		if conflictSolution != paramOrigin && conflictSolution != paramConvert && conflictSolution != paramError {
			return fmt.Errorf(conflictSolutionErrInfo, conflictSolution)
		}
	}

	return nil
}

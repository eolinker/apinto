package extra_params

import (
	"fmt"
	"strings"
)

type Config struct {
	Params       []*ExtraParam `json:"params"`
	ResponseType string        `json:"response_type"`
}

type ExtraParam struct {
	ParamName             string      `json:"name"`
	ParamPosition         string      `json:"position"`
	ParamValue            interface{} `json:"value"`
	ParamConflictSolution string      `json:"conflict"`
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

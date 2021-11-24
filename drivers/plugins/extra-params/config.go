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
	errType := strings.ToLower(c.ErrorType)
	c.ErrorType = errType
	if errType != "" && errType != "text" && errType != "json" {
		return fmt.Errorf(respTypeErrInfo, errType)
	}

	for _, param := range c.Params {
		position := strings.ToLower(param.Position)
		param.Position = position
		if position != "query" && position != "header" && position != "body" {
			return fmt.Errorf(paramPositionErrInfo, position)
		}

		conflictSolution := strings.ToLower(param.Conflict)
		param.Conflict = conflictSolution
		if conflictSolution != paramOrigin && conflictSolution != paramConvert && conflictSolution != paramError {
			return fmt.Errorf(conflictSolutionErrInfo, conflictSolution)
		}
	}

	return nil
}

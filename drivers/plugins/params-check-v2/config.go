package params_check_v2

import (
	"fmt"

	"github.com/eolinker/apinto/checker"
)

type Config Param

type Param struct {
	Name      string   `json:"name" label:"参数名"`
	Position  string   `json:"position" label:"参数位置" enum:"query,header,body"`
	MatchText string   `json:"match_text" label:"匹配文本"`
	MatchMode string   `json:"match_mode" label:"匹配模式" enum:"any,all"`
	Logic     string   `json:"logic" label:"逻辑" enum:"and,or"`
	Params    []*Param `json:"params" label:"参数列表"`
}

func checkParam(conf *Param) error {
	if conf.Name == "" && len(conf.Params) == 0 {
		return fmt.Errorf("name is empty")
	}
	switch conf.Position {
	case positionQuery, positionHeader, positionBody, "":
	default:
		return fmt.Errorf("position is error")
	}
	switch conf.MatchMode {
	case checker.JsonArrayMatchAll, checker.JsonArrayMatchAny:
	default:
		conf.MatchMode = checker.JsonArrayMatchAll
	}
	switch conf.Logic {
	case logicAnd, logicOr:
	default:
		conf.Logic = logicAnd
	}
	for _, p := range conf.Params {
		if err := checkParam(p); err != nil {
			return err
		}
	}
	return nil
}

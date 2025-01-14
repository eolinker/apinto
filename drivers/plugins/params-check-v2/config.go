package params_check_v2

import (
	"fmt"

	"github.com/eolinker/apinto/checker"
)

type Config struct {
	Logic  string   `json:"logic" label:"逻辑" enum:"and,or"`
	Params []*Param `json:"params" label:"参数列表"`
}

type Param struct {
	Name      string      `json:"name" label:"参数名"`
	Position  string      `json:"position" label:"参数位置" enum:"query,header,body"`
	MatchText string      `json:"match_text" label:"匹配文本"`
	MatchMode string      `json:"match_mode" label:"匹配模式" enum:"any,all"`
	Logic     string      `json:"logic" label:"逻辑" enum:"and,or"`
	Params    []*SubParam `json:"params" label:"参数列表"`
}

func (p *Param) check() error {
	if p.Name == "" && len(p.Params) == 0 {
		return fmt.Errorf("name is empty")
	}
	switch p.Position {
	case positionQuery, positionHeader, positionBody, "":
	default:
		return fmt.Errorf("position is error")
	}
	switch p.MatchMode {
	case checker.JsonArrayMatchAll, checker.JsonArrayMatchAny:
	default:
		p.MatchMode = checker.JsonArrayMatchAll
	}

	switch p.Logic {
	case logicAnd, logicOr:
	default:
		p.Logic = logicAnd
	}
	for _, sub := range p.Params {
		err := sub.check()
		if err != nil {
			return err
		}
	}
	return nil
}

type SubParam struct {
	Name      string `json:"name" label:"参数名"`
	Position  string `json:"position" label:"参数位置" enum:"query,header,body"`
	MatchText string `json:"match_text" label:"匹配文本"`
	MatchMode string `json:"match_mode" label:"匹配模式" enum:"any,all"`
}

func (p *SubParam) check() error {
	if p.Name == "" {
		return fmt.Errorf("name is empty")
	}
	switch p.Position {
	case positionQuery, positionHeader, positionBody:
	default:
		return fmt.Errorf("position is error")
	}
	switch p.MatchMode {
	case checker.JsonArrayMatchAll, checker.JsonArrayMatchAny:
	default:
		p.MatchMode = checker.JsonArrayMatchAll
	}
	return nil
}

func checkParam(conf *Config) error {
	for _, p := range conf.Params {
		err := p.check()
		if err != nil {
			return err
		}
	}
	if conf.Logic == "" {
		conf.Logic = logicAnd
	}
	return nil
}

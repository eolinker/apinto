package nsq

import (
	"context"
	"strconv"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
	"github.com/ohler55/ojg/jp"
)

type Config struct {
	Scopes        []string `json:"scopes" label:"作用域"`
	Topic         string   `json:"topic" yaml:"topic" label:"topic"`
	Address       []string `json:"address" yaml:"address" label:"请求地址"`
	AuthSecret    string   `json:"auth_secret" yaml:"auth_secret" label:"鉴权secret"`
	Params        []*Param `json:"params" label:"参数列表"`
	CountParamKey string   `json:"count_param_key" label:"计数参数"`
	PushMode      string   `json:"push_mode" label:"推送模式" enum:"single,multi" default:"multi"`
}

type Param struct {
	Key   string `json:"key" label:"参数名(JSON PATH格式)"`
	Value string `json:"value" label:"参数值"`
	Type  string `json:"type" label:"参数类型" enum:"string,int,boolean" default:"string"`
}

const (
	typeString = "string"
	typeInt    = "int"
	typeBool   = "boolean"
)

type paramExpr struct {
	expr       jp.Expr
	value      string
	isVariable bool
	typ        string
}

func (p *paramExpr) GetValue(variables map[string]string) interface{} {
	value := p.value
	if p.isVariable {
		v, ok := variables[p.value]
		if ok {
			// 存在该变量，替换
			value = v
		}
	}
	switch p.typ {
	case typeInt:
		v, _ := strconv.Atoi(value)
		return v
	case typeBool:
		v, _ := strconv.ParseBool(value)
		return v
	default:
		return value
	}
}

func (p *paramExpr) GetExpr() jp.Expr {
	return p.expr
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	ctx, cancel := context.WithCancel(context.Background())
	worker := &executor{
		WorkerBase: drivers.Worker(id, name),
		cancel:     cancel,
		ctx:        ctx,
	}
	err := worker.reset(conf)
	if err != nil {
		return nil, err
	}
	go worker.doLoop()
	return worker, nil
}

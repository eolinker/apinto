package credit

import (
	"fmt"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// RuleFilter 定义不同协议（ai / api）下的过滤列表
type RuleFilter struct {
	AI  []string `json:"ai" label:"AI维度列表"`
	API []string `json:"api" label:"API维度列表"`
}

// Filters 包含供应商维度与资源维度的约束
type Filters struct {
	Provider RuleFilter `json:"provider" label:"供应商或服务过滤"`
	Resource RuleFilter `json:"resource" label:"模型或接口资源过滤"`
}

// Config 代金券规则驱动配置
type Config struct {
	Filters Filters `json:"filters" label:"过滤条件"`
	Nbf     int64   `json:"nbf" label:"生效时间戳(Unix)"`
	Exp     int64   `json:"exp" label:"失效时间戳(Unix)"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.Exp > 0 && conf.Nbf > 0 && conf.Exp <= conf.Nbf {
		return nil, fmt.Errorf("credit rule exp (%d) must be greater than nbf (%d)", conf.Exp, conf.Nbf)
	}
	return conf, nil
}

// Create 驱动实例工厂入口
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	err = w.reset(cfg)
	return w, err
}

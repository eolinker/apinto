package counter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/plugins/counter/separator"
	"github.com/eolinker/eosc"
)

type Config struct {
	Match *Match               `json:"match" label:"响应匹配规则"`
	Count *separator.CountRule `json:"count" label:"计数规则"`
	Key   string               `json:"key" label:"计数字段名称"`
	Cache eosc.RequireId       `json:"cache" label:"缓存计数器"`
}

type Match struct {
	Params      []*MatchParam `json:"params" label:"匹配参数列表"`
	StatusCodes []int         `json:"status_codes" label:"匹配响应状态码列表"`
	Type        string        `json:"type" label:"匹配类型" enum:"json"`
}

func (m *Match) GenerateHandler() []IMatcher {
	matcher := make([]IMatcher, 0, 2)
	matcher = append(matcher, newStatusCodeMatcher(m.StatusCodes))
	matcher = append(matcher, newJsonMatcher(m.Params))
	return matcher
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	counter, err := separator.GetCounter(conf.Count)
	if err != nil {
		return nil, err
	}
	bc := &executor{
		WorkerBase:       drivers.Worker(id, name),
		matchers:         conf.Match.GenerateHandler(),
		separatorCounter: counter,
	}

	return bc, nil
}

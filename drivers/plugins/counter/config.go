package counter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/counter"
	"github.com/eolinker/apinto/drivers/plugins/counter/matcher"
	"github.com/eolinker/apinto/drivers/plugins/counter/separator"
	"github.com/eolinker/eosc"
)

type Config struct {
	Key string `json:"key" label:"格式化Key" required:"true"`
	//Cache   eosc.RequireId       `json:"cache" label:"缓存计数器" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"true"`
	//Counter eosc.RequireId       `json:"counter" label:"计数器" skill:"github.com/eolinker/apinto/drivers/counter.counter.IClient" required:"false"`
	Match *Match               `json:"match" label:"响应匹配规则"`
	Count *separator.CountRule `json:"count" label:"计数规则"`
}

type Match struct {
	Params      []*matcher.MatchParam `json:"params" label:"匹配参数列表"`
	StatusCodes []int                 `json:"status_codes" label:"匹配响应状态码列表"`
	Type        string                `json:"type" label:"匹配类型" enum:"json"`
}

func (m *Match) GenerateHandler() []matcher.IMatcher {
	return []matcher.IMatcher{
		matcher.NewStatusCodeMatcher(m.StatusCodes),
		matcher.NewJsonMatcher(m.Params),
	}
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	ct, err := separator.GetCounter(conf.Count)
	if err != nil {
		return nil, err
	}
	bc := &executor{
		WorkerBase:       drivers.Worker(id, name),
		matchers:         conf.Match.GenerateHandler(),
		separatorCounter: ct,
		counters:         eosc.BuildUntyped[string, counter.ICounter](),
		keyGenerate:      newKeyGenerate(conf.Key),
		//cache:            workers[conf.Cache].(resources.ICache),
		//cacheID:          conf.Cache,
	}

	return bc, nil
}

package grey_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc"
	"strings"
)

func checkConfig(conf *Config) error {
	if conf.Priority > 999 || conf.Priority < 1 {
		return fmt.Errorf("priority value %d not allow ", conf.Priority)
	}

	if conf.Rule.Distribution == percent && (conf.Rule.Percent < 0 || conf.Rule.Percent > 10000) {
		return fmt.Errorf("percent value %d not allow ", conf.Rule.Percent)
	} else if conf.Rule.Distribution == match && len(conf.Rule.Matching) == 0 {
		return fmt.Errorf("matching rule len is 0 ")
	}

	if len(conf.Rule.Nodes) == 0 {
		return fmt.Errorf("nodes len is 0 ")
	}
	//检查灰度节点是否正确
	for _, node := range conf.Rule.Nodes {
		if strings.Count(node, "http") > 0 || strings.Count(node, "https") > 0 {
			return fmt.Errorf("node value %s cannot be http or https ", node)
		}
	}

	_, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return err
	}

	return nil
}

func Check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	return checkConfig(cfg)
}

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if err := Check(v, workers); err != nil {
		return nil, err
	}

	lg := &Grey{
		WorkerBase: drivers.Worker(id, name),
	}

	err := lg.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	controller.Store(id)
	return lg, nil
}

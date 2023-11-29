package visit_strategy

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/strategy"
)

func checkConfig(conf *Config) error {
	//if conf.Priority > 999 || conf.Priority < 1 {
	//	return fmt.Errorf("priority value %d not allow ", conf.Priority)
	//}

	_, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return err
	}

	_, err = strategy.ParseFilter(conf.Rule.InfluenceSphere)
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

	lg := &Visit{
		id:   id,
		name: name,
	}

	err := lg.reset(v, workers)
	if err != nil {
		return nil, err
	}

	controller.Store(id)
	return lg, nil
}

package circuit_breaker

import (
	"github.com/eolinker/eosc"
)

func Check(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return conf.doCheck()

}

func check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigFieldUnknown
	}
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := Check(conf, workers)
	if err != nil {
		return nil, err
	}

	cb := &CircuitBreaker{

		counter: newCircuitCount(),
		conf:    conf,
	}

	return cb, nil
}

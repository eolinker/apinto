package template

import (
	"github.com/eolinker/eosc"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	t := NewTemplate(id, name)
	err := t.Reset(v, workers)
	if err != nil {
		return nil, err
	}
	return t, nil
}

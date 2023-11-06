package kafka

import (
	"reflect"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

type Output struct {
	drivers.WorkerBase
	producer Producer
	//scopes    []string
	config    *ProducerConfig
	isRunning bool
}

func (o *Output) Output(entry eosc.IEntry) error {
	p := o.producer
	if p != nil {
		return p.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (o *Output) Start() error {
	o.isRunning = true
	p := o.producer
	if p != nil {
		return nil
	}

	p = newTProducer(o.config)

	err := p.reset(o.config)
	if err != nil {
		return err
	}
	o.producer = p
	scope_manager.Set(o.Id(), o, o.config.Scopes...)
	return nil
}

func (o *Output) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {

	cfg, err := check(conf)

	if err != nil {
		return err
	}
	if reflect.DeepEqual(cfg, o.config) {
		return nil
	}
	o.config = cfg

	if o.isRunning {
		p := o.producer
		if p == nil {
			p = newTProducer(o.config)
		}
		err = p.reset(o.config)
		if err != nil {
			return err
		}
		o.producer = p
	}
	scope_manager.Set(o.Id(), o, o.config.Scopes...)
	return nil
}

func (o *Output) Stop() error {
	scope_manager.Del(o.Id())
	producer := o.producer
	if producer != nil {
		o.producer = nil
		producer.close()
	}

	return nil
}

func (o *Output) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}

package kafka

import (
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"reflect"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

type Output struct {
	id       string
	name     string
	producer Producer
	config   *ProducerConfig
}

func (o *Output) Output(entry eosc.IEntry) error {
	p := o.producer
	if p != nil {
		return p.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (o *Output) Id() string {
	return o.id
}

func (o *Output) Start() error {

	p := o.producer
	if p == nil {
		return nil
	}

	o.producer = newTProducer(o.config)
	o.producer.reset(o.config)

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

	p := o.producer
	if p != nil {
		return p.reset(cfg)
	}
	return nil
}

func (o *Output) Stop() error {
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

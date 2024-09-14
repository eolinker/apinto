package kafka

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/eosc/log"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

type filter struct {
	key string
	checker.Checker
}

func parseFilters(filters []*Filter) []*filter {
	result := make([]*filter, 0, len(filters))
	for _, f := range filters {
		c, err := checker.Parse(f.Value)
		if err != nil {
			log.Errorf("parse filter value(%s) error: %v", f.Value, err)
			continue
		}
		result = append(result, &filter{
			key:     f.Key,
			Checker: c,
		})
	}
	return result
}

type Output struct {
	drivers.WorkerBase
	producer  Producer
	filters   []*filter
	config    *ProducerConfig
	isRunning bool
}

func (o *Output) Output(entry eosc.IEntry) error {
	for _, f := range o.filters {
		val := entry.Read(f.key)
		switch v := val.(type) {
		case string:
			ok := f.Check(v, true)
			if !ok {
				return nil
			}
		case bool:
			ok := f.Check(strconv.FormatBool(v), true)
			if !ok {
				return nil
			}
		case int, int64:
			ok := f.Check(fmt.Sprintf("%d", v), true)
			if !ok {
				return nil
			}
		default:
			continue
		}
	}
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
	o.filters = parseFilters(o.config.Filters)
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
		//err = p.reset(o.config)
		//if err != nil {
		//	return err
		//}
		o.producer = p
	}
	o.filters = parseFilters(cfg.Filters)
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

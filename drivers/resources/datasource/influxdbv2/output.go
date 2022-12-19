package influxdbv2

import (
	"context"
	"fmt"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/monitor"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*output)(nil)
var _ monitor.IOutput = (*output)(nil)

type output struct {
	drivers.WorkerBase
	cfg     *Config
	client  monitor.IClient
	ctx     context.Context
	cancel  context.CancelFunc
	metrics chan monitor.IPoint
}

func (o *output) Start() error {
	o.metrics = make(chan monitor.IPoint, 1000)
	client := NewClient(o.cfg)
	if _, err := client.Ping(o.ctx); err != nil {
		return fmt.Errorf("connect influxdbv2 eror: %w", err)
	}
	o.client = client
	scopeManager.Set(o.Id(), o, o.cfg.Scopes)
	return nil
}

func (o *output) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	client := NewClient(cfg)
	if _, err := client.Ping(o.ctx); err != nil {
		return fmt.Errorf("connect influxdbv2 eror: %w", err)
	}
	o.client = client
	scopeManager.Set(o.Id(), o, o.cfg.Scopes)
	return nil
}

func (o *output) Stop() error {
	o.client.Close()
	o.cancel()
	close(o.metrics)
	return nil
}

func (o *output) CheckSkill(skill string) bool {
	return skill == monitor.Skill
}

func (o *output) Output(metrics monitor.IPoint) {
	if o.metrics == nil {
		return
	}
	o.metrics <- metrics
}

func (o *output) doLoop() {
	for {
		select {
		case <-o.ctx.Done():
			return
		case metrics, ok := <-o.metrics:
			if !ok {
				return
			}

			err := o.client.Write(metrics)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

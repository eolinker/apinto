package influxdbv2

import (
	"context"
	"fmt"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	monitor_entry "github.com/eolinker/apinto/entries/monitor-entry"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*output)(nil)
var _ monitor_entry.IOutput = (*output)(nil)

type output struct {
	drivers.WorkerBase
	cfg     *Config
	client  monitor_entry.IClient
	ctx     context.Context
	cancel  context.CancelFunc
	metrics chan []monitor_entry.IPoint
}

func (o *output) Start() error {
	o.metrics = make(chan []monitor_entry.IPoint, 100)

	go o.doLoop()
	client := NewClient(o.cfg)
	if _, err := client.Ping(o.ctx); err != nil {
		return fmt.Errorf("connect influxdbv2 eror: %w", err)
	}
	o.client = client
	scope_manager.Set(o.Id(), o, o.cfg.Scopes...)

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
	o.cfg = cfg
	scope_manager.Set(o.Id(), o, o.cfg.Scopes...)

	return nil
}

func (o *output) Stop() error {
	scope_manager.Del(o.Id())
	o.client.Close()
	o.cancel()
	close(o.metrics)
	return nil
}

func (o *output) CheckSkill(skill string) bool {
	return skill == monitor_entry.Skill
}

func (o *output) Output(metrics ...monitor_entry.IPoint) {
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
			for _, m := range metrics {
				err := o.client.Write(m)
				if err != nil {
					log.Error("influxdbv2 write error: ", err)
				}
			}

		}
	}
}

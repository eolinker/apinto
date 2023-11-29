package nsq

import (
	"context"
	"fmt"
	"time"

	"github.com/eolinker/eosc/log"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/counter"

	"github.com/eolinker/eosc"
)

var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	topic          string
	ctx            context.Context
	cancel         context.CancelFunc
	productPool    *producerPool
	counterHandler *CounterHandler
}

func (b *executor) Push(key string, count int64, variables map[string]string) error {
	c := b.counterHandler.GetCounter(key, variables)
	c.Add(count)
	return nil
}

func (b *executor) Start() error {
	return nil
}

func (b *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type,id is %s", b.Id())
	}

	return b.reset(cfg)
}

func (b *executor) reset(conf *Config) error {
	nsqPools, err := newProducerPool(conf.Address, conf.AuthSecret, nil)
	if err != nil {
		return err
	}
	counterHandler, err := newCounterHandler(conf.Params, conf.CountParamKey, conf.PushMode)
	if err != nil {
		return err
	}
	if b.productPool != nil {
		b.productPool.Close()
	}
	scope_manager.Set(b.Id(), b, conf.Scopes...)
	b.productPool = nsqPools
	b.counterHandler = counterHandler
	b.topic = conf.Topic
	return nil
}

func (b *executor) Stop() error {
	scope_manager.Del(b.Id())
	if b.productPool != nil {
		b.productPool.Close()
	}
	b.cancel()
	return nil
}

func (b *executor) CheckSkill(skill string) bool {
	return counter.FilterSkillName == skill
}

func (b *executor) doLoop() {
	ticket := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticket.C:
			if b.counterHandler == nil || b.productPool == nil {
				continue
			}
			data, err := b.counterHandler.Generate()
			if err != nil {
				log.Error(err)
				continue
			}
			for _, d := range data {
				err = b.productPool.PublishAsync(b.topic, d)
				if err != nil {
					log.Error(err)
					continue
				}
			}
		}

	}
}

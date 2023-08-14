package counter

import (
	"fmt"

	"github.com/eolinker/apinto/drivers/plugins/counter/counter"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/plugins/counter/separator"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	matchers         []IMatcher
	separatorCounter separator.ICounter
	counters         eosc.Untyped[string, counter.ICounter]
	client           counter.IClient
	keyGenerate      IKeyGenerator
}

func (b *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(b, ctx, next)
}

func (b *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	counter, has := b.counters.Get(b.keyGenerate.Key(ctx))
	if !has {

	}
	var count int64 = 1
	var err error
	if b.separatorCounter != nil {

		separatorCounter := b.separatorCounter
		count, err = separatorCounter.Count(ctx)
		if err != nil {
			ctx.Response().SetStatus(400, "400")
			return fmt.Errorf("%s count error", separatorCounter.Name())
		}
		if count > separatorCounter.Max() {
			ctx.Response().SetStatus(403, "not allow")
			return fmt.Errorf("%s number exceed", separatorCounter.Name())
		} else if count == 0 {
			ctx.Response().SetStatus(400, "400")
			return fmt.Errorf("%s value is missing", separatorCounter.Name())
		}
	}

	err = counter.Lock(count)
	if err != nil {
		// 次数不足，直接返回
		//return fmt.Errorf("no enough, key:%s, remain:%d, count:%d", b.counters.Name(), b.counters.Remain(), count
	}
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			// 转发失败，回滚次数
			return counter.RollBack(count)
			//return err
		}
	}
	match := true
	for _, matcher := range b.matchers {
		ok := matcher.Match(ctx)
		if !ok {
			match = false
			break
		}
	}
	if match {
		// 匹配，扣减次数
		return counter.Complete(count)
	}
	// 不匹配，回滚次数
	return counter.RollBack(count)
}

func (b *executor) Start() error {
	return nil
}

func (b *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("invalid config, driver: %s", Name)
	}
	counter, err := separator.GetCounter(cfg.Count)
	if err != nil {
		return err
	}
	b.separatorCounter = counter
	b.matchers = cfg.Match.GenerateHandler()
	return nil
}

func (b *executor) Stop() error {
	return nil
}

func (b *executor) Destroy() {
	return
}

func (b *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

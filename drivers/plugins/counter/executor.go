package counter

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers/plugins/counter/matcher"

	"github.com/eolinker/apinto/resources"

	"github.com/eolinker/apinto/drivers/counter"

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
	matchers         []matcher.IMatcher
	separatorCounter separator.ICounter
	counters         eosc.Untyped[string, counter.ICounter]
	cacheID          string
	cache            scope_manager.IProxyOutput[resources.ICache]
	clientID         string
	client           scope_manager.IProxyOutput[counter.IClient]
	countPusherID    string
	counterPusher    scope_manager.IProxyOutput[counter.ICountPusher]
	keyGenerate      IKeyGenerator
	once             sync.Once
}

func (b *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(b, ctx, next)
}

func (b *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	b.once.Do(func() {
		b.cache = scope_manager.Auto[resources.ICache](b.cacheID, "redis")
		b.client = scope_manager.Auto[counter.IClient](b.clientID, "counter")
		b.counterPusher = scope_manager.Auto[counter.ICountPusher](b.countPusherID, "counter-pusher")
	})

	key := b.keyGenerate.Key(ctx)
	ct, has := b.counters.Get(key)
	if !has {
		ct = NewRedisCounter(key, b.keyGenerate.Variables(ctx), b.cache, b.client, b.counterPusher)
		b.counters.Set(key, ct)
	}
	var count int64 = 1
	var err error
	if !reflect.ValueOf(b.separatorCounter).IsNil() {
		separatorCounter := b.separatorCounter
		count, err = separatorCounter.Count(ctx)
		if err != nil {
			errInfo := fmt.Sprintf("%s count error", separatorCounter.Name())
			ctx.Response().SetStatus(400, "400")
			ctx.Response().SetBody([]byte(errInfo))
			return errors.New(errInfo)
		}
		if count > separatorCounter.Max() {
			errInfo := fmt.Sprintf("%s number exceed", separatorCounter.Name())
			ctx.Response().SetStatus(400, "not allow")
			ctx.Response().SetBody([]byte(errInfo))
			return errors.New(errInfo)
		} else if count == 0 {
			errInfo := fmt.Sprintf("%s value is missing", separatorCounter.Name())
			ctx.Response().SetStatus(400, "400")
			ctx.Response().SetBody([]byte(errInfo))
			return errors.New(errInfo)
		}
	}

	err = ct.Lock(count)
	if err != nil {
		// 次数不足，直接返回
		ctx.Response().SetStatus(416, "416")
		ctx.Response().SetBody([]byte("out of calls"))
		return err
	}
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			// 转发失败，回滚次数
			return ct.RollBack(count)
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
		return ct.Complete(count)
	}
	// 不匹配，回滚次数
	return ct.RollBack(count)
}

func (b *executor) Start() error {
	return nil
}

func (b *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	// 插件不会执行reset，会先销毁再Create
	return nil
}

func (b *executor) Stop() error {
	b.Destroy()
	return nil
}

func (b *executor) Destroy() {
	b.cache = nil
	b.client = nil
	b.counters = nil
	b.separatorCounter = nil
	b.matchers = nil
	return
}

func (b *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

package fuse_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sort"
	"sync"
	"time"
)

var (
	actuatorSet ActuatorSet
)

func init() {
	actuator := newtActuator()
	actuatorSet = actuator
}

type ActuatorSet interface {
	Strategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error
	Set(string, *FuseHandler)
	Del(id string)
}

type tActuator struct {
	lock     sync.RWMutex
	all      map[string]*FuseHandler
	handlers []*FuseHandler
}

func (a *tActuator) Destroy() {

}

func (a *tActuator) Set(id string, val *FuseHandler) {
	// 调用来源有锁
	a.all[id] = val
	a.rebuild()

}

func (a *tActuator) Del(id string) {
	// 调用来源有锁
	delete(a.all, id)
	a.rebuild()
}

func (a *tActuator) rebuild() {

	handlers := make([]*FuseHandler, 0, len(a.all))
	for _, h := range a.all {
		if !h.stop {
			handlers = append(handlers, h)
		}
	}
	sort.Sort(handlerListSort(handlers))
	a.lock.Lock()
	defer a.lock.Unlock()
	a.handlers = handlers
}
func newtActuator() *tActuator {
	return &tActuator{
		all: make(map[string]*FuseHandler),
	}
}

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain, cache resources.ICache) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	var fuseHandler *FuseHandler
	for _, handler := range handlers {
		//check筛选条件
		if !handler.filter.Check(httpCtx) {
			continue
		}
		if handler.IsFuse(ctx, cache) {
			res := handler.rule.response
			httpCtx.Response().SetStatus(res.statusCode, "")
			for _, h := range res.headers {
				httpCtx.Response().SetHeader(h.key, h.value)
			}
			httpCtx.Response().SetHeader("Content-Type", fmt.Sprintf("%s; charset=%s", res.contentType, res.charset))
			httpCtx.Response().SetBody([]byte(res.body))
			return nil
		} else {
			fuseHandler = handler
			break
		}

	}

	if next != nil {
		if err = next.DoChain(ctx); err != nil {
			return err
		}
	}

	if fuseHandler != nil {
		ctx.SetFinish(newFuseFinishHandler(ctx.GetFinish(), cache, fuseHandler))
	}

	return nil
}

type fuseFinishHandler struct {
	orgHandler  eocontext.FinishHandler
	cache       resources.ICache
	fuseHandler *FuseHandler
}

func newFuseFinishHandler(orgHandler eocontext.FinishHandler, cache resources.ICache, fuseHandler *FuseHandler) *fuseFinishHandler {
	return &fuseFinishHandler{orgHandler: orgHandler, cache: cache, fuseHandler: fuseHandler}
}

func (f *fuseFinishHandler) Finish(eoCtx eocontext.EoContext) error {
	if f.orgHandler != nil {
		if err := f.orgHandler.Finish(eoCtx); err != nil {
			return err
		}
	}

	httpCtx, _ := http_service.Assert(eoCtx)

	fuseCondition := f.fuseHandler.rule.fuseCondition
	recoverCondition := f.fuseHandler.rule.recoverCondition
	fuseTime := f.fuseHandler.rule.fuseTime

	ctx := eoCtx.Context()
	statusCode := httpCtx.Response().StatusCode()

	//熔断状态
	status := f.fuseHandler.getFuseStatus(eoCtx, f.cache)

	for _, code := range fuseCondition.statusCodes {
		if statusCode != code {
			continue
		}

		//记录失败count
		countKey := f.fuseHandler.getFuseCountKey(eoCtx)

		errCount, _ := f.cache.IncrBy(ctx, countKey, 1, time.Second).Result()

		//清除恢复的计数器
		f.cache.Del(ctx, f.fuseHandler.getRecoverCountKey(eoCtx))

		if errCount >= fuseCondition.count {
			surplus := errCount % fuseCondition.count
			if surplus == 0 {
				//熔断持续时间=连续熔断次数*持续时间
				exp := time.Second * time.Duration((errCount/fuseCondition.count)*fuseTime.time)
				maxExp := time.Duration(fuseTime.maxTime) * time.Second
				if exp >= maxExp {
					exp = maxExp
				}

				f.fuseHandler.setFuseStatus(eoCtx, f.cache, fuseStatusFusing, exp+time.Minute)

				//因为观察期是熔断时间结束后的一秒内才算观察期，所以多设置个key用来做观察期状态的判断  判断逻辑在getFuseStatus中
				f.cache.Set(ctx, f.fuseHandler.getFuseTimeKey(eoCtx), []byte(fuseStatusFusing), exp)

			}
		}
		break
	}

	for _, code := range recoverCondition.statusCodes {
		if code != statusCode {
			continue
		}
		if status == fuseStatusObserve {
			successCount, _ := f.cache.IncrBy(ctx, f.fuseHandler.getRecoverCountKey(eoCtx), 1, time.Second).Result()

			//恢复正常期
			if successCount == recoverCondition.count {
				exp := time.Duration(fuseTime.maxTime) * time.Second
				f.fuseHandler.setFuseStatus(eoCtx, f.cache, fuseStatusHealthy, exp+time.Minute)
			}

		}
		break
	}

	return nil

}

type handlerListSort []*FuseHandler

func (hs handlerListSort) Len() int {
	return len(hs)
}

func (hs handlerListSort) Less(i, j int) bool {

	return hs[i].priority < hs[j].priority
}

func (hs handlerListSort) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain, iCache resources.ICache) error {
	return actuatorSet.Strategy(ctx, next, iCache)
}

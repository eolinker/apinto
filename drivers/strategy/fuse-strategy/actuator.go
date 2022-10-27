package fuse_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	actuatorSet ActuatorSet
)

const (
	fuseStatusTime = time.Minute * 30
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

	for _, handler := range handlers {
		//check筛选条件
		if !handler.filter.Check(httpCtx) {
			continue
		}

		metrics := handler.rule.metric.Metrics(ctx)

		if handler.IsFuse(ctx.Context(), metrics, cache) {
			res := handler.rule.response
			httpCtx.Response().SetStatus(res.statusCode, "")
			for _, h := range res.headers {
				httpCtx.Response().SetHeader(h.key, h.value)
			}
			httpCtx.Response().SetHeader("Content-Type", fmt.Sprintf("%s; charset=%s", res.contentType, res.charset))
			httpCtx.Response().SetBody([]byte(res.body))
			return nil
		} else {
			ctx.SetFinish(newFuseFinishHandler(ctx.GetFinish(), cache, handler, metrics))
			break
		}

	}

	if next != nil {
		if err = next.DoChain(ctx); err != nil {
			return err
		}
	}

	return nil
}

type fuseFinishHandler struct {
	orgHandler  eocontext.FinishHandler
	cache       resources.ICache
	fuseHandler *FuseHandler
	metrics     string
}

func newFuseFinishHandler(orgHandler eocontext.FinishHandler, cache resources.ICache, fuseHandler *FuseHandler, metrics string) *fuseFinishHandler {
	return &fuseFinishHandler{
		orgHandler:  orgHandler,
		cache:       cache,
		fuseHandler: fuseHandler,
		metrics:     metrics,
	}
}

func (f *fuseFinishHandler) Finish(eoCtx eocontext.EoContext) error {

	defer func() {
		if f.orgHandler != nil {
			f.orgHandler.Finish(eoCtx)
		}
	}()

	httpCtx, _ := http_service.Assert(eoCtx)

	fuseTime := f.fuseHandler.rule.fuseTime

	ctx := eoCtx.Context()
	statusCode := httpCtx.Response().StatusCode()

	//熔断状态
	status := getFuseStatus(ctx, f.metrics, f.cache)

	switch f.fuseHandler.rule.codeStatusMap[statusCode] {
	case codeStatusError:
		//记录失败count

		tx := f.cache.Tx()
		errCount, _ := tx.IncrBy(ctx, getErrorCountKey(f.metrics), 1, time.Second).Result()
		//清除恢复的计数器
		tx.Del(ctx, getSuccessCountKey(f.metrics))
		tx.Exec(ctx)

		if errCount == f.fuseHandler.rule.fuseConditionCount {

			lockerKey := fmt.Sprintf("fuse_locker_%s", f.metrics)
			ok, err := f.cache.SetNX(ctx, lockerKey, []byte(fuseStatusObserve), time.Second).Result()
			if err != nil {
				log.Infof("fuse strategy locker  %s to %s fail:%s", status, fuseStatusFusing, err.Error())
				return err
			}
			if !ok {
				return nil
			}

			fuseCountKey := getFuseCountKey(f.metrics)
			expUnix := int64(0)

			fuseCount, _ := f.cache.IncrBy(ctx, fuseCountKey, 1, time.Hour).Result()
			txDone := f.cache.Tx()
			if status == fuseStatusHealthy {
				fuseCount = 1
				txDone.Set(ctx, fuseCountKey, []byte("1"), time.Hour)
			}

			exp := time.Duration(fuseCount) * fuseTime.time
			if exp >= fuseTime.maxTime {
				exp = fuseTime.maxTime
			}

			expUnix = time.Now().Add(exp).UnixNano()

			txDone.Set(ctx, getFuseStatusKey(f.metrics), []byte(strconv.FormatInt(expUnix, 16)), fuseStatusTime)
			txDone.Del(ctx, lockerKey)
			txDone.Exec(ctx)
		}

	case codeStatusSuccess:
		if status == fuseStatusObserve {
			successCount, _ := f.cache.IncrBy(ctx, getSuccessCountKey(f.metrics), 1, time.Second).Result()

			//恢复正常期
			if successCount == f.fuseHandler.rule.recoverConditionCount {
				lockerKey := fmt.Sprintf("fuse_locker_%s", f.metrics)
				ok, err := f.cache.SetNX(ctx, lockerKey, []byte(fuseStatusObserve), time.Second).Result()
				if err != nil {
					log.Infof("fuse strategy locker  %s to %s fail:%s", fuseStatusObserve, fuseStatusHealthy, err.Error())
					return err
				}
				if ok {

					tx := f.cache.Tx()
					//删除熔断状态的key就是恢复正常期
					tx.Del(ctx, getFuseStatusKey(f.metrics))
					//删除已记录的熔断次数
					tx.Del(ctx, getFuseCountKey(f.metrics))
					tx.Del(ctx, lockerKey)
					tx.Exec(ctx)
				}
			}

		}
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

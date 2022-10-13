package grey_strategy

import (
	"fmt"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sort"
	"sync"
)

var (
	actuatorSet ActuatorSet
)

const cookieName = "grey-cookie"

func init() {
	actuator := newtActuator()
	actuatorSet = actuator
	strategy.AddStrategyHandler(actuator)
}

type ActuatorSet interface {
	Set(string, *GreyHandler)
	Del(id string)
}

type tActuator struct {
	lock     sync.RWMutex
	all      map[string]*GreyHandler
	handlers []*GreyHandler
}

func (a *tActuator) Destroy() {

}

func (a *tActuator) Set(id string, val *GreyHandler) {
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

	handlers := make([]*GreyHandler, 0, len(a.all))
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
		all: make(map[string]*GreyHandler),
	}
}

func (a *tActuator) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	for _, handler := range handlers {
		//check筛选条件
		if handler.filter.Check(httpCtx) {
			ctx.SetBalance(newGreyBalanceHandler(ctx.GetBalance(), handler))
			break
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

type handlerListSort []*GreyHandler

func (hs handlerListSort) Len() int {
	return len(hs)
}

func (hs handlerListSort) Less(i, j int) bool {

	return hs[i].priority < hs[j].priority
}

func (hs handlerListSort) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

type GreyBalanceHandler struct {
	orgHandler  eocontext.BalanceHandler
	greyHandler *GreyHandler
}

func newGreyBalanceHandler(orgHandler eocontext.BalanceHandler, greyHandler *GreyHandler) *GreyBalanceHandler {
	return &GreyBalanceHandler{orgHandler: orgHandler, greyHandler: greyHandler}
}

func (g *GreyBalanceHandler) Select(ctx eocontext.EoContext) (eocontext.INode, error) {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return nil, err
	}

	if g.greyHandler.rule.keepSession {
		cookie := httpCtx.Request().Header().GetCookie(cookieName)
		if cookie != "" {
			return g.greyHandler.selectNodes(), nil
		}
	}

	if g.greyHandler.rule.distribution == percent {

		//round-robin算法判断是走灰度流量还是正常流量
		flow := g.greyHandler.rule.flowRobin.Select()

		if flow.GetId() == 1 { //灰度流量
			if g.greyHandler.rule.keepSession {
				httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%s", cookieName, cookieName))
			}
			return g.greyHandler.selectNodes(), nil
		}

	} else {

		//按匹配规则
		if !g.greyHandler.ruleFilter.Check(ctx) {
			//匹配失败走正常节点
			return g.orgHandler.Select(ctx)
		}

		//匹配成功
		if g.greyHandler.rule.keepSession {
			httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%s", cookieName, cookieName))
		}

		return g.greyHandler.selectNodes(), nil
	}

	//走正常节点
	return g.orgHandler.Select(ctx)
}

package grey_strategy

import (
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sort"
	"sync"
)

var (
	actuatorSet ActuatorSet
)

const cookieName = "grey-cookie-%s"

func init() {
	actuator := newtActuator()
	actuatorSet = actuator
}

type ActuatorSet interface {
	strategy.IStrategyHandler
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

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return err
	}
	// greyset: grey, notgrey, unset
	// GreyRule: greyStrategyName

	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	for _, handler := range handlers {
		//check筛选条件

		if handler.filter.Check(httpCtx) {

			if handler.IsGrey(ctx) { //是否触发灰度
				ctx.SetBalance(newGreyBalanceHandler(handler))
				break
			}
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
	greyHandler *GreyHandler
}

func newGreyBalanceHandler(greyHandler *GreyHandler) *GreyBalanceHandler {
	return &GreyBalanceHandler{greyHandler: greyHandler}
}

func (g *GreyBalanceHandler) Select(ctx eocontext.EoContext) (eocontext.INode, int, error) {
	return g.greyHandler.selectNodes(), 0, nil
}

/*
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return nil, err
	}

	cookieKey := fmt.Sprintf(cookieName, g.greyHandler.name)

	if g.greyHandler.rule.keepSession {
		cookie := httpCtx.Request().Header().GetCookie(cookieKey)
		if cookie == grey {
			return g.greyHandler.selectNodes(), nil
		} else if cookie == normal {
			return g.orgHandler.Select(ctx)
		}
	}

	if g.greyHandler.rule.greyMatch.Match(ctx) { //灰度
		httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, grey))
		return g.greyHandler.selectNodes(), nil
	} else {
		httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, normal))
		return g.orgHandler.Select(ctx)
	}

*/

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	return actuatorSet.Strategy(ctx, next)
}

package failover_strategy

import (
	"sort"
	"sync"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

// IExtractor 灾备策略提取器
// 提取规则：
// 1. 若请求的 resource 在筛选范围，则命中，不考虑其他策略；
// 2. 若不在筛选范围，则去寻找 provider 与其匹配的策略；
// 3. 若均未命中，寻找全局通用策略兜底。
type IExtractor interface {
	Set(id string, handler *FailoverHandler)
	Del(id string)
	Extract(ctx http_service.IHttpContext) *FailoverHandler
	ExtractContext(ctx eocontext.EoContext) *FailoverHandler
}

type failoverExtractor struct {
	lock             sync.RWMutex
	all              map[string]*FailoverHandler
	resourceHandlers []*FailoverHandler
	providerHandlers []*FailoverHandler
	generalHandlers  []*FailoverHandler
}

// NewExtractor 创建灾备策略提取器实例
func NewExtractor() IExtractor {
	return &failoverExtractor{
		all: make(map[string]*FailoverHandler),
	}
}

func (e *failoverExtractor) Set(id string, handler *FailoverHandler) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.all[id] = handler
	e.rebuild()
}

func (e *failoverExtractor) Del(id string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.all, id)
	e.rebuild()
}

func (e *failoverExtractor) rebuild() {
	var resourceList []*FailoverHandler
	var providerList []*FailoverHandler
	var generalList []*FailoverHandler

	for _, h := range e.all {
		if h == nil || h.IsStop() {
			continue
		}
		if h.HasResourceFilter() {
			resourceList = append(resourceList, h)
		} else if h.HasProviderFilter() {
			providerList = append(providerList, h)
		} else {
			generalList = append(generalList, h)
		}
	}

	sort.Sort(handlerListSort(resourceList))
	sort.Sort(handlerListSort(providerList))
	sort.Sort(handlerListSort(generalList))

	e.resourceHandlers = resourceList
	e.providerHandlers = providerList
	e.generalHandlers = generalList
}

func (e *failoverExtractor) Extract(ctx http_service.IHttpContext) *FailoverHandler {
	if ctx == nil {
		return nil
	}
	CompleteAIContext(ctx)

	e.lock.RLock()
	resourceHandlers := e.resourceHandlers
	providerHandlers := e.providerHandlers
	generalHandlers := e.generalHandlers
	e.lock.RUnlock()

	// 1. 若请求的 resource 在筛选范围，则命中，不考虑其他策略
	for _, h := range resourceHandlers {
		if h.Check(ctx) {
			return h
		}
	}

	// 2. 若不在筛选范围，则去寻找 provider 与其匹配的策略
	for _, h := range providerHandlers {
		if h.Check(ctx) {
			return h
		}
	}

	// 3. 通用策略（若存在既没有限制 resource 也没有限制 provider 的策略）
	for _, h := range generalHandlers {
		if h.Check(ctx) {
			return h
		}
	}

	return nil
}

func (e *failoverExtractor) ExtractContext(ctx eocontext.EoContext) *FailoverHandler {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return nil
	}
	return e.Extract(httpCtx)
}

// CompleteAIContext 补齐请求上下文中的 AI Provider、Model 与 Resource 标签
func CompleteAIContext(ctx http_service.IHttpContext) {
	provider := ctx.GetLabel("provider")
	if provider == "" {
		provider = ai_convert.GetAIProvider(ctx)
		if provider != "" {
			ctx.SetLabel("provider", provider)
		}
	}
	model := ctx.GetLabel("model")
	if model == "" {
		model = ai_convert.GetAIModel(ctx)
		if model != "" {
			ctx.SetLabel("model", model)
		}
	}
	if ctx.GetLabel("resource") == "" && provider != "" && model != "" {
		ctx.SetLabel("resource", provider+"/"+model)
	}
}

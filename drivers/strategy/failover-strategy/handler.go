package failover_strategy

import (
	"context"
	"errors"
	"net/http"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var (
	_ IFailoverHandler = (*FailoverHandler)(nil)
)

type IFailoverHandler interface {
	strategy.IFilter
	Name() string
	Priority() int
	IsStop() bool
	HasResourceFilter() bool
	HasProviderFilter() bool
	TriggerType() string
	Providers() []*ProviderConf
	IsTriggerCondition(ctx http_service.IHttpContext, err error, cost time.Duration) bool
	ApplyFailover(ctx http_service.IHttpContext, providerIndex int) (*ProviderConf, bool)
}

type FailoverHandler struct {
	name               string
	priority           int
	stop               bool
	filter             strategy.IFilter
	hasResourceFilter  bool
	hasProviderFilter  bool
	triggerType        string
	triggers           TriggersConf
	providers          []*ProviderConf
	failureStatusCodes map[int]struct{}
}

func NewFailoverHandler(conf *Config) (*FailoverHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	failureCodes := make(map[int]struct{})
	for _, code := range conf.Triggers.Failure.StatusCodes {
		failureCodes[code] = struct{}{}
	}

	hasResource := false
	if rFilters, ok := conf.Filters["resource"]; ok && len(rFilters) > 0 {
		hasResource = true
	}

	hasProvider := false
	if pFilters, ok := conf.Filters["provider"]; ok && len(pFilters) > 0 {
		hasProvider = true
	}

	return &FailoverHandler{
		name:               conf.Name,
		priority:           conf.Priority,
		stop:               conf.Stop,
		filter:             filter,
		hasResourceFilter:  hasResource,
		hasProviderFilter:  hasProvider,
		triggerType:        conf.TriggerType,
		triggers:           conf.Triggers,
		providers:          conf.Providers,
		failureStatusCodes: failureCodes,
	}, nil
}

func (h *FailoverHandler) Name() string {
	return h.name
}

func (h *FailoverHandler) Priority() int {
	return h.priority
}

func (h *FailoverHandler) IsStop() bool {
	return h.stop
}

func (h *FailoverHandler) TriggerType() string {
	return h.triggerType
}

func (h *FailoverHandler) Providers() []*ProviderConf {
	return h.providers
}

func (h *FailoverHandler) HasResourceFilter() bool {
	return h.hasResourceFilter
}

func (h *FailoverHandler) HasProviderFilter() bool {
	return h.hasProviderFilter
}

// Check 检查请求上下文是否满足策略过滤规则
func (h *FailoverHandler) Check(ctx eocontext.EoContext) bool {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return false
	}
	// 若上下文未包含 resource 标签，但有 provider 与 model，自动拼装 resource 供过滤匹配
	provider := httpCtx.GetLabel("provider")
	if provider == "" {
		provider = ai_convert.GetAIProvider(httpCtx)
	}
	model := httpCtx.GetLabel("model")
	if model == "" {
		model = ai_convert.GetAIModel(httpCtx)
	}
	if httpCtx.GetLabel("resource") == "" && provider != "" && model != "" {
		httpCtx.SetLabel("resource", provider+"/"+model)
	}

	return h.filter.Check(httpCtx)
}

// IsTriggerCondition 检查是否满足条件触发（失败或超时）
func (h *FailoverHandler) IsTriggerCondition(ctx http_service.IHttpContext, err error, cost time.Duration) bool {
	// 1. 超时触发检查
	if h.triggers.Timeout.Enabled {
		timeoutDuration := time.Duration(h.triggers.Timeout.TimeoutSeconds) * time.Second
		if timeoutDuration > 0 && cost >= timeoutDuration {
			log.Warnf("[failover] strategy %s timeout triggered: cost %v >= %v", h.name, cost, timeoutDuration)
			return true
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warnf("[failover] strategy %s timeout triggered: context deadline exceeded", h.name)
			return true
		}
		statusCode := ctx.Response().StatusCode()
		if statusCode == http.StatusGatewayTimeout || statusCode == http.StatusRequestTimeout {
			log.Warnf("[failover] strategy %s timeout triggered: status code %d", h.name, statusCode)
			return true
		}
		if ai_convert.GetAIStatus(ctx) == ai_convert.StatusTimeout {
			log.Warnf("[failover] strategy %s timeout triggered: ai status timeout", h.name)
			return true
		}
	}

	// 2. 失败触发检查
	if h.triggers.Failure.Enabled {
		if err != nil {
			log.Warnf("[failover] strategy %s failure triggered: err=%v", h.name, err)
			return true
		}
		statusCode := ctx.Response().StatusCode()
		if len(h.failureStatusCodes) > 0 {
			if _, ok := h.failureStatusCodes[statusCode]; ok {
				log.Warnf("[failover] strategy %s failure triggered: status code %d in configured codes", h.name, statusCode)
				return true
			}
		} else {
			if statusCode >= http.StatusInternalServerError {
				log.Warnf("[failover] strategy %s failure triggered: 5xx status code %d", h.name, statusCode)
				return true
			}
		}
		aiStatus := ai_convert.GetAIStatus(ctx)
		switch aiStatus {
		case ai_convert.StatusQuotaExhausted, ai_convert.StatusExceeded, ai_convert.StatusInvalid, ai_convert.StatusTimeout:
			log.Warnf("[failover] strategy %s failure triggered: ai status %s", h.name, aiStatus)
			return true
		}
	}

	return false
}

// ApplyFailover 将上下文切换为灾备供应商
func (h *FailoverHandler) ApplyFailover(ctx http_service.IHttpContext, providerIndex int) (*ProviderConf, bool) {
	if providerIndex < 0 || providerIndex >= len(h.providers) {
		return nil, false
	}
	p := h.providers[providerIndex]

	ctx.SetLabel("strategy_failover", h.name)
	ctx.SetLabel("handler", "failover")
	ctx.SetLabel("provider", p.Name)
	ai_convert.SetAIProvider(ctx, p.Name)

	if p.Model != "" {
		ctx.SetLabel("model", p.Model)
		ai_convert.SetAIModel(ctx, p.Model)
		ctx.SetLabel("resource", p.Name+"/"+p.Model)
	} else {
		currModel := ai_convert.GetAIModel(ctx)
		if currModel == "" {
			currModel = ctx.GetLabel("model")
		}
		if currModel != "" {
			ctx.SetLabel("resource", p.Name+"/"+currModel)
		}
	}

	ctx.WithValue("failover_strategy", h.name)
	ctx.WithValue("failover_provider", p.Name)
	if len(p.Config) > 0 {
		ctx.WithValue("failover_provider_config", p.Config)
	}

	ctx.Response().SetHeader("Strategy-Failover", h.name)
	ctx.Response().SetHeader("X-Failover-Provider", p.Name)

	log.Infof("[failover] strategy %s switched to provider: %s (model: %s)", h.name, p.Name, p.Model)
	return p, true
}

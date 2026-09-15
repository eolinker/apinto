package failover_strategy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/valyala/fasthttp"
)

type IHandler interface {
	Name() string
	DirectProvider() string
	DirectKey() ai_convert.IKeyResource
	TimeoutDuration() time.Duration
	ProviderNames() []string
	Keys() []ai_convert.IKeyResource
	IsTriggerCondition(ctx http_service.IHttpContext, err error, cost time.Duration) bool
	CheckTriggerCondition(ctx http_service.IHttpContext, err error, cost time.Duration) (bool, string, string)
	ApplyFailover(ctx http_service.IHttpContext, providerIndex int) (string, bool)
}

var _ IHandler = (*Handler)(nil)

type Handler struct {
	name           string
	directProvider string
	directKey      ai_convert.IKeyResource
	triggers       TriggersConf
	providers      []string
	keys           []ai_convert.IKeyResource
}

func NewHandler(cfg *Config) (IHandler, error) {
	//filter, err := strategy.ParseFilter(cfg.Filters)
	//if err != nil {
	//	return nil, err
	//}

	createFunc, has := ai_convert.GetConverterCreateFunc(cfg.Template)
	if !has {
		createFunc, _ = ai_convert.GetConverterCreateFunc("customize-openai")
	}

	var directProvider string
	var directKey ai_convert.IKeyResource
	if cfg.Direct != nil && cfg.Direct.Name != "" {
		directProvider = cfg.Direct.Name
		if cfg.Direct.Config != nil && createFunc != nil {
			tmp := map[string]string{
				"api_key":  cfg.Direct.Config.APIKey,
				"base_url": cfg.Direct.Config.BaseUrl,
			}
			cfgByte, err := json.Marshal(tmp)
			if err != nil {
				log.Errorf("marshal direct config error: %v", err)
			} else {
				cv, err := createFunc(string(cfgByte))
				if err != nil {
					log.Errorf("create direct converter error: %v", err)
				} else {
					keyId := cfg.Name + ":direct:" + cfg.Direct.Name
					directKey = ai_convert.NewKey(keyId, cfg.Direct.Name, 0, cfg.Priority, cv)
				}
			}
		}
	}

	keys := make([]ai_convert.IKeyResource, 0, len(cfg.Providers))
	providerNames := make([]string, 0, len(cfg.Providers))
	for _, cs := range cfg.Providers {
		providerNames = append(providerNames, cs.Name)
		if cs.Config == nil || createFunc == nil {
			continue
		}
		tmp := map[string]string{
			"api_key":  cs.Config.APIKey,
			"base_url": cs.Config.BaseUrl,
		}
		cfgByte, err := json.Marshal(tmp)
		if err != nil {
			log.Errorf("marshal config error: %v", err)
			continue
		}
		cv, err := createFunc(string(cfgByte))
		if err != nil {
			log.Errorf("create converter error: %v", err)
			continue
		}
		//keyId := cfg.Name + ":" + cs.Name
		keys = append(keys, ai_convert.NewKey(cs.Name, cs.Name, 0, cfg.Priority, cv))
	}

	h := &Handler{
		name:           cfg.Name,
		directProvider: directProvider,
		directKey:      directKey,
		triggers:       cfg.Triggers,
		providers:      providerNames,
		keys:           keys,
	}
	return h, nil
}

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) DirectProvider() string {
	return h.directProvider
}

func (h *Handler) DirectKey() ai_convert.IKeyResource {
	if h.directKey != nil {
		return h.directKey
	}
	if h.directProvider != "" {
		if globalKeys, ok := ai_convert.KeyResources(h.directProvider); ok && len(globalKeys) > 0 {
			return globalKeys[0]
		}
	}
	return nil
}

func (h *Handler) TimeoutDuration() time.Duration {
	if h.triggers.Timeout.TimeoutSeconds > 0 {
		return time.Duration(h.triggers.Timeout.TimeoutSeconds) * time.Second
	}
	return 0
}

func (h *Handler) ProviderNames() []string {
	return h.providers
}

func (h *Handler) Keys() []ai_convert.IKeyResource {
	keys := make([]ai_convert.IKeyResource, 0, len(h.keys))
	keys = append(keys, h.keys...)
	for _, p := range h.providers {
		hasInternal := false
		targetPrefix := h.name + ":" + p
		for _, k := range h.keys {
			if k.ID() == targetPrefix {
				hasInternal = true
				break
			}
		}
		if !hasInternal {
			if globalKeys, ok := ai_convert.KeyResources(p); ok && len(globalKeys) > 0 {
				keys = append(keys, globalKeys...)
			}
		}
	}
	return keys
}

// ApplyFailover 将上下文标记为灾备状态，记录灾备供应商标签（不覆盖原始 provider 与 model 标签）
func (h *Handler) ApplyFailover(ctx http_service.IHttpContext, providerIndex int) (string, bool) {
	if providerIndex < 0 || providerIndex >= len(h.providers) {
		return "", false
	}
	p := h.providers[providerIndex]

	ctx.SetLabel("strategy_failover", h.name)
	ctx.SetLabel("handler", "failover")
	ctx.SetLabel("failover_provider", p)

	currModel := ai_convert.GetAIModel(ctx)
	if currModel == "" {
		currModel = ctx.GetLabel("model")
	}
	if currModel != "" {
		ctx.SetLabel("failover_model", currModel)
		ctx.SetLabel("failover_resource", p+"/"+currModel)
	}

	ctx.WithValue("failover_strategy", h.name)
	ctx.WithValue("failover_provider", p)

	return p, true
}

func (h *Handler) IsTriggerCondition(ctx http_service.IHttpContext, err error, cost time.Duration) bool {
	triggered, _, _ := h.CheckTriggerCondition(ctx, err, cost)
	return triggered
}

func (h *Handler) CheckTriggerCondition(ctx http_service.IHttpContext, err error, cost time.Duration) (bool, string, string) {
	// 1. 超时触发检查
	if h.triggers.Timeout.Enabled {
		// 优先从上下文获取网关向上游发起请求的真实耗时（从 client.DoTimeout 开始计时）
		if upstreamCost, ok := context_label.GetUpstreamCost(ctx); ok {
			cost = upstreamCost
		}
		// 优先检查上下文中是否显式设置了超时标签或错误
		if context_label.IsAITimeout(ctx) {
			log.Warnf("[failover] strategy %s timeout triggered: ai timeout label detected", h.name)
			return true, "timeout", "ai timeout label detected"
		}
		if errors.Is(err, context_label.ErrAITimeout) {
			log.Warnf("[failover] strategy %s timeout triggered: ai timeout error detected", h.name)
			return true, "timeout", "ai timeout error detected"
		}
		if errors.Is(err, fasthttp.ErrTimeout) || (err != nil && strings.Contains(strings.ToLower(err.Error()), "timeout")) {
			log.Warnf("[failover] strategy %s timeout triggered: timeout error detected: %v", h.name, err)
			return true, "timeout", fmt.Sprintf("timeout error: %v", err)
		}
		if timeoutErr := context_label.GetAITimeoutError(ctx); timeoutErr != nil {
			log.Warnf("[failover] strategy %s timeout triggered: %v", h.name, timeoutErr)
			return true, "timeout", fmt.Sprintf("ai timeout error: %v", timeoutErr)
		}
		timeoutDuration := h.TimeoutDuration()
		if timeoutDuration > 0 && cost >= timeoutDuration {
			log.Warnf("[failover] strategy %s timeout triggered: cost %v >= %v", h.name, cost, timeoutDuration)
			return true, "timeout", fmt.Sprintf("request timeout (cost %v >= limit %v)", cost, timeoutDuration)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warnf("[failover] strategy %s timeout triggered: context deadline exceeded", h.name)
			return true, "timeout", "context deadline exceeded"
		}
		statusCode := ctx.Response().StatusCode()
		if statusCode == http.StatusGatewayTimeout || statusCode == http.StatusRequestTimeout {
			log.Warnf("[failover] strategy %s timeout triggered: status code %d", h.name, statusCode)
			return true, "timeout", fmt.Sprintf("http status code %d", statusCode)
		}
		if ai_convert.GetAIStatus(ctx) == ai_convert.StatusTimeout {
			log.Warnf("[failover] strategy %s timeout triggered: ai status timeout", h.name)
			return true, "timeout", "ai status timeout"
		}
	}

	// 2. 失败触发检查：
	// 指平台与上游供应商之间的调用失败（包括上游鉴权失败、连接上游网络异常、上游服务器5xx错误、上游额度不足等情况，不包含客户端业务报错）。
	if h.triggers.Failure.Enabled {
		// 若当前中断或错误是超时导致的，属于超时触发范畴；若超时规则未启用，则不应被失败触发器误触发
		if context_label.IsAITimeout(ctx) || errors.Is(err, context_label.ErrAITimeout) || errors.Is(err, fasthttp.ErrTimeout) || context_label.GetAITimeoutError(ctx) != nil || (err != nil && strings.Contains(strings.ToLower(err.Error()), "timeout")) {
			return false, "", ""
		}

		if err != nil {
			log.Warnf("[failover] strategy %s failure triggered: err=%v", h.name, err)
			return true, "failure", fmt.Sprintf("upstream error: %v", err)
		}

		// 2.1 鉴权与权限失败检测：HTTP 401 Unauthorized / 403 Forbidden 直接判定为访问失败
		statusCode := ctx.Response().StatusCode()
		if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
			log.Warnf("[failover] strategy %s failure triggered: upstream status code %d", h.name, statusCode)
			return true, "failure", fmt.Sprintf("upstream unauthorized / forbidden (status %d)", statusCode)
		}

		// 2.2 检查供应商是否显式设置了失败标签
		if context_label.HasAIFailureLabel(ctx) && context_label.IsAIFailure(ctx) {
			log.Warnf("[failover] strategy %s failure triggered: ai provider failure label is true", h.name)
			return true, "failure", "ai provider failure label detected"
		}

		// 2.3 失败兜底检查（包含 401/403、429、5xx、配额不足、凭证失效等）
		if fallback, reason := isFailureFallbackWithReason(ctx); fallback {
			log.Warnf("[failover] strategy %s failure fallback triggered: %s", h.name, reason)
			return true, "failure", reason
		}
	}

	return false, "", ""
}

// isFailureFallback 失败兜底检查
func isFailureFallback(ctx http_service.IHttpContext) bool {
	fallback, _ := isFailureFallbackWithReason(ctx)
	return fallback
}

// isFailureFallbackWithReason 失败兜底检查及原因提取
func isFailureFallbackWithReason(ctx http_service.IHttpContext) (bool, string) {
	statusCode := ctx.Response().StatusCode()

	// 1. 鉴权与权限失败：HTTP 401 Unauthorized / 403 Forbidden
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return true, fmt.Sprintf("upstream unauthorized / forbidden (status %d)", statusCode)
	}

	// 2. 检查 AI 业务模型状态（ai-convert 已转换的状态）
	aiStatus := ai_convert.GetAIStatus(ctx)
	switch aiStatus {
	case ai_convert.StatusQuotaExhausted:
		return true, "ai quota exhausted"
	case ai_convert.StatusExceeded:
		return true, "ai rate limit exceeded"
	case ai_convert.StatusExpired:
		return true, "ai credentials expired"
	case ai_convert.StatusInvalid:
		return true, "ai credentials invalid"
	case ai_convert.StatusTimeout:
		return true, "ai status timeout"
	case ai_convert.StatusInvalidRequest:
		// 客户端请求参数非法，不属于上游失败
		return false, ""
	}

	// 3. 上游额度不足 / 请求速率受限：HTTP 429 Too Many Requests
	if statusCode == http.StatusTooManyRequests {
		return true, "upstream rate limited (status 429)"
	}

	// 4. 上游服务器 5xx 错误
	if statusCode >= http.StatusInternalServerError {
		return true, fmt.Sprintf("upstream server error (status %d)", statusCode)
	}

	// 其他 HTTP 4xx 属于客户端业务传参报错，不触发灾备
	return false, ""
}

// FailoverHandler 兼容旧名称别名
type FailoverHandler = Handler

// NewFailoverHandler 兼容旧名称构造别名
func NewFailoverHandler(cfg *Config) (IHandler, error) {
	return NewHandler(cfg)
}

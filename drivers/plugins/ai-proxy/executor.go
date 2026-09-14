package ai_proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers"
	failover_strategy "github.com/eolinker/apinto/drivers/strategy/failover-strategy"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/redis/go-redis/v9"
)

var _ eocontext.IFilter = (*executor)(nil)

var (
	errProviderNotFound  = errors.New("provider not found")
	errModelTypeNotFound = errors.New("model type not found")
	errKeyNotFound       = errors.New("key not found")
)

var (
	taskCommit       = "task-commit"
	taskQuery        = "task-query"
	taskKeyGenerator = context_label.NewKeyGenerator("{product}:model:task")
)

type executor struct {
	drivers.WorkerBase
	redisID     string
	modelType   ai_convert.ModelType
	labels      map[string]string
	modelIdFrom string
	modelIdKey  string
	bodyExpr    jp.Expr
	config      string
	taskMode    string
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) extractModelID(ctx http_context.IHttpContext) (string, error) {
	if e.modelIdFrom == "path" {
		path := ctx.Request().URI().Path()

		if e.modelIdKey != "" {
			reg, err := regexp.Compile(e.modelIdKey)
			if err != nil {
				return "", fmt.Errorf("compile path regex %s error: %v", e.modelIdKey, err)
			}
			matches := reg.FindStringSubmatch(path)
			if len(matches) > 1 {
				return matches[1], nil
			} else if len(matches) > 0 {
				return matches[0], nil
			}
			return "", fmt.Errorf("model id not found in path %s using regex %s", path, e.modelIdKey)
		}

		// 默认提取路径最后的有效两段：{供应商ID}/{模型ID}
		trimmed := strings.Trim(path, "/")
		if trimmed == "" {
			return "", fmt.Errorf("empty request path")
		}
		parts := strings.Split(trimmed, "/")
		if len(parts) < 2 {
			return "", fmt.Errorf("path segments length is less than 2: %s", path)
		}
		return parts[len(parts)-2] + "/" + parts[len(parts)-1], nil
	}

	// body
	body, err := ctx.Request().Body().RawBody()
	if err != nil {
		return "", fmt.Errorf("read body error: %v", err)
	}

	obj, err := oj.Parse(body)
	if err != nil {
		return "", fmt.Errorf("parse body to json error: %v", err)
	}

	results := e.bodyExpr.Get(obj)
	if len(results) == 0 {
		return "", fmt.Errorf("json path %s get value is empty", e.modelIdKey)
	}

	strVal, ok := results[0].(string)
	if !ok {
		return "", fmt.Errorf("json path %s get value is not string", e.modelIdKey)
	}
	return strVal, nil
}

func (e *executor) handleError(ctx http_context.IHttpContext, err error) error {
	ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
	ctx.Response().SetBody([]byte(err.Error()))
	return err
}

func getCache(redisId string) (resources.ICache, error) {
	var cache resources.ICache
	var cl []resources.ICache
	if redisId != "" {
		cl = scope_manager.Auto[resources.ICache](redisId, "redis").List()
	}
	if len(cl) == 0 {
		cl = scope_manager.Get[resources.ICache]("redis").List()
	}
	if len(cl) > 0 {
		cache = cl[0]
	} else {
		return nil, fmt.Errorf("cache not found")
	}
	return cache, nil
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	// 1. 提取 ModelID
	extractedModelID, err := e.extractModelID(ctx)
	if err != nil {
		return e.handleError(ctx, err)
	}
	var cache resources.ICache
	switch e.taskMode {
	case taskQuery, taskCommit:
		cache, err = getCache(e.redisID)
		if err != nil {
			return fmt.Errorf("get cache error: %v", err)
		}

		if e.taskMode == taskQuery {
			context_label.SetTaskID(ctx, extractedModelID)
			taskKey := fmt.Sprintf("%s:%s", taskKeyGenerator.Key(ctx), extractedModelID)
			extractedModelID, err = cache.Get(ctx.Context(), taskKey).Result()
			if err != nil && !errors.Is(err, redis.Nil) {
				return fmt.Errorf("get cache error: %v", err)
			}
		}
	}

	// 2. 解析供应商与模型
	var provider, model string
	if strings.Contains(extractedModelID, "/") {
		parts := strings.SplitN(extractedModelID, "/", 2)
		provider = parts[0]
		model = parts[1]
	} else {
		// 自动检索已注册的 Balances 进行智能匹配
		for _, b := range ai_convert.Balances() {
			if b.Model() == extractedModelID {
				provider = b.Provider()
				model = b.Model()
				break
			}
		}

		if provider == "" || model == "" {
			return e.handleError(ctx, fmt.Errorf("extracted model_id '%s' has no provider. Please use '{provider_id}/{model_id}' format, configure default_provider, or register the model in gateway", extractedModelID))
		}
	}
	ctx.SetLabel("provider", provider)
	ctx.SetLabel("model", model)
	ctx.SetLabel("resource", provider+"/"+model)
	ai_convert.SetAIProvider(ctx, provider)
	ai_convert.SetAIModel(ctx, model)
	for k, v := range e.labels {
		ctx.SetLabel(k, v)
	}

	// 3. 修改并替换请求体中转发的模型 ID 参数
	if e.modelIdFrom == "body" && e.bodyExpr != nil {
		body, err := ctx.Request().Body().RawBody()
		if err == nil {
			obj, err := oj.Parse(body)
			if err == nil {
				e.bodyExpr.Set(obj, model)
				newBody, err := oj.Marshal(obj)
				if err == nil {
					ctx.Proxy().Body().SetRaw("application/json", newBody)
				}
			}
		}
	}

	cloneProxy := ctx.ProxyClone()

	defer func() {
		// 成功时在响应头中设置 Provider 和 Model（优先使用灾备切换后的供应商/模型，若未发生灾备则使用原始值）
		outProvider := ctx.GetLabel("failover_provider")
		if outProvider == "" {
			outProvider = ai_convert.GetAIProvider(ctx)
			if outProvider == "" {
				outProvider = ctx.GetLabel("provider")
			}
		}
		if outProvider != "" {
			ctx.Response().SetHeader("X-AI-Provider", outProvider)
		}

		outModel := ctx.GetLabel("failover_model")
		if outModel == "" {
			outModel = ai_convert.GetAIModel(ctx)
			if outModel == "" {
				outModel = ctx.GetLabel("model")
			}
		}
		if outModel != "" {
			ctx.Response().SetHeader("X-AI-Model", outModel)
		}
	}()

	// 4. 调度执行：结合灾备策略（Failover）或默认密钥池转换与兜底
	if err = e.dispatchProxy(ctx, provider, cloneProxy, next); err != nil {
		return err
	}

	switch e.taskMode {
	case taskCommit:
		taskId := context_label.GetTaskID(ctx)
		taskKey := fmt.Sprintf("%s:%s", taskKeyGenerator.Key(ctx), taskId)

		ok, err := cache.SetNX(ctx.Context(), taskKey, []byte(extractedModelID), 24*time.Hour).Result()
		if err != nil {
			return fmt.Errorf("set cache error: %v", err)
		}
		if !ok {
			return fmt.Errorf("task %s is in progress", taskId)
		}
	}
	return nil
}

// dispatchProxy 结合灾备策略或默认密钥池转换与 Balance 兜底逻辑
func (e *executor) dispatchProxy(ctx http_context.IHttpContext, provider string, cloneProxy http_context.IRequest, next eocontext.IChain) error {
	// 1. 检查是否命中了灾备策略（优先 resource，次之 provider，其余维度及通用兜底，最多命中一个）
	extractors := failover_strategy.ListExtractor(nil)
	for _, ext := range extractors {
		if handlers, ok := ext.Get(ctx); ok && len(handlers) > 0 {
			for _, h := range handlers {
				if h != nil {
					return e.doFailover(ctx, cloneProxy, next, h)
				}
			}
		}
	}

	// 2. 未命中策略时，执行原默认链路
	if err := e.processKeyPool(ctx, provider, cloneProxy, next, 0); err != nil {
		balances := ai_convert.Balances()
		if len(balances) == 0 {
			body := ctx.Response().GetBody()
			if len(body) == 0 {
				if ctx.Response().StatusCode() != http.StatusGatewayTimeout {
					ctx.Response().SetBody([]byte(err.Error()))
					ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
				}
			}
			return err
		}
		err = e.doBalance(ctx, cloneProxy, next, 0)
		if err != nil {
			if ctx.Response().StatusCode() != http.StatusGatewayTimeout {
				ctx.Response().SetBody([]byte(err.Error()))
				ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
			}
			return err
		}
	}
	return nil
}

// doFailover 完整的灾备策略执行逻辑：直接触发、超时中断与注入、条件触发器判决、供应商顺序 fallback 与 Balance 兜底
func (e *executor) doFailover(ctx http_context.IHttpContext, cloneProxy http_context.IRequest, next eocontext.IChain, handler failover_strategy.IHandler) error {
	if cloneProxy == nil {
		cloneProxy = ctx.ProxyClone()
	}

	defer func() {
		// 成功时在响应头中设置 Provider 和 Model（优先使用灾备切换后的供应商/模型，若未发生灾备则使用原始值）
		outProvider := ctx.GetLabel("failover_provider")
		if outProvider == "" {
			outProvider = ai_convert.GetAIProvider(ctx)
			if outProvider == "" {
				outProvider = ctx.GetLabel("provider")
			}
		}
		if outProvider != "" {
			ctx.Response().SetHeader("X-AI-Provider", outProvider)
			if ctx.GetLabel("failover_provider") != "" {
				ctx.Response().SetHeader("Strategy-Failover-Provider", outProvider)
			}
		}

		outModel := ctx.GetLabel("failover_model")
		if outModel == "" {
			outModel = ai_convert.GetAIModel(ctx)
			if outModel == "" {
				outModel = ctx.GetLabel("model")
			}
		}
		if outModel != "" {
			ctx.Response().SetHeader("X-AI-Model", outModel)
		}
	}()

	// 1. 直接触发模式
	if handler.TriggerType() == failover_strategy.TriggerTypeDirect {
		ctx.WithValue("is_block", true)
		ctx.SetLabel("handler", "failover-direct")
		ctx.SetLabel("strategy_failover", handler.Name())
		ctx.SetLabel("strategy_failover_trigger_type", failover_strategy.TriggerTypeDirect)
		ctx.SetLabel("strategy_failover_trigger_condition", failover_strategy.TriggerTypeDirect)
		ctx.SetLabel("strategy_failover_trigger_reason", "direct trigger")
		ctx.Response().SetHeader("Strategy-Failover", handler.Name())
		ctx.Response().SetHeader("Strategy-Failover-Trigger-Type", failover_strategy.TriggerTypeDirect)
		ctx.Response().SetHeader("Strategy-Failover-Trigger-Condition", failover_strategy.TriggerTypeDirect)
		ctx.Response().SetHeader("Strategy-Failover-Trigger-Reason", "direct trigger")
		return e.fallback(ctx, cloneProxy, next, handler)
	}

	// 2. 条件触发模式（包含超时控制与中断处理）
	timeoutDuration := handler.TimeoutDuration()
	start := time.Now()
	provider := ctx.GetLabel("provider")
	if provider == "" {
		provider = ai_convert.GetAIProvider(ctx)
	}
	err, isTimeout := e.doProcessKeyPoolWithTimeout(ctx, provider, cloneProxy, next, timeoutDuration)
	cost := time.Since(start)

	// 若耗时超出请求超时限制，执行超时中断并写下上下文超时错误
	if timeoutDuration > 0 && cost >= timeoutDuration {
		isTimeout = true
		e.interruptTimeout(ctx, context_label.ErrAITimeout)
	}

	// 识别到错误或超时后，根据触发器决定是否要进行供应商的顺序执行
	triggered, condition, reason := handler.CheckTriggerCondition(ctx, err, cost)
	if !triggered {
		if isTimeout && err == nil {
			err = context_label.GetAITimeoutError(ctx)
		}
		if err != nil && ctx.Response().StatusCode() != http.StatusGatewayTimeout {
			body := ctx.Response().GetBody()
			if len(body) == 0 {
				ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
				ctx.Response().SetBody([]byte(err.Error()))
			}
		}
		return err
	}

	log.Warnf("[failover] strategy %s triggered fallback (%s: %s): err=%v, cost=%v, isTimeout=%v, statusCode=%d",
		handler.Name(), condition, reason, err, cost, isTimeout, ctx.Response().StatusCode())

	ctx.WithValue("is_block", true)
	ctx.SetLabel("handler", "failover-condition")
	ctx.SetLabel("strategy_failover", handler.Name())
	ctx.SetLabel("strategy_failover_trigger_type", failover_strategy.TriggerTypeCondition)
	ctx.SetLabel("strategy_failover_trigger_condition", condition)
	ctx.SetLabel("strategy_failover_trigger_reason", reason)
	ctx.Response().SetHeader("Strategy-Failover", handler.Name())
	ctx.Response().SetHeader("Strategy-Failover-Trigger-Type", failover_strategy.TriggerTypeCondition)
	ctx.Response().SetHeader("Strategy-Failover-Trigger-Condition", condition)
	ctx.Response().SetHeader("Strategy-Failover-Trigger-Reason", reason)

	return e.fallback(ctx, cloneProxy, next, handler)
}

func (e *executor) doProcessKeyPoolWithTimeout(ctx http_context.IHttpContext, provider string, cloneProxy http_context.IRequest, next eocontext.IChain, timeout time.Duration) (error, bool) {
	err := e.processKeyPool(ctx, provider, cloneProxy, next, timeout)
	if err != nil || context_label.IsAITimeout(ctx) || ctx.Response().StatusCode() == http.StatusGatewayTimeout {
		if errors.Is(err, context_label.ErrAITimeout) || context_label.IsAITimeout(ctx) || ctx.Response().StatusCode() == http.StatusGatewayTimeout {
			return err, true
		}
		return err, false
	}
	return nil, false
}

// interruptTimeout 中断请求，设置 504 响应码并在上下文写入超时错误与标签
func (e *executor) interruptTimeout(ctx http_context.IHttpContext, timeoutErr error) {
	if timeoutErr == nil {
		timeoutErr = context_label.ErrAITimeout
	}
	ctx.Response().SetStatus(http.StatusGatewayTimeout, "Gateway Timeout")
	ctx.Response().SetBody([]byte(timeoutErr.Error()))
	context_label.SetAITimeoutError(ctx, timeoutErr)
	ai_convert.SetAIStatusTimeout(ctx)
}

// doChainWithTimeout 在时限内执行链路，超时主动中断
func (e *executor) doChainWithTimeout(ctx http_context.IHttpContext, next eocontext.IChain, timeout time.Duration) (err error, isTimeout bool) {
	if next == nil {
		return nil, false
	}
	if timeout <= 0 {
		err = next.DoChain(ctx)
		if ctx.Response().StatusCode() == http.StatusGatewayTimeout || errors.Is(err, context_label.ErrAITimeout) || errors.Is(err, context.DeadlineExceeded) {
			e.interruptTimeout(ctx, err)
			return err, true
		}
		return err, false
	}

	done := make(chan error, 1)
	go func() {
		var chainErr error
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("[ai-proxy] doChain panic: %v", r)
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		chainErr = next.DoChain(ctx)
		done <- chainErr
	}()

	select {
	case err = <-done:
		if ctx.Response().StatusCode() == http.StatusGatewayTimeout || errors.Is(err, context_label.ErrAITimeout) || errors.Is(err, context.DeadlineExceeded) {
			e.interruptTimeout(ctx, err)
			return err, true
		}
		return err, false
	case <-time.After(timeout):
		e.interruptTimeout(ctx, context_label.ErrAITimeout)
		return context_label.ErrAITimeout, true
	}
}

// fallback 执行策略配置供应商顺序重试，若全部失败则走 Balances 兜底
func (e *executor) fallback(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain, handler failover_strategy.IHandler) error {
	var fallbackErr error = errors.New("failover triggered")
	timeout := handler.TimeoutDuration()
	for _, key := range handler.Keys() {
		c, ok := key.Get(e.modelType)
		if !ok {
			log.Errorf("[ai-proxy] failover strategy %s key %s not found for model type %s", handler.Name(), key, e.modelType)
			continue
		}
		ctx.SetProxy(originProxy)
		err := c.RequestConvert(ctx, nil)
		if err != nil {
			log.Errorf("[ai-proxy] failover strategy %s key %s request convert failed: %v", handler.Name(), key, err)
			continue
		}
		if next != nil {
			err, isTimeout := e.doChainWithTimeout(ctx, next, timeout)
			if err == nil {
				context_label.ClearAITimeout(ctx)
				context_label.SetAIFailure(ctx, false)
				providerName := c.Provider()
				ctx.SetLabel("strategy_failover", handler.Name())
				ctx.SetLabel("handler", "failover")
				ctx.SetLabel("failover_provider", providerName)
				currModel := ai_convert.GetAIModel(ctx)
				if currModel == "" {
					currModel = ctx.GetLabel("model")
				}
				if currModel != "" {
					ctx.SetLabel("failover_model", currModel)
					ctx.SetLabel("failover_resource", providerName+"/"+currModel)
				}
				ctx.WithValue("failover_strategy", handler.Name())
				ctx.WithValue("failover_provider", providerName)
				return nil
			}
			fallbackErr = err
			log.Warnf("[ai-proxy] failover strategy %s key %s failed: %v", handler.Name(), key, err)
			if isTimeout {
				log.Warnf("[ai-proxy] failover strategy %s key %s timeout: %v", handler.Name(), key, err)
			}
		} else {
			context_label.ClearAITimeout(ctx)
			context_label.SetAIFailure(ctx, false)
			providerName := c.Provider()
			ctx.SetLabel("strategy_failover", handler.Name())
			ctx.SetLabel("handler", "failover")
			ctx.SetLabel("failover_provider", providerName)
			currModel := ai_convert.GetAIModel(ctx)
			if currModel == "" {
				currModel = ctx.GetLabel("model")
			}
			if currModel != "" {
				ctx.SetLabel("failover_model", currModel)
				ctx.SetLabel("failover_resource", providerName+"/"+currModel)
			}
			ctx.WithValue("failover_strategy", handler.Name())
			ctx.WithValue("failover_provider", providerName)
			return nil
		}
	}
	//// 优先尝试策略中指定的灾备供应商列表（按顺序执行）
	//providers := handler.ProviderNames()
	//for i, p := range providers {
	//	handler.ApplyFailover(ctx, i)
	//	err := e.tryProviderConf(ctx, originProxy, next, handler, p, timeout)
	//	if err == nil {
	//		context_label.ClearAITimeout(ctx)
	//		context_label.SetAIFailure(ctx, false)
	//		return nil
	//	}
	//	fallbackErr = err
	//	log.Warnf("[ai-proxy] failover strategy %s provider %s failed: %v", handler.Name(), p, err)
	//}
	//
	//// 若策略配置的供应商耗尽或未配置，fallback 到 balance 机制
	//balances := ai_convert.Balances()
	//if len(balances) == 0 {
	//	body := ctx.Response().GetBody()
	//	if len(body) == 0 {
	//		if ctx.Response().StatusCode() != http.StatusGatewayTimeout {
	//			ctx.Response().SetBody([]byte(fallbackErr.Error()))
	//			ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
	//		}
	//	}
	//	return fallbackErr
	//}

	//err := e.doBalance(ctx, originProxy, next, timeout)
	//if err != nil {
	//	if ctx.Response().StatusCode() != http.StatusGatewayTimeout {
	//		ctx.Response().SetBody([]byte(err.Error()))
	//		ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
	//	}
	//	return err
	//}
	//
	//context_label.ClearAITimeout(ctx)
	//context_label.SetAIFailure(ctx, false)
	return fallbackErr
}

//// tryProviderConf 顺序尝试单个策略配置的供应商，支持独立配置覆盖，保留原请求标签供追溯
//func (e *executor) tryProviderConf(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain, handler failover_strategy.IHandler, p *failover_strategy.ProviderConf, timeout time.Duration) error {
//	providerName := p.Name
//	modelName := ai_convert.GetAIModel(ctx)
//	if modelName == "" {
//		modelName = ctx.GetLabel("model")
//	}
//
//	// 记录灾备标签，保留原始请求的 provider/model/resource 标签用于追溯
//	ctx.SetLabel("failover_provider", providerName)
//	if modelName != "" {
//		ctx.SetLabel("failover_model", modelName)
//		ctx.SetLabel("failover_resource", providerName+"/"+modelName)
//	}
//
//	sysProvider, hasSys := ai_convert.GetProvider(providerName)
//	var extender map[string]interface{}
//	if hasSys && sysProvider != nil {
//		if balanceHandler := sysProvider.BalanceHandler(); balanceHandler != nil {
//			ctx.SetBalance(balanceHandler)
//		}
//		extender = sysProvider.ModelConfig()
//	}
//	if p.Config != nil {
//		cfgMap := p.Config.ToMap()
//		if len(cfgMap) > 0 {
//			extender = cfgMap
//		}
//	}
//	if extender == nil {
//		extender = make(map[string]interface{})
//	}
//
//	// 优先使用策略自带的 KeyResource，若无再回退全局 KeyResources
//	var keyResource ai_convert.IKeyResource
//	if handler != nil {
//		targetKeyID := handler.Name() + ":" + providerName
//		for _, k := range handler.Keys() {
//			if k.ID() == targetKeyID || strings.HasSuffix(k.ID(), ":"+providerName) {
//				keyResource = k
//				break
//			}
//		}
//	}
//
//	var resources []ai_convert.IKeyResource
//	if keyResource != nil {
//		resources = []ai_convert.IKeyResource{keyResource}
//	} else if globalResources, has := ai_convert.KeyResources(providerName); has && len(globalResources) > 0 {
//		resources = globalResources
//	}
//
//	if len(resources) == 0 {
//		return fmt.Errorf("%w: provider %s", errKeyNotFound, providerName)
//	}
//
//	for _, resource := range resources {
//		if originProxy != nil {
//			ctx.SetProxy(originProxy)
//		}
//		ai_convert.SetAIKey(ctx, resource.ID())
//		err := e.doConverter(ctx, next, resource, sysProvider, extender, timeout)
//		if err != nil {
//			log.Errorf("[ai-proxy] try provider %s key %s error: %v", providerName, resource.ID(), err)
//			continue
//		}
//		return nil
//	}
//
//	return fmt.Errorf("provider %s exhausted", providerName)
//}

// processKeyPool handles processing using the key pool resources.
func (e *executor) processKeyPool(ctx http_context.IHttpContext, provider string, cloneProxy http_context.IRequest, next eocontext.IChain, timeout time.Duration) error {
	p, has := ai_convert.GetProvider(provider)
	if !has {
		if next != nil {
			return e.processNext(ctx, next, nil, timeout)
		}
		return errProviderNotFound
	}
	extender, err := p.GenExtender(e.config)
	if err != nil {
		return err
	}
	balanceHandler := p.BalanceHandler()
	if balanceHandler != nil {
		ctx.SetBalance(balanceHandler)
	}
	resources, has := ai_convert.KeyResources(provider)
	if !has || len(resources) == 0 {
		if next != nil {
			return e.processNext(ctx, next, p, timeout)
		}
		return errKeyNotFound
	}
	r := resources[0]
	if cloneProxy != nil {
		ctx.SetProxy(cloneProxy)
	}
	ai_convert.SetAIKey(ctx, r.ID())
	converter, has := r.Get(e.modelType)
	if !has {
		return fmt.Errorf("key %s does not support model type %s", r.ID(), e.modelType)
	}
	if err = converter.RequestConvert(ctx, extender); err != nil {
		return fmt.Errorf("request convert error: %v", err)
	}

	if next != nil {
		if err = e.processNext(ctx, next, p, timeout); err != nil {
			return err
		}
	}
	if ctx.Response().IsBodyStream() {
		contentType := ctx.GetLabel("response-content-type")
		if contentType != "" {
			ctx.Response().SetHeader("Content-Type", contentType)
		}
		return nil
	}
	if err = converter.ResponseConvert(ctx); err != nil {
		return fmt.Errorf("response convert error: %v", err)
	}
	return nil
}

// doBalance handles fallback logic for switching providers when keys are invalid or exhausted.
func (e *executor) doBalance(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain, timeout time.Duration) error {
	balances := ai_convert.Balances()
	if len(balances) == 0 {
		return nil
	}
	for _, balance := range balances {
		log.DebugF("[ai-proxy] trying balance %s, model: %s, health: %v", balance.Provider(), balance.Model(), balance.Health())
		if !balance.Health() {
			continue
		}
		balanceHandler := balance.BalanceHandler()
		if balanceHandler != nil {
			ctx.SetBalance(balanceHandler)
		}
		err := e.tryProvider(ctx, originProxy, next, balance, timeout)
		if err == nil {
			return nil
		}
		balance.Down() // Mark the balance as unhealthy
	}

	return errors.New("all balances exhausted or unavailable")
}

// tryProvider attempts to use a single provider and its resources for processing.
func (e *executor) tryProvider(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain, provider ai_convert.IProvider, timeout time.Duration) error {
	// 记录灾备标签，保留原始请求的 provider/model/resource 标签用于追溯
	ctx.SetLabel("failover_provider", provider.Provider())
	ctx.SetLabel("failover_model", provider.Model())
	ctx.SetLabel("failover_resource", provider.Provider()+"/"+provider.Model())

	ai_convert.SetAIProvider(ctx, provider.Provider())
	ai_convert.SetAIModel(ctx, provider.Model())
	extender := provider.ModelConfig()
	resources, has := ai_convert.KeyResources(provider.Provider())
	if !has || len(resources) == 0 {
		return errKeyNotFound
	}

	for _, resource := range resources {
		if originProxy != nil {
			ctx.SetProxy(originProxy)
		}
		ai_convert.SetAIKey(ctx, resource.ID())
		err := e.doConverter(ctx, next, resource, provider, extender, timeout)
		if err != nil {
			log.Errorf("[ai-proxy] try balance provider %s error: %v", provider.Provider(), err)
			continue
		}
		return nil
	}

	return errors.New("provider exhausted")
}

func (e *executor) doConverter(ctx http_context.IHttpContext, next eocontext.IChain, resource ai_convert.IKeyResource, provider ai_convert.IProvider, extender map[string]interface{}, timeout time.Duration) error {
	status := ai_convert.StatusInvalid
	defer func() {
		ai_convert.SetAIProviderStatuses(ctx, status)
	}()
	converter, has := resource.Get(e.modelType)
	if !has {
		return errModelTypeNotFound
	}

	if err := converter.RequestConvert(ctx, extender); err != nil {
		return err
	}

	if next != nil {
		if err := e.processNext(ctx, next, provider, timeout); err != nil {
			return err
		}
	}
	if ctx.Response().IsBodyStream() {
		contentType := ctx.GetLabel("response-content-type")
		if contentType != "" {
			ctx.Response().SetHeader("Content-Type", contentType)
		}
		return nil
	}
	if err := converter.ResponseConvert(ctx); err != nil {
		return err
	}
	status = ai_convert.GetAIStatus(ctx)
	switch status {
	case ai_convert.StatusInvalid, ai_convert.StatusExpired, ai_convert.StatusQuotaExhausted:
		resource.Down()
	case ai_convert.StatusExceeded:
		resource.Breaker()
	}

	return nil
}

// processNext processes the next chain in the filter, handling 504 errors and timeouts.
func (e *executor) processNext(ctx http_context.IHttpContext, next eocontext.IChain, provider ai_convert.IProvider, timeout time.Duration) error {
	err, isTimeout := e.doChainWithTimeout(ctx, next, timeout)
	if err != nil || isTimeout {
		if ctx.Response().StatusCode() == http.StatusGatewayTimeout || isTimeout {
			ai_convert.SetAIStatusTimeout(ctx)
			if provider != nil {
				provider.Down()
			}
		}
		if err != nil {
			return err
		}
		return context_label.ErrAITimeout
	}
	return nil
}

// DoFailoverWithModelType 供外部策略插件或独立模块复用完整的灾备执行核心逻辑
func DoFailoverWithModelType(ctx http_context.IHttpContext, next eocontext.IChain, handler failover_strategy.IHandler, modelType ai_convert.ModelType) error {
	if modelType == "" {
		modelType = ai_convert.ModelTypeOpenAIChat
	}
	exec := &executor{
		modelType: modelType,
		config:    "{}",
	}
	return exec.doFailover(ctx, nil, next, handler)
}

func (e *executor) Destroy() {
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg := conf.(*Config)
	return e.reset(cfg)
}

func (e *executor) reset(cfg *Config) error {
	e.modelType = ai_convert.ModelType(cfg.ModelType)
	e.labels = cfg.Labels
	e.modelIdFrom = cfg.ModelIdFrom
	e.modelIdKey = cfg.ModelIdKey
	if e.modelIdFrom == "body" {
		expr, err := jp.ParseString(cfg.ModelIdKey)
		if err != nil {
			return err
		}
		e.bodyExpr = expr
	}

	e.config = "{}"
	if strings.Contains(cfg.ModelType, "task-commit") {
		e.taskMode = taskCommit
	} else if strings.Contains(cfg.ModelType, "task-query") {
		e.taskMode = taskQuery
	}

	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

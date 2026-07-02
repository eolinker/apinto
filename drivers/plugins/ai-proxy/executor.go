package ai_proxy

import (
	"errors"
	"fmt"
	context_label2 "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/redis/go-redis/v9"
	"regexp"
	"strings"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
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
	taskKeyGenerator = context_label2.NewKeyGenerator("{product}:model:task")
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

		// 默认提取路径最后的有效两段段：{供应商ID}/{模型ID}
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
	ctx.Response().SetStatus(400, "Bad Request")
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
			context_label2.SetTaskID(ctx, extractedModelID)
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
		//// 若未匹配到，使用配置的默认供应商作为兜底
		//if provider == "" && e.defaultProvider != "" {
		//	provider = e.defaultProvider
		//	model = extractedModelID
		//}

		if provider == "" || model == "" {
			return e.handleError(ctx, fmt.Errorf("extracted model_id '%s' has no provider. Please use '{provider_id}/{model_id}' format, configure default_provider, or register the model in gateway", extractedModelID))
		}
	}
	ctx.SetLabel("provider", provider)
	ctx.SetLabel("model", model)
	ai_convert.SetAIProvider(ctx, provider)
	ai_convert.SetAIModel(ctx, model)
	for k, v := range e.labels {
		ctx.SetLabel(k, v)
	}

	// 4. 修改并替换请求体中转发的模型 ID 参数
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
		// If the request is successful, set the AI provider and model in the response headers
		ctx.Response().SetHeader("X-AI-Provider", ai_convert.GetAIProvider(ctx))
		ctx.Response().SetHeader("X-AI-Model", ai_convert.GetAIModel(ctx))
	}()

	// 5. 调用密钥池与转换代理核心
	if err = e.processKeyPool(ctx, provider, cloneProxy, next); err != nil {
		balances := ai_convert.Balances()
		if len(balances) == 0 {
			body := ctx.Response().GetBody()
			if len(body) == 0 {
				if ctx.Response().StatusCode() != 504 {
					ctx.Response().SetBody([]byte(err.Error()))
					ctx.Response().SetStatus(400, "Bad Request")
				}
			}
			return err
		}
		err = e.doBalance(ctx, cloneProxy, next) // Fallback to balance logic
		if err != nil {
			if ctx.Response().StatusCode() != 504 {
				ctx.Response().SetBody([]byte(err.Error()))
				ctx.Response().SetStatus(400, "Bad Request")
			}
			return err
		}
	}

	switch e.taskMode {
	case taskCommit:
		taskId := context_label2.GetTaskID(ctx)
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

// processKeyPool handles processing using the key pool resources.
func (e *executor) processKeyPool(ctx http_context.IHttpContext, provider string, cloneProxy http_context.IRequest, next eocontext.IChain) error {
	p, has := ai_convert.GetProvider(provider)
	if !has {
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
	if !has {
		return errKeyNotFound
	}
	r := resources[0]
	ctx.SetProxy(cloneProxy)
	ai_convert.SetAIKey(ctx, r.ID())
	converter, has := r.Get(e.modelType)
	if !has {
		return fmt.Errorf("key %s does not support model type %s", r.ID(), e.modelType)
	}
	if err = converter.RequestConvert(ctx, extender); err != nil {
		return fmt.Errorf("request convert error: %v", err)
	}

	if next != nil {
		if err = e.processNext(ctx, next, p); err != nil {
			return err
		}
	}
	if ctx.Response().IsBodyStream() {
		contentType := ctx.GetLabel("response-content-type")
		if ctx.GetLabel("response-content-type") != "" {
			ctx.Response().SetHeader("Content-Type", contentType)
		}
		return nil
	}
	if err = converter.ResponseConvert(ctx); err != nil {
		return fmt.Errorf("response convert error: %v", err)
	}
	//aiStatus := ai_convert.GetAIStatus(ctx)
	//switch aiStatus {
	//case ai_convert.StatusInvalidRequest, ai_convert.StatusNormal:
	//	return nil
	//default:
	//	continue
	//}
	return nil
}

// doBalance handles fallback logic for switching providers when keys are invalid or exhausted.
func (e *executor) doBalance(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain) error {
	balances := ai_convert.Balances()
	if len(balances) == 0 {
		return nil
	}
	for _, balance := range balances {
		log.DebugF("trying balance %s,model:%s,health:%s", balance.Provider(), balance.Model(), balance.Health())
		if !balance.Health() {
			continue
		}
		balanceHandler := balance.BalanceHandler()
		if balanceHandler != nil {
			ctx.SetBalance(balanceHandler)
		}
		err := e.tryProvider(ctx, originProxy, next, balance)
		if err == nil {
			return nil
		}
		balance.Down() // Mark the balance as unhealthy
	}

	return errors.New("all balances exhausted or unavailable")
}

// tryProvider attempts to use a single provider and its resources for processing.
func (e *executor) tryProvider(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain, provider ai_convert.IProvider) error {
	ai_convert.SetAIProvider(ctx, provider.Provider())
	ai_convert.SetAIModel(ctx, provider.Model())
	extender := provider.ModelConfig()
	resources, has := ai_convert.KeyResources(provider.Provider())
	if !has {
		return errKeyNotFound
	}

	for _, resource := range resources {
		ctx.SetProxy(originProxy)
		ai_convert.SetAIKey(ctx, resource.ID())
		err := e.doConverter(ctx, next, resource, provider, extender)
		if err != nil {
			log.Errorf("try provider error: %v", err)
			continue
		}
		return nil
	}

	return errors.New("provider exhausted")
}

func (e *executor) doConverter(ctx http_context.IHttpContext, next eocontext.IChain, resource ai_convert.IKeyResource, provider ai_convert.IProvider, extender map[string]interface{}) error {
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
		if err := e.processNext(ctx, next, provider); err != nil {
			return err
		}
	}
	if ctx.Response().IsBodyStream() {
		contentType := ctx.GetLabel("response-content-type")
		if ctx.GetLabel("response-content-type") != "" {
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

// processNext processes the next chain in the filter, handling 504 errors.
func (e *executor) processNext(ctx http_context.IHttpContext, next eocontext.IChain, provider ai_convert.IProvider) error {
	if err := next.DoChain(ctx); err != nil {
		if ctx.Response().StatusCode() == 504 {
			ai_convert.SetAIStatusTimeout(ctx)
			provider.Down() // Mark provider as unhealthy on timeout
		}
		return err
	}
	return nil
}

func (e *executor) Destroy() {
	return
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

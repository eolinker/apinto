package failover

import (
	"errors"
	"fmt"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/drivers"
	failover_strategy "github.com/eolinker/apinto/drivers/strategy/failover-strategy"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

var (
	errKeyNotFound       = errors.New("key not found")
	errModelTypeNotFound = errors.New("model type not found")
)

type Strategy struct {
	drivers.WorkerBase
	modelType ai_convert.ModelType
}

func (s *Strategy) DoFilter(ctx eoscContext.EoContext, next eoscContext.IChain) (err error) {
	return http_service.DoHttpFilter(s, ctx, next)
}

func (s *Strategy) DoHttpFilter(ctx http_service.IHttpContext, next eoscContext.IChain) error {
	// 1. 先根据 resource 进行命中，若未命中，则根据 provider 进行命中，命中的策略最多只有一个
	matchedHandler, has := failover_strategy.GetStrategy(ctx)
	if !has || matchedHandler == nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return nil
	}

	// 2. 命中后参考 ai-proxy/executor.go#L212-230 实现灾备
	return s.doFailover(ctx, next, matchedHandler)
}

func (s *Strategy) doFailover(ctx http_service.IHttpContext, next eoscContext.IChain, handler *failover_strategy.FailoverHandler) error {
	cloneProxy := ctx.ProxyClone()

	defer func() {
		// 成功时在响应头中设置 Provider 和 Model
		provider := ai_convert.GetAIProvider(ctx)
		if provider == "" {
			provider = ctx.GetLabel("provider")
		}
		if provider != "" {
			ctx.Response().SetHeader("X-AI-Provider", provider)
		}
		model := ai_convert.GetAIModel(ctx)
		if model == "" {
			model = ctx.GetLabel("model")
		}
		if model != "" {
			ctx.Response().SetHeader("X-AI-Model", model)
		}
	}()

	// 1. 直接触发模式
	if handler.TriggerType() == failover_strategy.TriggerTypeDirect {
		ctx.WithValue("is_block", true)
		ctx.SetLabel("handler", "failover-direct")
		ctx.Response().SetHeader("Strategy-Failover", handler.Name())
		return s.fallback(ctx, cloneProxy, next, handler)
	}

	// 2. 条件触发模式
	start := time.Now()
	var err error
	if next != nil {
		err = next.DoChain(ctx)
	}
	cost := time.Since(start)

	if !handler.IsTriggerCondition(ctx, err, cost) && err == nil {
		return nil
	}

	log.Warnf("[failover] strategy %s triggered fallback: err=%v, cost=%v, statusCode=%d",
		handler.Name(), err, cost, ctx.Response().StatusCode())

	ctx.WithValue("is_block", true)
	ctx.SetLabel("handler", "failover-condition")
	ctx.Response().SetHeader("Strategy-Failover", handler.Name())

	return s.fallback(ctx, cloneProxy, next, handler)
}

func (s *Strategy) fallback(ctx http_service.IHttpContext, originProxy http_service.IRequest, next eoscContext.IChain, handler *failover_strategy.FailoverHandler) error {
	var fallbackErr error = errors.New("failover triggered")

	// 优先尝试策略中指定的灾备供应商列表
	providers := handler.Providers()
	for i, p := range providers {
		handler.ApplyFailover(ctx, i)
		err := s.tryProviderConf(ctx, originProxy, next, p)
		if err == nil {
			return nil
		}
		fallbackErr = err
		log.Warnf("[failover] strategy %s provider %s failed: %v", handler.Name(), p.Name, err)
	}

	// 若策略配置的供应商耗尽或未配置，fallback 到 balance 机制
	balances := ai_convert.Balances()
	if len(balances) == 0 {
		body := ctx.Response().GetBody()
		if len(body) == 0 {
			if ctx.Response().StatusCode() != 504 {
				ctx.Response().SetBody([]byte(fallbackErr.Error()))
				ctx.Response().SetStatus(400, "Bad Request")
			}
		}
		return fallbackErr
	}

	err := s.doBalance(ctx, originProxy, next)
	if err != nil {
		if ctx.Response().StatusCode() != 504 {
			ctx.Response().SetBody([]byte(err.Error()))
			ctx.Response().SetStatus(400, "Bad Request")
		}
		return err
	}

	return nil
}

func (s *Strategy) doBalance(ctx http_service.IHttpContext, originProxy http_service.IRequest, next eoscContext.IChain) error {
	balances := ai_convert.Balances()
	if len(balances) == 0 {
		return nil
	}
	for _, balance := range balances {
		log.DebugF("failover trying balance %s, model: %s, health: %v", balance.Provider(), balance.Model(), balance.Health())
		if !balance.Health() {
			continue
		}
		balanceHandler := balance.BalanceHandler()
		if balanceHandler != nil {
			ctx.SetBalance(balanceHandler)
		}
		err := s.tryProvider(ctx, originProxy, next, balance)
		if err == nil {
			return nil
		}
		balance.Down()
	}
	return errors.New("all balances exhausted or unavailable")
}

func (s *Strategy) tryProviderConf(ctx http_service.IHttpContext, originProxy http_service.IRequest, next eoscContext.IChain, p *failover_strategy.ProviderConf) error {
	providerName := p.Name
	modelName := p.Model
	if modelName == "" {
		modelName = ai_convert.GetAIModel(ctx)
		if modelName == "" {
			modelName = ctx.GetLabel("model")
		}
	}

	ai_convert.SetAIProvider(ctx, providerName)
	ctx.SetLabel("provider", providerName)
	if modelName != "" {
		ai_convert.SetAIModel(ctx, modelName)
		ctx.SetLabel("model", modelName)
		ctx.SetLabel("resource", providerName+"/"+modelName)
	}

	sysProvider, hasSys := ai_convert.GetProvider(providerName)
	var extender map[string]interface{}
	if hasSys && sysProvider != nil {
		if balanceHandler := sysProvider.BalanceHandler(); balanceHandler != nil {
			ctx.SetBalance(balanceHandler)
		}
		extender = sysProvider.ModelConfig()
	}
	if len(p.Config) > 0 {
		extender = p.Config
	}
	if extender == nil {
		extender = make(map[string]interface{})
	}

	resources, has := ai_convert.KeyResources(providerName)
	if !has || len(resources) == 0 {
		return fmt.Errorf("%w: provider %s", errKeyNotFound, providerName)
	}

	for _, resource := range resources {
		if originProxy != nil {
			ctx.SetProxy(originProxy)
		}
		ai_convert.SetAIKey(ctx, resource.ID())
		err := s.doConverter(ctx, next, resource, sysProvider, extender)
		if err != nil {
			log.Errorf("[failover] try provider %s key %s error: %v", providerName, resource.ID(), err)
			continue
		}
		return nil
	}

	return fmt.Errorf("provider %s exhausted", providerName)
}

func (s *Strategy) tryProvider(ctx http_service.IHttpContext, originProxy http_service.IRequest, next eoscContext.IChain, provider ai_convert.IProvider) error {
	ai_convert.SetAIProvider(ctx, provider.Provider())
	ai_convert.SetAIModel(ctx, provider.Model())
	ctx.SetLabel("provider", provider.Provider())
	ctx.SetLabel("model", provider.Model())
	ctx.SetLabel("resource", provider.Provider()+"/"+provider.Model())

	extender := provider.ModelConfig()
	if extender == nil {
		extender = make(map[string]interface{})
	}
	resources, has := ai_convert.KeyResources(provider.Provider())
	if !has || len(resources) == 0 {
		return errKeyNotFound
	}

	for _, resource := range resources {
		if originProxy != nil {
			ctx.SetProxy(originProxy)
		}
		ai_convert.SetAIKey(ctx, resource.ID())
		err := s.doConverter(ctx, next, resource, provider, extender)
		if err != nil {
			log.Errorf("[failover] try balance provider %s key %s error: %v", provider.Provider(), resource.ID(), err)
			continue
		}
		return nil
	}

	return errors.New("provider exhausted")
}

func (s *Strategy) doConverter(ctx http_service.IHttpContext, next eoscContext.IChain, resource ai_convert.IKeyResource, provider ai_convert.IProvider, extender map[string]interface{}) error {
	status := ai_convert.StatusInvalid
	defer func() {
		ai_convert.SetAIProviderStatuses(ctx, status)
	}()

	modelType := s.modelType
	if modelType == "" {
		modelType = ai_convert.ModelTypeOpenAIChat
	}

	converter, has := resource.Get(modelType)
	if !has {
		return errModelTypeNotFound
	}

	if err := converter.RequestConvert(ctx, extender); err != nil {
		return err
	}

	if next != nil {
		if err := s.processNext(ctx, next, provider); err != nil {
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

func (s *Strategy) processNext(ctx http_service.IHttpContext, next eoscContext.IChain, provider ai_convert.IProvider) error {
	if err := next.DoChain(ctx); err != nil {
		if ctx.Response().StatusCode() == 504 {
			ai_convert.SetAIStatusTimeout(ctx)
			if provider != nil {
				provider.Down()
			}
		}
		return err
	}
	return nil
}

func (s *Strategy) Destroy() {
}

func (s *Strategy) Start() error {
	return nil
}

func (s *Strategy) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := drivers.Assert[Config](conf)
	if err != nil {
		return err
	}
	if cfg.ModelType == "" {
		s.modelType = ai_convert.ModelTypeOpenAIChat
	} else {
		s.modelType = ai_convert.ModelType(cfg.ModelType)
	}
	return nil
}

func (s *Strategy) Stop() error {
	return nil
}

func (s *Strategy) CheckSkill(skill string) bool {
	return eoscContext.FilterSkillName == skill
}

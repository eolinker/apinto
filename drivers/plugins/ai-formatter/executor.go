package ai_formatter

import (
	"errors"
	"fmt"

	ai_convert "github.com/eolinker/apinto/ai-convert"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	errProviderNotFound = errors.New("provider not found")
	errKeyNotFound      = errors.New("key not found")
)

type executor struct {
	drivers.WorkerBase
	provider string
	model    string
	modelCfg string
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) error {
	return http_context.DoHttpFilter(e, ctx, next)
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

func (e *executor) doConverter(ctx http_context.IHttpContext, next eocontext.IChain, resource ai_convert.IKeyResource, provider ai_convert.IProvider, extender map[string]interface{}) error {
	status := ai_convert.StatusInvalid
	defer func() {
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.AIProviderStatus{
			Provider: provider.Provider(),
			Model:    provider.Model(),
			Key:      resource.ID(),
			Status:   status,
		})
	}()

	if err := resource.RequestConvert(ctx, extender); err != nil {
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
	if err := resource.ResponseConvert(ctx); err != nil {
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

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {

	cloneProxy := ctx.ProxyClone()
	provider := ai_convert.GetAIProvider(ctx)

	if provider == "" {
		provider = e.provider
		ai_convert.SetAIProvider(ctx, e.provider)
	}
	_, has := ai_convert.GetProvider(provider)
	if !has {
		ai_convert.SetAIProvider(ctx, e.provider)
		provider = e.provider
		ai_convert.SetAIModel(ctx, e.model)
	}
	if ai_convert.GetAIModel(ctx) == "" {
		ai_convert.SetAIModel(ctx, e.model)
	}

	defer func() {
		// If the request is successful, set the AI provider and model in the response headers
		ctx.Response().SetHeader("X-AI-Provider", ai_convert.GetAIProvider(ctx))
		ctx.Response().SetHeader("X-AI-Model", ai_convert.GetAIModel(ctx))
	}()
	if err := e.processKeyPool(ctx, provider, cloneProxy, next); err != nil {
		balances := ai_convert.Balances()
		if len(balances) == 0 {
			body := ctx.Response().GetBody()
			if len(body) == 0 {
				ctx.Response().SetBody([]byte(err.Error()))
				ctx.Response().SetStatus(400, "Bad Request")
			}
			return err
		}
		err = e.doBalance(ctx, cloneProxy, next) // Fallback to balance logic
		if err != nil {
			ctx.Response().SetBody([]byte(err.Error()))
			ctx.Response().SetStatus(400, "Bad Request")
			return err
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
	extender, err := p.GenExtender(e.modelCfg)
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
	for _, r := range resources {
		if !r.Health() {
			continue
		}
		ctx.SetProxy(cloneProxy)
		ai_convert.SetAIKey(ctx, r.ID())
		if err = r.RequestConvert(ctx, extender); err != nil {
			ai_convert.SetAIProviderStatuses(ctx, ai_convert.AIProviderStatus{
				Provider: e.provider,
				Model:    e.model,
				Key:      r.ID(),
				Status:   ai_convert.StatusInvalid,
			})
			continue
		}

		if next != nil {
			if err = e.processNext(ctx, next, p); err != nil {
				if ctx.Response().StatusCode() == 504 {
					ai_convert.SetAIProviderStatuses(ctx, ai_convert.AIProviderStatus{
						Provider: e.provider,
						Model:    e.model,
						Key:      r.ID(),
						Status:   ai_convert.StatusTimeout,
					})
				}
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
		if err = r.ResponseConvert(ctx); err != nil {
			//return err
			log.Errorf("response convert error: %v", err)
			continue
		}
		aiStatus := ai_convert.GetAIStatus(ctx)
		switch aiStatus {
		case ai_convert.StatusInvalidRequest, ai_convert.StatusNormal:
			return nil
		default:
			continue
		}
	}
	return fmt.Errorf("all key resources for provider %s is invalid", e.provider)
}

// handleNoKeyResource handles the case when no key resources are available.
func (e *executor) handleNoKeyResource(ctx http_context.IHttpContext) error {
	err := fmt.Errorf("no key resource for provider %s", e.provider)
	ctx.Response().SetStatus(402, "Payment Required")
	ctx.Response().SetBody([]byte(err.Error()))
	return err
}

//// initializeExtender initializes the extender for a resource.
//func (e *executor) initializeExtender(resource ai_convert.IKeyResource) (map[string]interface{}, error) {
//	fn, has := resource.ConverterDriver().GetModel(e.model)
//	if !has {
//		return nil, fmt.Errorf("model %s not found", e.model)
//	}
//	return fn(e.modelCfg)
//}

func (e *executor) Destroy() {}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

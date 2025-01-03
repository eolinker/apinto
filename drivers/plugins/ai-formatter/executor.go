package ai_formatter

import (
	"errors"
	"fmt"

	"github.com/eolinker/apinto/convert"
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
	providers := convert.Providers()
	for _, provider := range providers {
		if !provider.Health() {
			continue
		}
		balanceHandler := provider.BalanceHandler()
		if balanceHandler != nil {
			ctx.SetBalance(balanceHandler)
		}
		if err := e.tryProvider(ctx, originProxy, next, provider); err == nil {
			return nil
		}
		provider.Down() // Mark the provider as unhealthy
	}

	return errors.New("all providers exhausted or unavailable")
}

// tryProvider attempts to use a single provider and its resources for processing.
func (e *executor) tryProvider(ctx http_context.IHttpContext, originProxy http_context.IRequest, next eocontext.IChain, provider convert.IProvider) error {
	convert.SetAIProvider(ctx, provider.Provider())
	convert.SetAIModel(ctx, provider.Model())
	extender := provider.ModelConfig()
	resources, has := convert.KeyResources(e.provider)
	if !has {
		return errKeyNotFound
	}
	for _, resource := range resources {

		converter, has := resource.ConverterDriver().GetConverter(provider.Model())
		if !has {
			return errors.New("invalid model")
		}

		ctx.SetProxy(originProxy)
		if err := converter.RequestConvert(ctx, extender); err != nil {
			continue
		}

		if next != nil {
			if err := e.processNext(ctx, next, provider); err != nil {
				return err
			}
		}

		if err := converter.ResponseConvert(ctx); err != nil {
			return err
		}

		e.handleAIStatus(ctx, resource)
		return nil
	}

	return errors.New("provider exhausted")
}

// processNext processes the next chain in the filter, handling 504 errors.
func (e *executor) processNext(ctx http_context.IHttpContext, next eocontext.IChain, provider convert.IProvider) error {
	if err := next.DoChain(ctx); err != nil {
		if ctx.Response().StatusCode() == 504 {
			convert.SetAIStatusTimeout(ctx)
			provider.Down() // Mark provider as unhealthy on timeout
		}
		return err
	}
	return nil
}

// handleAIStatus handles the status of the AI context and adjusts resource availability.
func (e *executor) handleAIStatus(ctx http_context.IHttpContext, resource convert.IKeyResource) {
	switch convert.GetAIStatus(ctx) {
	case convert.StatusInvalid, convert.StatusExpired, convert.StatusQuotaExhausted:
		resource.Down()
	case convert.StatusExceeded:
		resource.Breaker()
	}
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {

	cloneProxy := ctx.ProxyClone()
	convert.SetAIProvider(ctx, e.provider)
	convert.SetAIModel(ctx, e.model)

	if err := e.processKeyPool(ctx, cloneProxy, next); err != nil {
		return e.doBalance(ctx, cloneProxy, next) // Fallback to balance logic
	}
	return nil
}

// processKeyPool handles processing using the key pool resources.
func (e *executor) processKeyPool(ctx http_context.IHttpContext, cloneProxy http_context.IRequest, next eocontext.IChain) error {

	p, has := convert.GetProvider(e.provider)
	if !has {
		return errProviderNotFound
	}
	balanceHandler := p.BalanceHandler()
	if balanceHandler != nil {
		ctx.SetBalance(balanceHandler)
	}
	resources, has := convert.KeyResources(e.provider)
	if !has {
		return errKeyNotFound
	}
	var extender map[string]interface{}
	var err error
	for _, resource := range resources {
		if extender == nil {
			if extender, err = e.initializeExtender(resource); err != nil {
				return err
			}
		}

		converter, has := resource.ConverterDriver().GetConverter(e.model)
		if !has {
			return errors.New("invalid model")
		}

		ctx.SetProxy(cloneProxy)
		if err = converter.RequestConvert(ctx, extender); err != nil {
			continue
		}

		if next != nil {
			if err = e.processNext(ctx, next, p); err != nil {
				return err
			}
		}

		if err = converter.ResponseConvert(ctx); err != nil {
			return err
		}

		if e.isAIRequestValid(ctx) {
			return nil
		}
	}
	return fmt.Errorf("")
}

// handleNoKeyResource handles the case when no key resources are available.
func (e *executor) handleNoKeyResource(ctx http_context.IHttpContext) error {
	err := fmt.Errorf("no key resource for provider %s", e.provider)
	ctx.Response().SetStatus(402, "Payment Required")
	ctx.Response().SetBody([]byte(err.Error()))
	return err
}

// initializeExtender initializes the extender for a resource.
func (e *executor) initializeExtender(resource convert.IKeyResource) (map[string]interface{}, error) {
	fn, has := resource.ConverterDriver().GetModel(e.model)
	if !has {
		return nil, fmt.Errorf("model %s not found", e.model)
	}
	return fn(e.modelCfg)
}

// isAIRequestValid checks if the AI request is valid and should complete.
func (e *executor) isAIRequestValid(ctx http_context.IHttpContext) bool {
	switch convert.GetAIStatus(ctx) {
	case convert.StatusInvalidRequest, convert.StatusNormal:
		return true
	case convert.StatusInvalid, convert.StatusExpired, convert.StatusQuotaExhausted, convert.StatusExceeded:
		return false
	default:
		return false
	}
}

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

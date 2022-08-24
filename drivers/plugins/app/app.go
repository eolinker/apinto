package app

import (
	"errors"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type App struct {
	id string
}

func (a *App) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(a, ctx, next)
}

func (a *App) Destroy() {
}

func (a *App) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	err := a.auth(ctx)
	if err != nil {
		ctx.Response().SetStatus(403, "403")
		return err
	}
	if next != nil {
		err = next.DoChain(ctx)
	}
	if err != nil {
		return err
	}
	return nil
}

func (a *App) auth(ctx http_service.IHttpContext) error {
	driver := ctx.Request().Header().GetHeader("Authorization-Type")
	filters := appManager.ListByDriver(driver)
	if len(filters) < 1 && appManager.Count() > 0 {
		filters = appManager.List()
	}
	for _, filter := range filters {
		err := filter.Auth(ctx)
		if err == nil {
			return nil
		}
		log.DebugF("auth error: %s", err.Error())
	}
	return errors.New("invalid user")
}

func (a *App) Id() string {
	return a.id
}

func (a *App) Start() error {
	return nil
}

func (a *App) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (a *App) Stop() error {
	return nil
}

func (a *App) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

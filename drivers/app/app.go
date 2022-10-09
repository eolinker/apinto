package app

import (
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/eosc"
)

type app struct {
	id        string
	name      string
	driverIDs []string
	config    *Config
	executor  application.IAppExecutor
}

func (a *app) Execute(ctx http_service.IHttpContext) error {
	if a.executor == nil {
		return nil
	}
	return a.executor.Execute(ctx)
}

func (a *app) Name() string {
	return a.name
}

func (a *app) Labels() map[string]string {
	if a.config == nil {
		return nil
	}
	return a.config.Labels
}

func (a *app) Disable() bool {
	if a.config == nil {
		return true
	}
	return a.config.Disable
}

func (a *app) Destroy() error {
	appManager.Del(a.id)
	return nil
}

func (a *app) Id() string {
	return a.id
}

func (a *app) Start() error {
	if a.config == nil {
		return nil
	}
	return a.set(a.config)
}

func (a *app) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	err = a.set(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) set(cfg *Config) error {
	filters, users, err := createFilters(a.id, cfg.Auth)
	if err != nil {
		return err
	}

	//cfg.Labels["application"] = strings.TrimSuffix(app., "@app")

	appManager.Set(a, filters, users)
	e := newExecutor()
	e.append(newAdditionalParam(cfg.Additional))
	a.executor = e
	a.config = cfg
	return nil
}

func (a *app) Stop() error {
	return a.Destroy()
}

func (a *app) CheckSkill(skill string) bool {
	return false
}

func createFilters(id string, auths []*Auth) ([]application.IAuth, map[string][]application.ITransformConfig, error) {
	filters := make([]application.IAuth, 0, len(auths))
	userMap := make(map[string][]application.ITransformConfig)
	for _, v := range auths {

		filter, err := createFilter(v.Type, v.TokenName, v.Position, v.Config)
		if err != nil {
			return nil, nil, err
		}
		users := make([]application.ITransformConfig, 0, len(v.Users))
		for _, u := range v.Users {
			users = append(users, u)
		}
		err = checkUsers(id, filter, users)
		if err != nil {
			return nil, nil, err
		}
		filters = append(filters, filter)
		userMap[filter.ID()] = users
	}
	return filters, userMap, nil
}

func checkUsers(id string, filter application.IAuth, users []application.ITransformConfig) error {
	return filter.Check(id, users)
}

func createFilter(driver string, tokenName string, position string, rule interface{}) (application.IAuth, error) {
	fac, err := auth.GetFactory(driver)
	if err != nil {
		return nil, err
	}
	filter, err := fac.Create(tokenName, position, rule)
	if err != nil {
		return nil, err
	}
	old, has := appManager.Get(filter.ID())
	if !has {
		return filter, nil
	}

	return old, nil
}

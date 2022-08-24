package app

import (
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/eosc"
)

type app struct {
	id        string
	driverIDs []string
	config    *Config
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
	filters, users, err := createFilters(a.config.Auth)
	if err != nil {
		return err
	}
	appManager.Set(a.Id(), a.config.Labels, a.config.Disable, filters, users)
	
	return nil
}

func (a *app) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	err = set(a.id, cfg)
	if err != nil {
		return err
	}
	a.config = cfg
	return nil
}

func set(id string, cfg *Config) error {
	filters, users, err := createFilters(cfg.Auth)
	if err != nil {
		return err
	}
	appManager.Set(id, cfg.Labels, cfg.Disable, filters, users)
	return nil
}

func (a *app) Stop() error {
	return a.Destroy()
}

func (a *app) CheckSkill(skill string) bool {
	return true
}

func createFilters(auths []*Auth) ([]application.IAuth, map[string][]*application.User, error) {
	filters := make([]application.IAuth, 0, len(auths))
	userMap := make(map[string][]*application.User)
	for _, v := range auths {
		filter, err := createFilter(v.Type, v.TokenName, v.Position, v.Users, v.Config)
		if err != nil {
			return nil, nil, err
		}
		filters = append(filters, filter)
		userMap[filter.ID()] = v.Users
	}
	return filters, userMap, nil
}

func createFilter(driver string, tokenName string, position string, users []*application.User, rule interface{}) (application.IAuth, error) {
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
	
	return filter, old.Check(users)
}

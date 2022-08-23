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
	
	return nil
}

func (a *app) Id() string {
	return a.id
}

func (a *app) Start() error {
	if a.config == nil {
		return nil
	}
	filters, err := createFilters(a.config.Auth)
	if err != nil {
		return err
	}
	
	return nil
}

func (a *app) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (a *app) Stop() error {
	return a.Destroy()
}

func (a *app) CheckSkill(skill string) bool {
	return true
}

func createFilters(auths []*Auth) ([]application.IAuth, error) {
	filters := make([]application.IAuth, 0, len(auths))
	for _, v := range auths {
		filter, err := createFilter(v.Type, v.TokenName, v.Position, v.Users, v.Config)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	return filters, nil
}

func createFilter(driver string, tokenName string, position string, users []*application.User, rule interface{}) (application.IAuth, error) {
	fac, err := auth.GetFactory(driver)
	if err != nil {
		return nil, err
	}
	filter, err := fac.Create(tokenName, position, users, rule)
	if err != nil {
		return nil, err
	}
	old, has := appManager.Get(filter.ID())
	if !has {
		return filter, nil
	}
	
	return filter, old.Check(users)
}

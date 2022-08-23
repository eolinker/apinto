package app

import (
	"github.com/eolinker/apinto/drivers/app/manager"
	"github.com/eolinker/eosc"
)

type app struct {
	id         string
	driverIDs  []string
	config     *Config
	appManager manager.IManager
}

func (a *app) Destroy() error {
	ids := a.driverIDs
	a.del(ids)
	return nil
}

func (a *app) Id() string {
	return a.id
}

func (a *app) Start() error {
	if a.config == nil {
		return nil
	}
	ids := make([]string, 0, len(a.config.Auth))
	for _, auth := range a.config.Auth {
		id, err := a.appManager.Set(a.id, auth.Type, auth.TokenName, auth.Position, auth.Users, auth.Config)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	a.driverIDs = ids
	return nil
}

func (a *app) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	ids := a.driverIDs
	a.del(ids)
	
	newIDs := make([]string, 0, len(cfg.Auth))
	for _, auth := range cfg.Auth {
		id, err := a.appManager.Set(a.id, auth.Type, auth.TokenName, auth.Position, auth.Users, auth.Config)
		if err != nil {
			a.del(newIDs)
			return err
		}
		newIDs = append(ids, id)
	}
	a.driverIDs = newIDs
	return nil
}

func (a *app) del(ids []string) {
	for _, id := range ids {
		a.appManager.DelByAppID(id, a.id)
	}
}

func (a *app) Stop() error {
	return a.Destroy()
}

func (a *app) CheckSkill(skill string) bool {
	return true
}

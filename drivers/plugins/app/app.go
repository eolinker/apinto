package app

import (
	"github.com/eolinker/apinto/auth"
	"github.com/eolinker/eosc"
)

type App struct {
	auths []auth.IAuth
}

func (a *App) Id() string {
	//TODO implement me
	panic("implement me")
}

func (a *App) Start() error {
	//TODO implement me
	panic("implement me")
}

func (a *App) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	//TODO implement me
	panic("implement me")
}

func (a *App) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (a *App) CheckSkill(skill string) bool {
	//TODO implement me
	panic("implement me")
}

package app

import (
	"github.com/eolinker/apinto/drivers/app/manager"
	"github.com/eolinker/eosc"
)

var (
	appManager manager.IManager
)

type App struct {
	id string
}

func (a *App) Id() string {
	return a.id
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

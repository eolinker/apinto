package service

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type IService interface {
	eosc.IWorker
	eoscContext.CompleteHandler
	eoscContext.FinishHandler
}

type ITemplate interface {
	eosc.IWorker
	Create(id string, conf map[string]*plugin.Config) eoscContext.IChain
}

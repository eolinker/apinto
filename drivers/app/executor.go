package app

import (
	"github.com/eolinker/apinto/application"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

const (
	conflictConvert = "convert"
	conflictError   = "error"
	conflictOrigin  = "origin"
)

var (
	errorExist = "%s: %s is already exists"
)

type executor struct {
	executors []application.IAppExecutor
}

func newExecutor() *executor {
	return &executor{executors: make([]application.IAppExecutor, 0, 3)}
}

func (e *executor) Execute(ctx http_service.IHttpContext) error {
	for _, f := range e.executors {
		err := f.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *executor) append(f ...application.IAppExecutor) {
	e.executors = append(e.executors, f...)
}

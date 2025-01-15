package script_handler

import (
	"fmt"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Script struct {
	drivers.WorkerBase
	stage string
	fn    func(ctx http_service.IHttpContext) error
}

func (a *Script) Destroy() {
	a.fn = nil
	return
}

// 拦截请求过滤，内部转换为http类型再处理
func (a *Script) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(a, ctx, next)
}

// 插件添加时执行
func (a *Script) Start() error {
	return nil
}

// 插件被修改时执行
func (a *Script) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func getFunc(conf *Config) (func(http_service.IHttpContext) error, error) {
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}

	i := interp.New(interp.Options{})
	err = i.Use(stdlib.Symbols)
	if err != nil {
		return nil, err
	}
	_, err = i.Eval(conf.Script)
	if err != nil {
		return nil, err
	}
	v, err := i.Eval(conf.Package + "." + conf.Fname)
	if err != nil {
		return nil, err
	}
	fn, ok := v.Interface().(func(http_service.IHttpContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid function")
	}
	return fn, nil
}

// 插件删除时间执行
func (a *Script) Stop() error {
	return nil
}

func (a *Script) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

// DoHttpFilter 核心http处理方法
func (a *Script) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("script invoke error: %v", err)
		}
	}()
	if a.stage == "" || a.stage == "request" {
		err := a.fn(ctx)
		if err != nil {
			log.Errorf("exec request script error: %s", err.Error())
			return err
		}
	}

	err := next.DoChain(ctx)
	if err != nil {
		return err
	}

	if a.stage == "response" {
		err = a.fn(ctx)
		if err != nil {
			log.Errorf("exec response script error: %s", err.Error())
			return err
		}
	}
	return nil
}

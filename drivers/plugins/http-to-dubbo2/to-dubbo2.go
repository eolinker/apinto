package http_to_dubbo2

import (
	"encoding/json"
	"errors"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"time"
)

var _ eocontext.IFilter = (*ToDubbo2)(nil)
var _ http_context.HttpFilter = (*ToDubbo2)(nil)

type ToDubbo2 struct {
	drivers.WorkerBase
	service string
	method  string
	params  []param
}

func (p *ToDubbo2) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {

	var err error
	defer func() {
		if err != nil {
			ctx.Response().SetStatus(400, "400")
			ctx.Response().SetBody([]byte(err.Error()))
		}
	}()

	body, _ := ctx.Request().Body().RawBody()

	var types []string
	var valuesList []hessian.Object

	for _, v := range p.params {
		types = append(types, v.className)
	}

	//从body中提取内容
	if len(p.params) == 1 && p.params[0].fieldName == "" {
		var val interface{}

		if err = json.Unmarshal(body, &val); err != nil {
			log.Errorf("doHttpFilter jsonUnmarshal err:%v body:%v", err, body)
			return err
		}

		valuesList = append(valuesList, val)
	} else if len(p.params) == 1 && p.params[0].fieldName != "" {
		var maps map[string]interface{}

		if err = json.Unmarshal(body, &maps); err != nil {
			log.Errorf("doHttpFilter jsonUnmarshal err:%v body:%v", err, body)
			return err
		}

		if val, ok := maps[p.params[0].fieldName]; ok {
			valuesList = append(valuesList, val)
		} else {
			err = errors.New(fmt.Sprintf("参数解析错误，body中未包含%s的参数名", p.params[0].fieldName))
			return err
		}

	} else {
		var maps map[string]interface{}

		if err = json.Unmarshal(body, &maps); err != nil {
			log.Errorf("doHttpFilter jsonUnmarshal err:%v body:%v", err, body)
			return err
		}

		for _, v := range p.params {
			if val, ok := maps[v.fieldName]; ok {
				valuesList = append(valuesList, val)
			} else {
				err = errors.New(fmt.Sprintf("参数解析错误，body中未包含%s的参数名", p.params[0].fieldName))
				return err
			}
		}

	}

	client := newDubbo2Client(p.service, p.method, types, valuesList)
	complete := NewComplete(0, time.Second*30, client)
	ctx.SetCompleteHandler(complete)

	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}

func (p *ToDubbo2) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(p, ctx, next)
}

type param struct {
	className string
	fieldName string
}

func (p *ToDubbo2) Start() error {
	return nil
}

func (p *ToDubbo2) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
	p.service = conf.Service
	p.method = conf.Method

	params := make([]param, 0, len(conf.Params))

	for _, val := range conf.Params {
		params = append(params, param{
			className: val.ClassName,
			fieldName: val.FieldName,
		})
	}
	p.params = params
	return nil
}

func (p *ToDubbo2) Stop() error {
	return nil
}

func (p *ToDubbo2) Destroy() {
}

func (p *ToDubbo2) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

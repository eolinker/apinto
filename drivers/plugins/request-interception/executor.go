package request_interception

import (
	"strconv"
	"strings"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/drivers"
)

var _ eocontext.IFilter = (*executor)(nil)
var _ http_service.HttpFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	status      int
	body        string
	headers     []*Header
	contentType string
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	ctx.Response().SetStatus(e.status, strconv.Itoa(e.status))
	ctx.Response().SetBody([]byte(e.body))
	entry := ctx.GetEntry()
	for _, header := range e.headers {
		value := ""
		for _, v := range header.Value {
			if strings.HasPrefix(v, "$") {
				// 变量，从请求中获取
				value += eosc.ReadStringFromEntry(entry, v[1:])
			} else {
				value += v
			}
		}
		ctx.Response().SetHeader(header.Key, value)
	}
	ctx.Response().SetHeader("Content-Type", e.contentType)
	return nil
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) Destroy() {
	return
}

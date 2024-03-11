package js_inject

import (
	"fmt"
	"strings"

	"github.com/eolinker/apinto/utils"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ eocontext.IFilter = (*executor)(nil)
var _ http_service.HttpFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	injectCode       string
	matchContentType []string
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if next != nil {
		err = next.DoChain(ctx)
		if err != nil {
			return err
		}
	}
	contentType := ctx.Response().ContentType()
	for _, ct := range e.matchContentType {
		if contentType == ct {
			body := ctx.Response().GetBody()
			res, err := injectJavaScript(string(body), e.injectCode)
			if err != nil {
				log.Error(err)
				return nil
			}
			ctx.Response().SetBody([]byte(res))
		}
	}
	return
}

func (e *executor) Destroy() {
	return
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) reset(conf *Config) error {
	decodeData, err := utils.B64Decode(conf.InjectCode)
	if err != nil {
		return err
	}
	injectCode := string(decodeData)
	for _, v := range conf.Variables {
		injectCode = strings.Replace(injectCode, fmt.Sprintf("{{%s}}", v.Key), v.Value, -1)
	}
	e.injectCode = fmt.Sprintf("%s", injectCode)
	e.matchContentType = conf.MatchContentType
	return nil
}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

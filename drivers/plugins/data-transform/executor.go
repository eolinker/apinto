package data_transform

import (
	"mime"
	"net/http"
	"strings"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	conf *Config
}

func (b *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(b, ctx, next)
}

func (b *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	if b.conf.RequestTransform && (ctx.Proxy().Method() == http.MethodPost || ctx.Proxy().Method() == http.MethodPut || ctx.Proxy().Method() == http.MethodPatch) {
		// 对请求体做转换
		body, _ := ctx.Proxy().Body().RawBody()
		contentType, _, _ := mime.ParseMediaType(ctx.Request().ContentType())
		if strings.Contains(contentType, "/json") {
			result, err := json2xml(body, b.conf.XMLRootTag, b.conf.XMLDeclaration)
			if err != nil {
				errInfo := "fail to transform request json to xml"
				ctx.Response().SetStatus(http.StatusBadRequest, "400")
				ctx.Response().SetBody([]byte(encode(b.conf.ErrorType, errInfo, http.StatusBadRequest)))
				log.Errorf("%s,body is %s", errInfo, string(body))
				return err
			}
			ctx.Proxy().Body().SetRaw("application/xml", result)
		} else if strings.Contains(contentType, "/xml") {
			result, err := xml2json(body, b.conf.XMLDeclaration)
			if err != nil {
				errInfo := "fail to transform request xml to json"
				ctx.Response().SetStatus(http.StatusBadRequest, "400")
				ctx.Response().SetBody([]byte(encode(b.conf.ErrorType, errInfo, http.StatusBadRequest)))
				log.Errorf("%s,body is %s", errInfo, string(body))
				return err
			}
			ctx.Proxy().Body().SetRaw("application/json", result)
		}
	}
	err := next.DoChain(ctx)
	if err != nil {
		return err
	}
	if b.conf.ResponseTransform {
		// 对请求体做转换
		body := ctx.Response().GetBody()
		contentType, _, _ := mime.ParseMediaType(ctx.Response().ContentType())
		if strings.Contains(contentType, "/json") {
			result, err := json2xml(body, b.conf.XMLRootTag, b.conf.XMLDeclaration)
			if err != nil {
				errInfo := "fail to transform response json to xml"
				ctx.Response().SetStatus(http.StatusBadRequest, "400")
				ctx.Response().SetBody([]byte(encode(b.conf.ErrorType, errInfo, http.StatusBadRequest)))
				log.Errorf("%s,body is %s", errInfo, string(body))
				return err
			}
			ctx.Response().SetBody(result)
			ctx.Response().Headers().Set("Content-Type", "application/xml")
		} else if strings.Contains(contentType, "/xml") {
			result, err := xml2json(body, b.conf.XMLDeclaration)
			if err != nil {
				errInfo := "fail to transform response xml to json"
				ctx.Response().SetStatus(http.StatusBadRequest, "400")
				ctx.Response().SetBody([]byte(encode(b.conf.ErrorType, errInfo, http.StatusBadRequest)))
				log.Errorf("%s,body is %s", errInfo, string(body))
				return err
			}
			ctx.Response().SetBody(result)
			ctx.Response().Headers().Set("Content-Type", "application/json")
		}
	}
	return nil
}

func (b *executor) Start() error {
	return nil
}

func (b *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (b *executor) Stop() error {
	b.Destroy()
	return nil
}

func (b *executor) Destroy() {
	b.conf = nil
	return
}

func (b *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

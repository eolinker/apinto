package ip_restriction

import (
	"encoding/json"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*IPHandler)(nil)
var _ eocontext.IFilter = (*IPHandler)(nil)

type IPHandler struct {
	drivers.WorkerBase
	responseType string
	filter       IPFilter
}

func (I *IPHandler) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(I, ctx, next)
}

func (I *IPHandler) doRestriction(ctx http_service.IHttpContext) error {
	realIP := ctx.Request().RealIp()
	if I.filter != nil {
		ok, err := I.filter(realIP)
		if !ok {
			return err
		}
	}
	return nil
}

func (I *IPHandler) Start() error {
	return nil
}

func (I *IPHandler) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	confObj, err := check(conf)
	if err != nil {
		return err
	}
	I.filter = confObj.genFilter()
	return nil
}

func (I *IPHandler) Stop() error {
	return nil
}

func (I *IPHandler) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func (I *IPHandler) responseEncode(origin string, statusCode int) string {
	if I.responseType == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": statusCode,
		}
		newInfo, _ := json.Marshal(tmp)
		return string(newInfo)
	}
	return origin
}
func (I *IPHandler) Destroy() {
	I.filter = nil
	I.responseType = ""
}

func (I *IPHandler) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	err := I.doRestriction(ctx)
	if err != nil {
		resp := ctx.Response()
		info := I.responseEncode(err.Error(), 403)
		resp.SetStatus(403, "403")
		resp.SetBody([]byte(info))
		return err
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

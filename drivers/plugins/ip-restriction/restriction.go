package ip_restriction

import (
	"encoding/json"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
)

type IPHandler struct {
	*Driver
	id    string
	name  string
	responseType string
	filter IPFilter
}

func (I *IPHandler) doRestriction(ctx http_service.IHttpContext) error {
	realIP := ctx.Request().ReadIP()
	if I.filter != nil {
		ok, err :=  I.filter(realIP)
		if !ok {
			return err
		}
	}
	return nil
}

func (I *IPHandler) Id() string {
	return I.id
}

func (I *IPHandler) Start() error {
	return nil
}

func (I *IPHandler) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	confObj, err := I.check(conf)
	if err != nil {
		return err
	}
	I.filter = confObj.genFilter()
	I.responseType = confObj.ResponseType
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

func (I *IPHandler) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) error {
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

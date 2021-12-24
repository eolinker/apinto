package response_rewrite

import (
	"strconv"

	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/utils"
)

type ResponseRewrite struct {
	*Driver
	id         string
	name       string
	statusCode int
	body       string
	headers    map[string]string
	match      *MatchConf
}

func (r *ResponseRewrite) Id() string {
	return r.id
}

func (r *ResponseRewrite) Start() error {
	return nil
}

func (r *ResponseRewrite) Reset(v interface{}, workers map[eosc.RequireId]interface{}) error {
	conf, err := r.check(v)
	if err != nil {
		return err
	}

	//若body非空且需要base64转码
	if conf.Body != "" && conf.BodyBase64 {
		conf.Body, err = utils.B64Decode(conf.Body)
		if err != nil {
			return err
		}
	}

	r.statusCode = conf.StatusCode
	r.body = conf.Body
	r.headers = conf.Headers
	r.match = conf.Match

	return nil
}

func (r *ResponseRewrite) Stop() error {
	return nil
}

func (r *ResponseRewrite) Destroy() {
	r.statusCode = 0
	r.body = ""
	r.headers = nil
	r.match = nil
}

func (r *ResponseRewrite) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func (r *ResponseRewrite) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	if next != nil {
		err = next.DoChain(ctx)
	}

	return r.rewrite(ctx)
}

func (r *ResponseRewrite) rewrite(ctx http_service.IHttpContext) error {
	//匹配状态码
	if !r.matchStatusCode(ctx.Response().StatusCode()) {
		return nil
	}

	//重写响应状态码
	if r.statusCode != 0 {
		ctx.Response().SetStatus(r.statusCode, strconv.Itoa(r.statusCode))
	}

	//重写body
	if r.body != "" {
		ctx.Response().SetBody([]byte(r.body))
	}

	//新增header
	for k, v := range r.headers {
		ctx.Response().SetHeader(k, v)
	}

	return nil
}

func (r *ResponseRewrite) matchStatusCode(code int) bool {
	for _, c := range r.match.Code {
		if c == code {
			return true
		}
	}

	return false
}

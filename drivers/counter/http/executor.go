package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/ohler55/ojg/jp"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/counter"
	"github.com/ohler55/ojg/oj"
	"github.com/valyala/fasthttp"

	"github.com/eolinker/eosc"
)

var _ counter.IClient = (*Executor)(nil)
var _ eosc.IWorker = (*Executor)(nil)

type Executor struct {
	drivers.WorkerBase
	request *fasthttp.Request
	expr    jp.Expr
}

func (b *Executor) Start() error {
	return nil
}

func (b *Executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type,id is %s", b.Id())
	}

	return b.reset(cfg)
}

func (b *Executor) reset(conf *Config) error {
	expr, err := jp.ParseString(conf.ResponseJsonPath)
	if err != nil {
		return err
	}
	request := fasthttp.AcquireRequest()
	request.SetRequestURI(conf.URI)
	request.Header.SetMethod(conf.Method)
	for key, value := range conf.Headers {
		request.Header.Set(key, value)
	}
	for key, value := range conf.QueryParam {
		request.URI().QueryArgs().Set(key, value)
	}
	if conf.ContentType == "json" {
		request.Header.SetContentType("application/json")
		body, _ := json.Marshal(conf.BodyParam)
		request.SetBody(body)
	} else {
		request.Header.SetContentType("application/x-www-form-urlencoded")
		bodyParams := url.Values{}
		for key, value := range conf.BodyParam {
			bodyParams.Set(key, value)
		}
		request.SetBodyString(bodyParams.Encode())
	}
	b.request = request
	b.expr = expr
	scope_manager.Set(b.Id(), b, conf.Scopes...)
	return nil
}

func (b *Executor) Stop() error {
	fasthttp.ReleaseRequest(b.request)
	scope_manager.Del(b.Id())
	return nil
}

func (b *Executor) CheckSkill(skill string) bool {
	return counter.FilterSkillName == skill
}

var httpClient = fasthttp.Client{
	Name: "apinto-counter",
}

func (b *Executor) Get(key string) (int64, error) {
	req := fasthttp.AcquireRequest()
	b.request.CopyTo(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)

	req.URI().SetQueryStringBytes(bytes.Replace(req.URI().QueryString(), []byte("$key"), []byte(key), -1))
	req.SetBody(bytes.Replace(req.Body(), []byte("$key"), []byte(key), -1))

	err := httpClient.Do(req, resp)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return 0, fmt.Errorf("error http status code: %d,key: %s,id: %s", resp.StatusCode(), key, b.Id())
	}
	result, err := oj.Parse(resp.Body())
	if err != nil {
		return 0, err
	}
	// 解析JSON
	v := b.expr.Get(result)
	if v == nil || len(v) < 1 {
		return 0, fmt.Errorf("no found key: %s,id: %s", key, b.Id())
	}
	if len(v) != 1 {
		return 0, fmt.Errorf("invalid value: %v,key: %s,id: %s", v, key, b.Id())
	}
	return v[0].(int64), nil
}

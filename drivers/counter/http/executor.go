package http

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ohler55/ojg/oj"

	"github.com/ohler55/ojg/jp"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/counter"
	"github.com/valyala/fasthttp"

	"github.com/eolinker/eosc"
)

var _ counter.IClient = (*Executor)(nil)
var _ eosc.IWorker = (*Executor)(nil)

type Executor struct {
	drivers.WorkerBase
	req         *fasthttp.Request
	contentType string
	query       map[string]string
	header      map[string]string
	body        map[string]string
	expr        jp.Expr
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

	b.contentType = conf.ContentType
	b.header = conf.Headers
	b.query = conf.QueryParam
	b.body = conf.BodyParam

	b.req = request
	b.expr = expr
	scope_manager.Set(b.Id(), b, conf.Scopes...)
	return nil
}

func (b *Executor) Stop() error {
	fasthttp.ReleaseRequest(b.req)
	scope_manager.Del(b.Id())
	return nil
}

func (b *Executor) CheckSkill(skill string) bool {
	return counter.FilterSkillName == skill
}

var httpClient = fasthttp.Client{
	Name: "apinto-counter",
}

func retrieveValue(variables map[string]string, value string) string {
	if !strings.HasPrefix(value, "$") {
		return value
	}
	v, ok := variables[value]
	if !ok {
		v = value
	}
	return v
}

func (b *Executor) Get(variables map[string]string) (int64, error) {
	req := fasthttp.AcquireRequest()
	b.req.CopyTo(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	for key, value := range b.header {
		req.Header.Set(key, retrieveValue(variables, value))
	}
	for key, value := range b.query {
		req.URI().QueryArgs().Set(key, retrieveValue(variables, value))
	}

	var body []byte
	switch b.contentType {
	case "json":
		tmp := make(map[string]string)
		for key, value := range b.body {
			tmp[key] = retrieveValue(variables, value)
		}
		body, _ = json.Marshal(tmp)
		req.Header.SetContentType("application/json")
	case "form-data":
		params := url.Values{}
		for key, value := range b.body {
			params.Add(key, retrieveValue(variables, value))
		}
		body = []byte(params.Encode())
		req.Header.SetContentType("application/x-www-form-urlencoded")
	}

	req.SetBody(body)
	err := httpClient.DoTimeout(req, resp, 10*time.Second)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return 0, fmt.Errorf("http status code is %d", resp.StatusCode())
	}
	result, err := oj.Parse(resp.Body())
	if err != nil {
		return 0, err
	}
	// 解析JSON
	v := b.expr.Get(result)
	if v == nil || len(v) < 1 {
		return 0, fmt.Errorf("json path %s not found,id is %s", b.expr.String(), b.Id())
	}
	if len(v) != 1 {
		return 0, fmt.Errorf("json path %s found more than one,id is %s", b.expr.String(), b.Id())
	}
	intV, ok := v[0].(int64)
	if !ok {
		return 0, fmt.Errorf("json path %s found not int64,id is %s", b.expr.String(), b.Id())
	}
	return intV, nil
}

package service_http

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/eolinker/goku-eosc/service"

	upstream_http "github.com/eolinker/goku-eosc/upstream/upstream-http"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery/static"
	"github.com/eolinker/goku-eosc/upstream"
	round_robin "github.com/eolinker/goku-eosc/upstream/round-robin"
)

type routerDemo struct {
	location string
	host     string
	header   map[string]string
	query    url.Values
}

func (r *routerDemo) Location() string {
	return r.location
}

func (r *routerDemo) Host() string {
	return r.host
}

func (r *routerDemo) Header() map[string]string {
	return r.header
}

func (r *routerDemo) Query() url.Values {
	return r.query
}

func TestService(t *testing.T) {
	round_robin.Register()

	staticConfig := &upstream_http.Config{
		Name:      "product-user",
		Driver:    "http_proxy",
		Desc:      "生产环境-用户模块",
		Scheme:    "http",
		Type:      "round-robin",
		Config:    "127.0.0.1:8580 weight=10;10.1.1.1:8080 weight=20",
		Discovery: "static_1@discovery",
	}

	staticWorker, err := getWorker(static.NewFactory(), &static.Config{
		Name:   "static_1",
		Driver: "static",
		Labels: nil,
		Health: &static.HealthConfig{
			Protocol:    "http",
			Method:      "POST",
			URL:         "/Web/Test/params/print",
			SuccessCode: 200,
			Period:      30,
			Timeout:     3000,
		},
		HealthOn: true,
	}, "discovery", "static", "", "静态服务发现", nil, "", "static_1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	allWorker := make(map[eosc.RequireId]interface{})
	allWorker["static_1@discovery"] = staticWorker
	upstreamWorker, err := getWorker(upstream_http.NewFactory(), staticConfig, "upstream", "http_proxy", "", "http转发驱动", nil, "", "product-user", allWorker)
	if err != nil {
		t.Error(err)
		return
	}
	upstreamWorker.Start()
	hUpstream, ok := upstreamWorker.(upstream.IUpstream)
	if !ok {
		t.Error(ErrorStructType)
		return

	}
	allWorker["product-user@upstream"] = hUpstream

	serviceWorker, err := getWorker(NewFactory(), &Config{
		Name:       "guest",
		Driver:     "http",
		Desc:       "游客服务",
		Timeout:    30000,
		Retry:      2,
		Scheme:     "http",
		RewriteURL: "/Web/Test",
		Upstream:   "product-user@upstream",
	}, "service", "http", "", "http服务驱动", nil, "", "guest", allWorker)
	if err != nil {
		t.Error(err)
		return
	}
	serv, ok := serviceWorker.(service.IService)
	if !ok {
		t.Error(ErrorStructType)
		return
	}
	data := url.Values{}
	data.Set("name", "eolinker")
	r, err := http.NewRequest("POST", "http://localhost:8080/product/params/print", strings.NewReader(data.Encode()))
	if err != nil {
		t.Error(err)
		return
	}
	serv.Handle(&response{}, r, &routerDemo{
		location: "/product",
		host:     "localhost:8080",
		header:   nil,
		query:    nil,
	})
}

func getWorker(factory eosc.IProfessionDriverFactory, cfg interface{}, profession string, name string, label string, desc string, params map[string]string, workerID, workerName string, worker map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	driver, err := factory.Create(profession, name, label, desc, params)
	if err != nil {
		return nil, err
	}

	return driver.Create(workerID, workerName, cfg, worker)
}

type response struct {
}

func (r *response) Header() http.Header {
	panic("implement me")
}

func (r *response) Write(bytes []byte) (int, error) {
	panic("implement me")
}

func (r *response) WriteHeader(statusCode int) {
	panic("implement me")
}

package upstream_http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/goku-eosc/service"

	round_robin "github.com/eolinker/goku-eosc/upstream/round-robin"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/upstream"

	discovery_static "github.com/eolinker/goku-eosc/discovery/static"

	"github.com/eolinker/eosc"
)

type Service struct {
	name    string
	desc    string
	retry   int
	timeout time.Duration
	scheme  string
	addr    string
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Desc() string {
	return s.desc
}

func (s *Service) Retry() int {
	return s.retry
}

func (s *Service) Timeout() time.Duration {
	return s.timeout
}

func (s *Service) Scheme() string {
	return s.scheme
}

func (s *Service) ProxyAddr() string {
	return s.ProxyAddr()
}

func TestSend(t *testing.T) {
	round_robin.Register()
	s := &Service{
		name:    "参数打印",
		desc:    "打印所有参数",
		retry:   3,
		timeout: time.Second * 10,
		scheme:  "http",
	}
	factory := NewFactory()
	t.Log("upstream extend info:", factory.ExtendInfo())
	driver, err := factory.Create("upstream", "http_proxy", "", "http转发驱动", nil)
	if err != nil {
		t.Error(err)
		return
	}
	cfg := &Config{
		Name:      "product-user",
		Driver:    "http_proxy",
		Desc:      "生产环境-用户模块",
		Scheme:    "http",
		Type:      "round-robin",
		Config:    "127.0.0.1:8580 weight=10;47.95.203.198:8080 weight=15",
		Discovery: "static_1@discovery",
	}

	staticDiscovery := discovery_static.NewFactory()
	t.Log("static discovery extend info:", staticDiscovery.ExtendInfo())
	staticDriver, err := staticDiscovery.Create("discovery", "static", "", "静态服务发现驱动", nil)
	if err != nil {
		t.Error(err)
		return
	}
	staticCfg := &discovery_static.Config{
		Name:   "static_1",
		Driver: "static",
		Labels: nil,
		Health: &discovery_static.HealthConfig{
			Protocol:    "http",
			Method:      "GET",
			URL:         "/",
			SuccessCode: 404,
			Period:      30,
			Timeout:     3000,
		},
		HealthOn: true,
	}
	staticWorker, err := staticDriver.Create("", "static_1", staticCfg, nil)
	if err != nil {
		t.Error(err)
		return
	}

	worker, err := driver.Create(
		"",
		"product-user",
		cfg,
		map[eosc.RequireId]interface{}{
			"static_1@discovery": staticWorker,
		})
	if err != nil {
		t.Error(err)
		return
	}
	worker.Start()
	hUpstream, ok := worker.(upstream.IUpstream)
	if !ok {
		t.Error(ErrorStructType)
		return
	}
	data := url.Values{}
	data.Set("name", "eolinker")
	r, err := http.NewRequest("POST", "http://localhost:8080/Web/Test/params/print", strings.NewReader(data.Encode()))
	if err != nil {
		t.Error(ErrorStructType)
		return
	}

	ctx := http_context.NewContext(r, &response{})
	// 设置目标URL
	ctx.ProxyRequest.SetTargetURL(r.URL.Path)
	for i := 0; i < 10; i++ {
		now := time.Now()
		err = send(ctx, s, hUpstream)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("spend time is", time.Now().Sub(now))
	}
}

func send(ctx *http_context.Context, s service.IServiceDetail, hUpstream upstream.IUpstream) error {
	resp, err := hUpstream.Send(ctx, s)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
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

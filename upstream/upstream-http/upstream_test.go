package upstream_http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/goku-eosc/discovery/consul"
	"github.com/eolinker/goku-eosc/discovery/eureka"

	round_robin "github.com/eolinker/goku-eosc/upstream/round-robin"

	"github.com/eolinker/goku-eosc/discovery/static"

	"github.com/eolinker/goku-eosc/service"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
)

var (
	ErrorStructType = "error struct type: %s, need struct type: %s"
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

func getWorker(factory eosc.IProfessionDriverFactory, cfg interface{}, profession string, name string, label string, desc string, params map[string]string, workerID, workerName string, worker map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	driver, err := factory.Create(profession, name, label, desc, params)
	if err != nil {
		return nil, err
	}

	return driver.Create(workerID, workerName, cfg, worker)
}

var s = &Service{
	name:    "参数打印",
	desc:    "打印所有参数",
	retry:   3,
	timeout: time.Second * 10,
	scheme:  "http",
}

func TestConsul(t *testing.T) {
	round_robin.Register()
	consulConfig := &Config{
		Name:      "product-user",
		Driver:    "http_proxy",
		Desc:      "生产环境-用户模块",
		Scheme:    "http",
		Type:      "round-robin",
		Config:    "consul",
		Discovery: "consul_1@discovery",
	}

	consulWorker, err := getWorker(consul.NewFactory(), &consul.Config{
		Name:   "consul_1",
		Driver: "consul",
		Labels: map[string]string{
			"scheme": "http",
		},
		Config: consul.AccessConfig{
			Address: []string{"39.108.94.48:8500", "39.108.94.48:8501"},
			Params:  map[string]string{"token": "a92316d8-5c99-4fa0-b4cd-30b9e66718aa"},
		},
	}, "discovery", "consul", "", "consul", nil, "", "consul_1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	consulWorker.Start()
	allWorker := make(map[eosc.RequireId]interface{})
	allWorker["consul_1@discovery"] = consulWorker
	worker, err := getWorker(NewFactory(), consulConfig, "upstream", "http_proxy", "", "http转发驱动", nil, "", "product-user", allWorker)
	if err != nil {
		t.Error(err)
		return
	}

	hUpstream, ok := worker.(upstream.IUpstream)
	if !ok {
		t.Error(ErrorStructType)

	}
	data := url.Values{}
	data.Set("name", "eolinker")
	r, err := http.NewRequest("POST", "http://localhost:8080/Web/Test/params/print", strings.NewReader(data.Encode()))
	if err != nil {
		t.Error(ErrorStructType)
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

func TestEureka(t *testing.T) {
	round_robin.Register()
	eurekaConfig := &Config{
		Name:      "product-user",
		Driver:    "http_proxy",
		Desc:      "生产环境-用户模块",
		Scheme:    "http",
		Type:      "round-robin",
		Config:    "DEMO",
		Discovery: "eureka_1@discovery",
	}

	eurekaWorker, err := getWorker(eureka.NewFactory(), &eureka.Config{
		Name:   "eureka_1",
		Driver: "eureka",
		Labels: map[string]string{
			"scheme": "http",
		},
		Config: eureka.AccessConfig{
			Address: []string{"39.108.94.48:8761/eureka"},
			Params: map[string]string{
				"username": "test",
				"password": "test"},
		},
	}, "discovery", "eureka", "", "eureka", nil, "", "eureka_1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	eurekaWorker.Start()
	allWorker := make(map[eosc.RequireId]interface{})
	allWorker["eureka_1@discovery"] = eurekaWorker
	worker, err := getWorker(NewFactory(), eurekaConfig, "upstream", "http_proxy", "", "http转发驱动", nil, "", "product-user", allWorker)
	if err != nil {
		t.Error(err)
		return
	}

	hUpstream, ok := worker.(upstream.IUpstream)
	if !ok {
		t.Error(ErrorStructType)

	}
	data := url.Values{}
	data.Set("name", "eolinker")
	r, err := http.NewRequest("POST", "http://localhost:8080/Web/Test/params/print", strings.NewReader(data.Encode()))
	if err != nil {
		t.Error(ErrorStructType)
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

func TestStatic(t *testing.T) {
	round_robin.Register()
	staticConfig := &Config{
		Name:      "product-user",
		Driver:    "http_proxy",
		Desc:      "生产环境-用户模块",
		Scheme:    "http",
		Type:      "round-robin",
		Config:    "127.0.0.1:8580 weight=10;47.95.203.198:8080 weight=20",
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
	worker, err := getWorker(NewFactory(), staticConfig, "upstream", "http_proxy", "", "http转发驱动", nil, "", "product-user", allWorker)
	if err != nil {
		t.Error(err)
		return
	}

	hUpstream, ok := worker.(upstream.IUpstream)
	if !ok {
		t.Error(ErrorStructType)

	}
	data := url.Values{}
	data.Set("name", "eolinker")
	r, err := http.NewRequest("POST", "http://localhost:8080/Web/Test/params/print", strings.NewReader(data.Encode()))
	if err != nil {
		t.Error(ErrorStructType)
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

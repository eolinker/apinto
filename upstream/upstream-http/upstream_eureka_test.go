package upstream_http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/goku-eosc/discovery/nacos"

	"github.com/eolinker/goku-eosc/discovery/eureka"

	round_robin "github.com/eolinker/goku-eosc/upstream/round-robin"

	"github.com/eolinker/goku-eosc/discovery/static"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
)

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

func TestNacos(t *testing.T) {
	round_robin.Register()
	nacosConfig := &Config{
		Name:      "product-user",
		Driver:    "http_proxy",
		Desc:      "生产环境-用户模块",
		Scheme:    "http",
		Type:      "round-robin",
		Config:    "nacos.naming.serviceName",
		Discovery: "nacos_1@discovery",
	}

	nacosWorker, err := getWorker(nacos.NewFactory(), &nacos.Config{
		Name:   "nacos_1",
		Driver: "nacos",
		Labels: map[string]string{
			"scheme": "http",
		},
		Config: nacos.AccessConfig{
			Address: []string{"39.108.94.48:8848"},
			Params: map[string]string{
				"username":   "test",
				"password":   "test",
				"healthOnly": "false",
			},
		},
	}, "discovery", "nacos", "", "nacos", nil, "", "nacos_1", nil)
	if err != nil {
		t.Error(err)
		return
	}
	nacosWorker.Start()
	allWorker := make(map[eosc.RequireId]interface{})
	allWorker["nacos_1@discovery"] = nacosWorker
	worker, err := getWorker(NewFactory(), nacosConfig, "upstream", "http_proxy", "", "http转发驱动", nil, "", "product-user", allWorker)
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

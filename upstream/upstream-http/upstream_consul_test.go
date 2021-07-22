package upstream_http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/goku-eosc/discovery/consul"

	round_robin "github.com/eolinker/goku-eosc/upstream/round-robin"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
)

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
			Address: []string{"10.1.94.48:8500", "10.1.94.48:8501"},
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

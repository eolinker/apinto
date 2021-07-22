package upstream_http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/goku-eosc/discovery/nacos"

	round_robin "github.com/eolinker/goku-eosc/upstream/round-robin"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
)

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
			Address: []string{"10.1.94.48:8848"},
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

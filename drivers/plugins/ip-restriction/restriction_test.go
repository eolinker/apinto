package ip_restriction

import (
	"fmt"
	http_service "github.com/eolinker/eosc/http-service"
	http_context "github.com/eolinker/goku/node/http-context"
	"github.com/valyala/fasthttp"
	"net"
	"testing"
)

// 127.0.0.1:8080
var ctx http_service.IHttpContext

func getContext() (http_service.IHttpContext, error) {
	if ctx == nil {
		return nil, fmt.Errorf("please init test context")
	}
	return ctx, nil
}
func initTestContext(address string) error {
	fast := &fasthttp.RequestCtx{}
	freq := fasthttp.AcquireRequest()
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}
	fast.Init(freq, addr, nil)
	ctx = http_context.NewContext(fast)
	return nil
}
func TestDoRestriction(t *testing.T) {
	http_ctx, err := getContext()

	if err != nil {
		t.Fatal(err)
	}
	f := NewFactory()
	d, err := f.Create("plugin@setting", "ip_restriction", "ip_restriction", "service", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "limit_black_all",
			config: &Config{
				IPListType: "black",
				IPBlackList: []string{
					"*",
				},
			},
			want: "403",
		},
		{
			name: "limit_black",
			config: &Config{
				IPListType: "black",
				IPBlackList: []string{
					"127.0.0.1",
				},
			},
			want: "403",
		},
		{
			name: "pass_black",
			config: &Config{
				IPListType: "black",
				IPBlackList: []string{
					"127.0.0.2",
				},
			},
			want: "200",
		},
		{
			name: "limit_white",
			config: &Config{
				IPListType: "white",
				IPWhiteList: []string{
					"127.0.0.1",
				},
			},
			want: "200",
		},
		{
			name: "pass_white",
			config: &Config{
				IPListType: "white",
				IPWhiteList: []string{
					"127.0.0.2",
				},
			},
			want: "403",
		},
		{
			name: "pass_white_all",
			config: &Config{
				IPListType: "white",
				IPWhiteList: []string{
					"*",
				},
			},
			want: "200",
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			http_ctx.Response().SetStatus(200, "200")
			ip, err := d.Create("ip_restriction@plugin", "ip_restriction", cc.config, nil)
			if err != nil {
				t.Errorf("create handler error : %v", err)
			}
			h, ok := ip.(http_service.IFilter)
			if !ok {
				t.Errorf("parse filter error")
				return
			}
			h.DoFilter(http_ctx, nil)
			if http_ctx.Response().Status() != cc.want {
				t.Errorf("do restriction error; want %s, got %s", cc.want, http_ctx.Response().Status())
			}
		})
	}
}

func TestMain(m *testing.M) {
	err := initTestContext("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	m.Run()
}

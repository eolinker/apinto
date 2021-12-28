package gzip

import (
	"fmt"
	http_service "github.com/eolinker/eosc/http-service"
	http_context "github.com/eolinker/goku/node/http-context"
	"github.com/valyala/fasthttp"
	"net"
	"testing"
)

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

func TestMain(m *testing.M) {
	err := initTestContext("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestFilter(t *testing.T) {
	http_ctx, err := getContext()
	if err != nil {
		t.Fatal(err)
	}
	f := NewFactory()
	d, err := f.Create("plugin@setting", "ip_restriction", "ip_restriction", "service", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	g, err := d.Create("gzip@plugin", "gzip", &Config{Types: nil, MinLength: 10, Vary: true}, nil)
	if err != nil {
		t.Errorf("create handler error : %v", err)
	}
	h, ok := g.(http_service.IFilter)
	if !ok {
		t.Errorf("parse filter error")
		return
	}

	cases := []struct {
		name         string
		header       string
		body         string
		wantCompress bool
	}{
		{
			name:         "wantCompress",
			wantCompress: true,
			body:         "eolink;goku;apinto;test;gzip;eolink;goku;apinto;test;gzip;eolink;goku;apinto;test;gzip;eolink;goku;apinto;test;gzip;eolink;goku;apinto;test;gzip;eolink;goku;apinto;test;gzip;",
			header:       "gzip",
		},
		{
			name:         "notCompress",
			wantCompress: false,
			body:         "eolink",
			header:       "",
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctx.Response().SetBody([]byte(cc.body))
			ctx.Request().Header().Headers().Set("Accept-Encoding", cc.header)
			before := ctx.Response().BodyLen()
			h.DoFilter(http_ctx, nil)
			after := ctx.Response().BodyLen()
			if cc.wantCompress && before <= after {
				t.Errorf("want compress; before %d, after %d", before, after)
			}
			if !cc.wantCompress && before != after {
				t.Errorf("do not want compress; before %d, after %d", before, after)
			}
		})
	}
}

package auth

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/auth"
	"github.com/eolinker/goku/drivers/auth/apikey"
	http_context "github.com/eolinker/goku/node/http-context"
	"github.com/valyala/fasthttp"
	"net"
	"testing"
)

var ctx http_service.IHttpContext

func getContext(address string) (http_service.IHttpContext, error) {
	if ctx == nil {
		return initTestContext(address)
	}
	if address == ctx.Request().RemoteAddr() {
		return ctx, nil
	}
	return initTestContext(address)
}
func initTestContext(address string) (http_service.IHttpContext, error) {
	fast := &fasthttp.RequestCtx{}
	freq := fasthttp.AcquireRequest()
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	fast.Init(freq, addr, nil)
	return http_context.NewContext(fast), nil
}

type workers struct {
	data eosc.IUntyped
}

func (w *workers) Get(id string) (eosc.IWorker, bool) {
	wo, has := w.data.Get(id)
	if has {
		return wo.(eosc.IWorker), true
	}
	return nil, false
}
func createWorkers() eosc.IWorkers {
	worker := &workers{data: eosc.NewUntyped()}
	// 写入待应用的鉴权worker
	ad, _ := apikey.NewFactory().Create("apikey@auth", "apikey", "apikey", "service", map[string]interface{}{})
	aconf := &apikey.Config{
		Driver:          "apikey",
		HideCredentials: true,
		User: []apikey.User{
			{Apikey: "eolink", Expire: 0},
			{Apikey: "apinto", Expire: 0},
		},
	}
	wo, _ := ad.Create("apikey_test@apikey", "apikey_test", aconf, nil)
	worker.data.Set(wo.Id(), wo)
	return worker
}

func TestMain(m *testing.M) {
	w := createWorkers()
	bean.Injection(&w)
	bean.Check()
	m.Run()
}

func TestFilter(t *testing.T) {
	context, err := getContext("127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	f := NewFactory()
	d, err := f.Create("auth@setting", "auth", "auth", "service", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	g, err := d.Create("auth@plugin", "auth", &Config{Auth: []eosc.RequireId{"apikey_test@apikey"}}, nil)
	if err != nil {
		t.Errorf("create handler error : %v", err)
	}
	h, ok := g.(http_service.IFilter)
	if !ok {
		t.Errorf("parse filter error")
		return
	}
	cases := []struct {
		name      string
		authType  string
		authValue string
		wantErr   bool
	}{
		{
			name:      "no_auth_header",
			authType:  "",
			authValue: "",
			wantErr:   true,
		},
		{
			name:      "pass_auth_apikey",
			authType:  "apikey",
			authValue: "eolink",
			wantErr:   false,
		},
		{
			name:      "intercept_auth_apikey",
			authType:  "apikey",
			authValue: "goku",
			wantErr:   true,
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			context.Request().Header().Headers().Set(auth.AuthorizationType, cc.authType)
			context.Request().Header().Headers().Set(auth.Authorization, cc.authValue)
			context.Proxy().Header().SetHeader(auth.AuthorizationType, cc.authType)
			context.Proxy().Header().SetHeader(auth.Authorization, cc.authValue)
			err = h.DoFilter(context, nil)
			if cc.wantErr {
				if err == nil {
					t.Errorf("auth plugins error;")
				}
			} else {
				if err != nil {
					t.Errorf("auth plugins error;")
				}
			}
		})
	}
}

package fasthttp_client

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestMyselfProxyTimeout(t *testing.T) {
	//addr := "https://gw.kuaidizs.cn"
	addr := "http://127.0.0.1:8099"
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.URI().SetPath("/open/api")
	req.URI().SetQueryString("test=1")
	req.URI().SetHost("127.0.0.1:8099")
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	t.Log(string(req.URI().RequestURI()), req.URI().String(), string(req.URI().Host()), string(req.URI().Scheme()))
	req.SetBody([]byte(`{"cpCode":"YTO","province":"广东省","city":"广州市"}`))
	err := defaultClient.ProxyTimeout(addr, "", req, resp, 0)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(resp.Body()))
	return
}

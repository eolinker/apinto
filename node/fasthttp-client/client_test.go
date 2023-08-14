package fasthttp_client

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestProxyTimeout(t *testing.T) {
	//addr := "https://gw.kuaidizs.cn"
	addr := fmt.Sprintf("%s://%s", "https", "gw.kuaidizs.cn")
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.URI().SetPath("/open/api")
	req.URI().SetHost("gw.kuaidizs.cn")
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody([]byte(`{"cpCode":"YTO","version":"1.0","timestamp":"2023-08-10 11:57:13","province":"广东省","city":"广州市","appKey":"DBB812347A1E44829159FE82F5C4303E","format":"json","sign_method":"md5","method":"kdzs.address.reachable","sign":"10A4B5A59340F9B98DAFA3CFCCF65449"}`))
	err := defaultClient.ProxyTimeout(addr, req, resp, 0)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(resp.Body()))
	return
}

func TestMyselfProxyTimeout(t *testing.T) {
	//addr := "https://gw.kuaidizs.cn"
	addr := "http://127.0.0.1:8099"
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.URI().SetPath("/open/api")
	req.URI().SetHost("127.0.0.1:8099")
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody([]byte(`{"cpCode":"YTO","province":"广东省","city":"广州市"}`))
	err := defaultClient.ProxyTimeout(addr, req, resp, 0)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(resp.Body()))
	return
}

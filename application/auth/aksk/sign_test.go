package aksk

import (
	"net/url"
	"testing"

	http_context "github.com/eolinker/apinto/node/http-context"
	"github.com/valyala/fasthttp"
)

func TestAksk_Check(t *testing.T) {
	req := fasthttp.AcquireRequest()
	u := fasthttp.AcquireURI()
	uri, _ := url.Parse("https://gateway.hr-soft.cn/api/redapi/tech/redteams/tag/tagUser?userId=8e211885-b622-11ec-8e03-14187749a8c0&tagId=37")
	u.SetPath(uri.Path)
	u.SetScheme(uri.Scheme)
	u.SetHost(uri.Host)
	u.SetQueryString(uri.RawQuery)
	req.SetURI(u)
	req.SetHostBytes(u.Host())
	req.Header.Set("x-gateway-date", "20200605T104456Z")
	req.Header.Set("content-type", "application/json")
	req.Header.SetHostBytes(u.Host())

	ctx := &fasthttp.RequestCtx{
		Request: *req,
	}
	origin := buildToSign(http_context.NewContext(ctx, 9000), "HMAC-SHA256", []string{"content-type", "host", "x-gateway-date"})
	t.Log("origin", origin)
	sign := hMaxBySHA256("8f8154ff07f7153eea59a2ba44b5fcfe443dba1e4c45f87c549e6a05f699145d", origin)
	if sign != "523a1d901873fbbc19df355dba6f8bb695a2ef3435206615e7724490054b0529" {
		t.Fatalf("sign error %s", sign)
	}
	t.Logf("sign %s", sign)
}

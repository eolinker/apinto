package counter

import (
	"fmt"

	"github.com/ohler55/ojg/oj"

	"github.com/ohler55/ojg/jp"
	"github.com/valyala/fasthttp"
)

var _ IClient = (*HTTPClient)(nil)

var httpClient = fasthttp.Client{
	Name: "apinto-counter",
}

type HTTPClient struct {
	uri     string
	headers map[string]string
	// jsonExpr 经过编译的JSONPath表达式
	jsonExpr jp.Expr
}

func NewHTTPClient(uri string, jsonExpr jp.Expr) *HTTPClient {
	return &HTTPClient{uri: uri, jsonExpr: jsonExpr}
}

func (H *HTTPClient) Get(key string) (int64, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(H.uri)
	req.Header.SetMethod("GET")
	for name, value := range H.headers {
		req.Header.Set(name, value)
	}
	err := httpClient.Do(req, resp)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return 0, fmt.Errorf("error http status code: %d,key: %s,uri %s", resp.StatusCode(), key, H.uri)
	}
	result, err := oj.Parse(resp.Body())
	if err != nil {
		return 0, err
	}
	// 解析JSON
	v := H.jsonExpr.Get(result)
	if v == nil || len(v) < 1 {
		return 0, fmt.Errorf("no found key: %s,uri: %s", key)
	}
	if len(v) != 1 {
		return 0, fmt.Errorf("invalid value: %v,key: %s,uri: %s", v, key, H.uri)
	}
	return v[0].(int64), nil
}

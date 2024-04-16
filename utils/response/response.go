package response

import (
	"fmt"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/metrics"
)

type IResponse interface {
	Response(ctx eocontext.EoContext)
}

type Response struct {
	StatusCode  int      `json:"status_code" label:"HTTP状态码"`
	ContentType string   `json:"content_type" label:"Content-Type"`
	Charset     string   `json:"charset" label:"Charset"`
	Headers     []Header `json:"headers" label:"Header参数"` //key:value
	Body        string   `json:"body" label:"Body"`
}

type Header struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

type header struct {
	key   metrics.Metrics
	value metrics.Metrics
}

func Parse(c *Response) IResponse {
	if c == nil {
		return nil
	}
	hm := make([]header, 0, len(c.Headers))
	for _, h := range c.Headers {
		hm = append(hm, header{
			key:   metrics.Parse(h.Key),
			value: metrics.Parse(h.Value),
		})
	}
	return &responseHandler{
		status:      c.StatusCode,
		contentType: metrics.Parse(c.ContentType),
		charset:     metrics.Parse(c.Charset),
		headers:     hm,
		body:        metrics.Parse(c.Body),
	}
}

type responseHandler struct {
	status      int
	contentType metrics.Metrics
	charset     metrics.Metrics
	headers     []header
	body        metrics.Metrics
}

func (r *responseHandler) Response(ctx eocontext.EoContext) {
	httpCtx, err := http_context.Assert(ctx)
	if err != nil {
		return
	}
	httpCtx.Response().SetStatus(r.status, "")
	for _, h := range r.headers {
		k := h.key.Metrics(httpCtx)
		v := h.value.Metrics(httpCtx)
		httpCtx.Response().SetHeader(k, v)
	}
	httpCtx.Response().SetHeader("Content-Type", fmt.Sprintf("%s; charset=%s", r.contentType.Metrics(httpCtx), r.charset.Metrics(httpCtx)))
	httpCtx.Response().SetBody([]byte(r.body.Metrics(httpCtx)))
}

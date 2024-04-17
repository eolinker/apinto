package response

import (
	"fmt"
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/metrics"
	"net/http"
)

type IResponse interface {
	Response(ctx eocontext.EoContext)
}

type Response struct {
	StatusCode  int      `json:"status_code" label:"HTTP状态码"`
	ContentType string   `json:"content_type" label:"Content-Type"`
	Charset     string   `json:"charset" label:"Charset"`
	Headers     []Header `json:"headers" label:"Header参数"` //key:value
	Body        string   `json:"body" label:"Body" description:"body模版, 支持 ${label} 语法"`
}

type Header struct {
	Key   string `json:"key" yaml:"key" label:"header key" description:"header 的key,支持 ${label}"`
	Value string `json:"value" yaml:"value" label:"header value" description:"header 的值, 支持${label}"`
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
	entry := http_entry.NewEntry(httpCtx)
	httpCtx.Response().SetStatus(r.status, http.StatusText(r.status))
	for _, h := range r.headers {
		k := h.key.Metrics(entry)
		v := h.value.Metrics(entry)
		httpCtx.Response().SetHeader(k, v)
	}
	httpCtx.Response().SetHeader("Content-Type", fmt.Sprintf("%s; charset=%s", r.contentType.Metrics(entry), r.charset.Metrics(entry)))
	httpCtx.Response().SetBody([]byte(r.body.Metrics(entry)))
}

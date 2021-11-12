package http_context

import (
	"net/http"
	"net/url"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IRequestReader = (*RequestReader)(nil)

type RequestReader struct {
	req         *fasthttp.Request
	bodyHandler *BodyRequestHandler
	remoteAddr  string
	clientIP    string
	host        string
	method      string
	rawBody     []byte
	headers     http.Header
	scheme      string
	uri         *url.URL
	contentType string
}

func NewRequestReader(req *fasthttp.Request, remoteAddr string) *RequestReader {
	return &RequestReader{req: req, remoteAddr: remoteAddr}
}

func (r *RequestReader) ContentType() string {
	if r.contentType == "" {
		r.contentType = string(r.req.Header.ContentType())
	}
	return r.contentType
}

func (r *RequestReader) BodyForm() (url.Values, error) {
	if r.bodyHandler == nil {
		r.bodyHandler = newBodyRequestHandler(r.ContentType(), r.req.Body())
	}
	return r.bodyHandler.BodyForm()
}

func (r *RequestReader) Files() (map[string]*http_service.FileHeader, error) {
	if r.bodyHandler == nil {
		r.bodyHandler = newBodyRequestHandler(r.ContentType(), r.req.Body())
	}
	return r.bodyHandler.Files()
}

func (r *RequestReader) GetForm(key string) string {
	if r.bodyHandler == nil {
		r.bodyHandler = newBodyRequestHandler(r.ContentType(), r.req.Body())
	}
	return r.bodyHandler.GetForm(key)
}

func (r *RequestReader) GetFile(key string) (file *http_service.FileHeader, has bool) {
	if r.bodyHandler == nil {
		r.bodyHandler = newBodyRequestHandler(r.ContentType(), r.req.Body())
	}
	return r.bodyHandler.GetFile(key)
}

func (r *RequestReader) RawBody() ([]byte, error) {
	if r.bodyHandler == nil {
		r.bodyHandler = newBodyRequestHandler(r.ContentType(), r.req.Body())
	}
	return r.bodyHandler.RawBody()
}

func (r *RequestReader) GetHeader(name string) string {
	return r.Headers().Get(name)
}

func (r *RequestReader) Headers() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
		hs := strings.Split(r.req.Header.String(), "\r\n")
		for _, h := range hs {
			vs := strings.Split(h, ":")
			if len(vs) < 2 {
				if vs[0] == "" {
					continue
				}
				r.headers[vs[0]] = []string{""}
				continue
			}
			r.headers[vs[0]] = []string{strings.TrimSpace(vs[1])}

		}
	}
	return r.headers
}

func (r *RequestReader) Method() string {
	if r.method == "" {
		r.method = string(r.req.Header.Method())
	}
	return r.method
}

func (r *RequestReader) URL() url.URL {
	if r.uri == nil {
		r.uri, _ = url.Parse(r.req.URI().String())
	}
	return *r.uri
}

func (r *RequestReader) RequestURI() string {
	return string(r.req.RequestURI())
}

func (r *RequestReader) Host() string {
	if r.host == "" {
		r.host = strings.Split(string(r.req.Header.Host()), ":")[0]
	}
	return r.host
}

func (r *RequestReader) RemoteAddr() string {
	if r.clientIP == "" {
		clientIP := string(r.req.Header.Peek("X-Forwarded-For"))
		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}
		clientIP = strings.TrimSpace(clientIP)
		if len(clientIP) < 1 {
			clientIP = strings.TrimSpace(string(r.req.Header.Peek("X-Real-Ip")))
			if len(clientIP) < 1 {
				clientIP = r.remoteAddr
			}
		}
		r.clientIP = clientIP
	}
	return r.clientIP
}

func (r *RequestReader) Scheme() string {
	if r.scheme == "" {
		r.scheme = string(r.req.URI().Scheme())
	}
	return r.scheme
}

func (r *RequestReader) Request() *fasthttp.Request {
	return r.req
}

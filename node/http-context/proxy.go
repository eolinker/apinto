package http_context

import (
	"net/url"

	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/http-service"
)

var _ http_service.IRequest = (*ProxyRequest)(nil)

type ProxyRequest struct {
	*RequestReader
}

func (r *ProxyRequest) SetPath(s string) {
	r.Request().URI().SetPath(s)
}

func NewProxyRequest(request *fasthttp.Request) *ProxyRequest {
	return &ProxyRequest{
		RequestReader: NewRequestReader(request, ""),
	}
}

func (r *ProxyRequest) SetHeader(key, value string) {
	if r.headers != nil {
		r.headers.Set(key, value)
	}
	r.Request().Header.Set(key, value)
}

func (r *ProxyRequest) AddHeader(key, value string) {
	if r.headers != nil {
		r.headers.Add(key, value)
	}
	r.Request().Header.Add(key, value)
}

func (r *ProxyRequest) DelHeader(key string) {
	if r.headers != nil {
		r.headers.Del(key)
	}
	r.Request().Header.Del(key)
}

func (r *ProxyRequest) SetForm(values url.Values) error {
	r.form = values
	return nil
}

func (r *ProxyRequest) SetToForm(key, value string) error {
	if r.form == nil {
		form, err := r.BodyForm()
		if err != nil {
			return err
		}
		r.form = form
	}
	r.form.Set(key, value)
	return nil
}

func (r *ProxyRequest) AddForm(key, value string) error {
	if r.form == nil {
		form, err := r.BodyForm()
		if err != nil {
			return err
		}
		r.form = form
	}
	r.form.Set(key, value)
	return nil
}

func (r *ProxyRequest) AddFile(key string, file *http_service.FileHeader) error {
	if r.form == nil {
		file, err := r.Files()
		if err != nil {
			return err
		}
		r.files = file
	}
	r.files[key] = file
	return nil
}

func (r *ProxyRequest) SetRaw(contentType string, body []byte) {
	r.contentType, r.rawBody = contentType, body
}

func (r *ProxyRequest) SetMethod(s string) {
	r.Request().Header.SetMethod(s)
}

func (r *ProxyRequest) SetScheme(scheme string) {
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	r.Request().URI().SetScheme(scheme)
}

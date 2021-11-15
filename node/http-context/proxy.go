package http_context

import (
	"net/http"
	"net/url"

	http_service "github.com/eolinker/eosc/http-service"
)

var _ http_service.IRequest = (*ProxyRequest)(nil)

type ProxyRequest struct {
	*RequestReader
	headers     http.Header
	form        url.Values
	file        map[string]*http_service.FileHeader
	contentType string
	body        []byte
	uri         *url.URL
	method      string
}

func (r *ProxyRequest) SetPath(s string) {
	r.initUrl()
	r.uri.Path = s
}

func NewProxyRequest(requestReader *RequestReader) *ProxyRequest {
	return &ProxyRequest{RequestReader: requestReader}
}

func (r *ProxyRequest) SetHeader(key, value string) {
	if r.headers == nil {
		r.headers = r.Headers()
	}
	r.headers.Set(key, value)
}

func (r *ProxyRequest) AddHeader(key, value string) {
	if r.headers == nil {
		r.headers = r.Headers()
	}
	r.headers.Add(key, value)
}

func (r *ProxyRequest) DelHeader(key string) {
	if r.headers == nil {
		r.headers = r.Headers()
	}
	r.headers.Del(key)
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
		r.file = file
	}
	r.file[key] = file
	return nil
}

func (r *ProxyRequest) SetRaw(contentType string, body []byte) {
	r.contentType, r.body = contentType, body
}

func (r *ProxyRequest) TargetServer() string {
	r.initUrl()
	return r.uri.Host
}

func (r *ProxyRequest) TargetURL() string {
	r.initUrl()
	return r.uri.Path
}

func (r *ProxyRequest) initUrl() {
	if r.uri == nil {
		uri := r.URL()
		r.uri = &uri
	}
}

func (r *ProxyRequest) SetMethod(s string) {
	r.method = s
}

func (r *ProxyRequest) SetURL(url url.URL) {
	r.uri = &url
}

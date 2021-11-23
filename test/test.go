package test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	http_service "github.com/eolinker/eosc/http-service"
	uuid "github.com/satori/go.uuid"
)

type IConfig interface {
	Create() http_service.IHttpContext
}
type RequestConfig struct {
	ContentType string
	Body        []byte
	Header      http.Header
}
type ResponseConfig struct {
	Body          []byte
	ContendType   string
	Header        http.Header
	StatusCode    int
	ResponseError error
}

func (r *ResponseConfig) Response(request *RequestReader) (*Response, error) {

	header := r.Header
	if header == nil {
		header = make(http.Header)
	}
	header.Set("content-type", r.ContendType)
	header.Set("content-length", strconv.Itoa(len(r.Body)))

	return &Response{
		ResponseHeader: &ResponseHeader{header: r.Header},
		body:           r.Body,
		statusCode:     r.StatusCode,
	}, nil
}

type Config struct {
	Method string
	Url    string

	Request *RequestConfig

	Response ResponseConfig
}

func (c *Config) Create() http_service.IHttpContext {
	requestUri, err := url.Parse(c.Url)
	if err != nil {
		panic(err)
	}
	c.Request.Header.Set("contend-type", c.Response.ContendType)
	if c.Request == nil {
		c.Request = &RequestConfig{
			ContentType: "",
			Body:        nil,
			Header:      make(http.Header),
		}
	} else {
		if c.Request.Header == nil {
			c.Request.Header = make(http.Header)
		}
	}
	c.Request.Header.Set("contend-type", c.Response.ContendType)
	c.Request.Header.Set("content-length", strconv.Itoa(len(c.Request.Body)))
	return &Context{

		proxyRequest:  &ProxyRequest{},
		requestID:     uuid.NewV4().String(),
		response:      NewResponse(),
		responseError: nil,
		requestReader: &RequestReader{
			headers: &RequestHeader{
				header: c.Request.Header,
			},
			body: &BodyRequestHandler{
				contentType: c.Request.ContentType,
				raw:         c.Request.Body,
			},
			uri: &URIRequest{
				url: requestUri,
			},
			remoteAddr: "127.0.0.1",
			method:     c.Method,
		},
		ctx: nil,
	}
}
func NewGet(url string, response ResponseConfig) http_service.IHttpContext {
	c := &Config{
		Method:   http.MethodGet,
		Url:      url,
		Request:  nil,
		Response: response,
	}
	return c.Create()
}
func NewPostJson(url string, body interface{}, response ResponseConfig) http_service.IHttpContext {
	bodyData, _ := json.Marshal(body)

	c := &Config{
		Method: http.MethodPost,
		Url:    url,
		Request: &RequestConfig{
			ContentType: "application/json",
			Body:        bodyData,
			Header:      make(http.Header),
		},
		Response: response,
	}
	return c.Create()

}

func JsonResponse(body interface{}) *ResponseConfig {
	bodyData, _ := json.Marshal(body)
	return &ResponseConfig{
		Body:          bodyData,
		ContendType:   "application/json",
		Header:        make(http.Header),
		StatusCode:    200,
		ResponseError: nil,
	}
}

type PrintResponse struct {
}

func (p *PrintResponse) Response(request *RequestReader) (*Response, error) {
	body := make(map[string]interface{})
	body["url"] = request.uri.RequestURI()
	body["host"] = request.uri.Host()
	body["body"], _ = request.Body().RawBody()

	return JsonResponse(body).Response(request)
}

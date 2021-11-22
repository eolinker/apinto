package test

import (
	"net/http"
	"net/url"

	http_service "github.com/eolinker/eosc/http-service"
	uuid "github.com/satori/go.uuid"
)

type IConfig interface {
	Create() http_service.IHttpContext
}

type Config struct {
	Method  string
	Url     string
	Request struct {
		ContentType string
		Body        []byte
		Header      http.Header
	}

	Response struct {
		Body          []byte
		ContendType   string
		Header        http.Header
		StatusCode    int
		ResponseError error
	}
}

func (c *Config) Create() http_service.IHttpContext {
	requestUri, err := url.Parse(c.Url)
	if err != nil {
		panic(err)
	}
	c.Response.Header.Set("contend-type", c.Response.ContendType)
	return &Context{
		proxyRequest: &ProxyRequest{},
		requestID:    uuid.NewV4().String(),
		response: &Response{
			ResponseHeader: &ResponseHeader{header: c.Response.Header},
			body:           c.Response.Body,
			statusCode:     c.Response.StatusCode,
		},
		responseError: c.Response.ResponseError,
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
			remoteAddr: "",
			method:     "",
		},
		ctx: nil,
	}
}

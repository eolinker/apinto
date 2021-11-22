package test

import (
	"net/http"
)

type RequestHeader struct {
	host   string
	method string
	header http.Header
}

func (r *RequestHeader) SetHeader(key, value string) {

	r.header.Set(key, value)
}

func (r *RequestHeader) AddHeader(key, value string) {
	r.header.Add(key, value)
}

func (r *RequestHeader) DelHeader(key string) {
	r.header.Del(key)
}

func (r *RequestHeader) SetHost(host string) {
	r.host = host
}

func (r *RequestHeader) GetHeader(name string) string {
	return r.header.Get(name)
}

func (r *RequestHeader) Headers() http.Header {
	return r.header
}

func (r *RequestHeader) Host() string {
	return r.host
}

func (r *RequestHeader) Method() string {
	return r.method
}
func (r *RequestHeader) SetMethod(method string) {
	r.method = method
}

type ResponseHeader struct {
	header http.Header
}

func NewResponseHeader() *ResponseHeader {
	return &ResponseHeader{header: make(http.Header)}
}

func (r *ResponseHeader) GetHeader(name string) string {
	return r.header.Get(name)
}

func (r *ResponseHeader) Headers() http.Header {

	return r.header
}

func (r *ResponseHeader) SetHeader(key, value string) {

	r.header.Set(key, value)
}

func (r *ResponseHeader) AddHeader(key, value string) {

	r.header.Add(key, value)
}

func (r *ResponseHeader) DelHeader(key string) {

	r.header.Del(key)
}

package http_proxy

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/eolinker/goku/node/http-proxy/backend"
)

type response struct {
	body []byte
	resp *http.Response
}

//Body 响应体
func (r *response) Body() []byte {
	return r.body
}

//StatusCode 状态码
func (r *response) StatusCode() int {
	return r.resp.StatusCode
}

//Header 响应头部
func (r *response) Header() http.Header {
	return r.resp.Header
}

//Proto 协议
func (r *response) Proto() string {
	return r.resp.Proto
}

//Status 状态
func (r *response) Status() string {
	return r.resp.Status
}

//NewResponse 新建响应，返回IResponse节点
func NewResponse(resp *http.Response) (backend.IResponse, error) {
	defer resp.Body.Close()
	bd := resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		bd, _ = gzip.NewReader(resp.Body)
	}
	body, err := ioutil.ReadAll(bd)
	if err != nil {
		return nil, err
	}

	r := &response{
		body: body,
		resp: resp,
	}
	return r, nil
}

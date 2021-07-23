package http_proxy

import (
	"io/ioutil"
	"net/http"

	"github.com/eolinker/goku-eosc/node/http-proxy/backend"
)

type response struct {
	body []byte
	resp *http.Response
}

//Body 响应体
func (r *response) Body() []byte {
	// TODO: 该处可进行gzip等算法解压
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

//NewResponse 新建响应，返回IResponse节点
func NewResponse(resp *http.Response) (backend.IResponse, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r := &response{
		body: body,
		resp: resp,
	}
	return r, nil
}

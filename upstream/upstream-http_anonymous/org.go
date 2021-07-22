package upstream_http_anonymous

import (
	"fmt"
	"reflect"

	"github.com/eolinker/goku-eosc/node/http-proxy/backend"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/service"

	http_proxy "github.com/eolinker/goku-eosc/node/http-proxy"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/utils"
)

//Http org
type httpUpstream struct {
	id     string
	name   string
	driver string
}

//Id 返回worker id
func (h *httpUpstream) Id() string {
	return h.id
}

func (h *httpUpstream) Start() error {
	return nil
}

func (h *httpUpstream) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	return nil
}

func (h *httpUpstream) Stop() error {
	return nil
}

//CheckSkill 检查目标能力是否存在
func (h *httpUpstream) CheckSkill(skill string) bool {
	return upstream.CheckSkill(skill)
}

//send 请求发送，忽略重试
func (h *httpUpstream) Send(ctx *http_context.Context, serviceDetail service.IServiceDetail) (backend.IResponse, error) {
	var response backend.IResponse
	var err error
	path := utils.TrimPrefixAll(ctx.ProxyRequest.TargetURL(), "/")
	for doTrice := serviceDetail.Retry() + 1; doTrice > 0; doTrice-- {
		u := fmt.Sprintf("%s://%s/%s", serviceDetail.Scheme(), serviceDetail.ProxyAddr(), path)
		response, err = http_proxy.DoRequest(ctx, u, serviceDetail.Timeout())

		if err != nil {
			continue
		} else {
			return response, err
		}
	}

	return response, err
}

//GetType 获取匿名http_proxy负载配置的反射类型
func GetType() reflect.Type {
	return reflect.TypeOf((*httpUpstream)(nil))
}

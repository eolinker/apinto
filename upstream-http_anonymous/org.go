package upstream_http_anonymous

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/eolinker/goku-eosc/upstream"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/service"

	http_proxy "github.com/eolinker/goku-eosc/http-proxy"

	http_context "github.com/eolinker/eosc/node/http-context"

	"github.com/eolinker/eosc/utils"
)

//Http org
type httpUpstream struct {
	id     string
	name   string
	driver string
}

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

func (h *httpUpstream) CheckSkill(skill string) bool {
	return upstream.CheckSkill(skill)
}

//send 请求发送，忽略重试
func (h *httpUpstream) Send(ctx *http_context.Context, serviceDetail service.IServiceDetail) (*http.Response, error) {
	var response *http.Response
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

func GetType() reflect.Type {
	return reflect.TypeOf((*httpUpstream)(nil))
}

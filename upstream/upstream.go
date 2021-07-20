package upstream

import (
	"net/http"

	"github.com/eolinker/goku-eosc/service"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/goku-eosc/upstream.upstream.IUpstream"
}

type IUpstream interface {
	Send(ctx *http_context.Context, serviceDetail service.IServiceDetail) (*http.Response, error)
}

type Factory struct {
}

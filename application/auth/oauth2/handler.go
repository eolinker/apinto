package oauth2

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/eolinker/eosc/log"

	eoscContext "github.com/eolinker/eosc/eocontext"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type IHandler interface {
	Handle(ctx http_context.IHttpContext, client *Client, params url.Values)
}

type Handler struct {
	handler IHandler
}

func NewHandler(handler IHandler) *Handler {
	return &Handler{handler: handler}
}

func (h *Handler) Server(eoContext eoscContext.EoContext) (isContinue bool) {
	// 简化模式/授权码模式执行该流程
	ctx, err := http_context.Assert(eoContext)
	if err != nil {
		log.Errorf("assert http context error: %s", err)
		return true
	}
	params := retrieveParameters(ctx)
	clientId := params.Get("client_id")
	if clientId == "" {
		// 当空时视为正常请求，不做拦截
		return true
	}
	client, has := getClient(clientId)
	if !has {
		ctx.Response().SetBody([]byte("invalid client id"))
		ctx.Response().SetStatus(http.StatusNotFound, "not found")
		return false
	}

	if strings.ToUpper(ctx.Request().URI().Scheme()) != "HTTPS" && !client.AcceptHttpIfAlreadyTerminated {
		return false
	}
	if client.Expire > 0 && client.Expire < time.Now().Unix() {
		ctx.Response().SetBody([]byte("client id is expired"))
		ctx.Response().SetStatus(http.StatusForbidden, "forbidden")
		return false
	}
	if h.handler != nil {
		h.handler.Handle(ctx, client, params)
	}

	return false
}

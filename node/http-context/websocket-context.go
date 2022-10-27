package http_context

import (
	"errors"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/fasthttp/websocket"
)

var _ http_context.IWebsocketContext = (*WebsocketContext)(nil)

type WebsocketContext struct {
	*Context
	clientConn   *websocket.Conn
	upstreamConn *websocket.Conn
}

func (w *WebsocketContext) Upgrade() error {
	return nil
}

func (w *WebsocketContext) IsWebsocket() bool {
	return websocket.FastHTTPIsWebSocketUpgrade(w.fastHttpRequestCtx)
}

func NewWebsocketContext(ctx http_context.IHttpContext) (*WebsocketContext, error) {
	httpCtx, ok := ctx.(*Context)
	if !ok {
		return nil, errors.New("unsupported context type")
	}
	return &WebsocketContext{Context: httpCtx}, nil
}

func (w *WebsocketContext) GetClientConn() *websocket.Conn {
	return w.clientConn
}

func (w *WebsocketContext) GetUpstreamConn() *websocket.Conn {
	return w.upstreamConn
}

func (w *WebsocketContext) SetUpstreamConn(conn *websocket.Conn) {
	w.upstreamConn = conn
}

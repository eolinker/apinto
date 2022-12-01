package websocket

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"github.com/fasthttp/websocket"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

type Complete struct {
	retry   int
	timeOut time.Duration
}

func NewComplete(retry int, timeOut time.Duration) *Complete {
	return &Complete{retry: retry, timeOut: timeOut}
}

func (h *Complete) Complete(org eocontext.EoContext) error {
	ctx, err := http_service.WebsocketAssert(org)
	if err != nil {
		return err
	}

	balance := ctx.GetBalance()
	app := ctx.GetApp()

	scheme := app.Scheme()
	switch strings.ToLower(scheme) {
	case "http":
		scheme = "ws"
	case "https":
		scheme = "wss"
	default:
		//return fmt.Errorf("invalid scheme:%s", scheme)
		scheme = "ws"
	}

	proxyTime := time.Now()
	timeOut := app.TimeOut()
	var lastErr error
	var conn *websocket.Conn
	var resp *http.Response

	for index := 0; index <= h.retry; index++ {

		if h.timeOut > 0 && time.Now().Sub(proxyTime) > h.timeOut {
			return ErrorTimeoutComplete
		}
		node, err := balance.Select(ctx)
		if err != nil {
			log.Error("select error: ", lastErr)
			return err
		}

		log.Debug("node: ", node.Addr())
		u := url.URL{Scheme: "ws", Host: node.Addr(), Path: ctx.Proxy().URI().Path(), RawQuery: ctx.Proxy().URI().RawQuery()}
		conn, resp, lastErr = DialWithTimeout(u.String(), ctx.Proxy().Header().Headers(), timeOut)
		if lastErr == nil {
			resp.Body.Close()
			ctx.SetUpstreamConn(conn)
			break
		}
		log.Error("websocket upstream send error: ", lastErr)
	}
	if lastErr == nil {
		lastErr = ctx.Upgrade()
	}

	return lastErr
}

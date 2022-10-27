package websocket

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"github.com/fasthttp/websocket"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

//设置websocket
//CheckOrigin防止跨站点的请求伪造
var upgrader = websocket.FastHTTPUpgrader{}

type Complete struct {
	retry   int
	timeOut time.Duration
}

func NewComplete(retry int, timeOut time.Duration) *Complete {
	return &Complete{retry: retry, timeOut: timeOut}
}

func (h *Complete) Complete(org eocontext.EoContext) error {
	ctx, err := websocket_context.
	if err != nil {
		return err
	}
	conn, err := ctx.Upgrade()
	if err != nil {
		return err
	}
	balance := ctx.GetBalance()
	app := ctx.GetApp()
	var lastErr error
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
		addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
		lastErr = ctx.Dial(addr, timeOut)
		if lastErr == nil {
			return nil
		}
		log.Error("http upstream send error: ", lastErr)
	}

	return lastErr
}

type httpCompleteCaller struct {
}

func NewHttpCompleteCaller() *httpCompleteCaller {
	return &httpCompleteCaller{}
}

func (h *httpCompleteCaller) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return ctx.GetComplete().Complete(ctx)
}

func (h *httpCompleteCaller) Destroy() {

}

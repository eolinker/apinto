package websocket_context

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/fasthttp/websocket"

	websocket_context "github.com/eolinker/eosc/eocontext/websocket-context"

	"github.com/eolinker/eosc/utils/config"

	eoscContext "github.com/eolinker/eosc/eocontext"
	uuid "github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

var _ websocket_context.IWebsocketContext = (*Context)(nil)

//Context fasthttpRequestCtx
type Context struct {
	fastHttpRequestCtx  *fasthttp.RequestCtx
	requestID           string
	header              http.Header
	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	labels              map[string]string
	port                int
	connChan            chan *websocket.Conn
}

func (ctx *Context) Dial(address string, timeout time.Duration) error {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: timeout,
	}

	conn, _, err := dialer.Dial(address, ctx.header)
	if err != nil {
		return err
	}
	ctx.connChan <- conn
	return nil
}

func (ctx *Context) GetUpstreamHostHandler() eoscContext.UpstreamHostHandler {
	return ctx.upstreamHostHandler
}

func (ctx *Context) SetUpstreamHostHandler(handler eoscContext.UpstreamHostHandler) {
	ctx.upstreamHostHandler = handler
}

func (ctx *Context) LocalIP() net.IP {
	return ctx.fastHttpRequestCtx.LocalIP()
}

func (ctx *Context) LocalAddr() net.Addr {
	return ctx.fastHttpRequestCtx.LocalAddr()
}

func (ctx *Context) LocalPort() int {
	return ctx.port
}

func (ctx *Context) GetApp() eoscContext.EoApp {
	return ctx.app
}

func (ctx *Context) SetApp(app eoscContext.EoApp) {
	ctx.app = app
}

func (ctx *Context) GetBalance() eoscContext.BalanceHandler {
	return ctx.balance
}

func (ctx *Context) SetBalance(handler eoscContext.BalanceHandler) {
	ctx.balance = handler
}

func (ctx *Context) SetLabel(name, value string) {
	ctx.labels[name] = value
}

func (ctx *Context) GetLabel(name string) string {
	return ctx.labels[name]
}

func (ctx *Context) Labels() map[string]string {
	return ctx.labels
}

func (ctx *Context) GetComplete() eoscContext.CompleteHandler {
	return ctx.completeHandler
}

func (ctx *Context) SetCompleteHandler(handler eoscContext.CompleteHandler) {
	ctx.completeHandler = handler
}

func (ctx *Context) GetFinish() eoscContext.FinishHandler {
	return ctx.finishHandler
}

func (ctx *Context) SetFinish(handler eoscContext.FinishHandler) {
	ctx.finishHandler = handler
}

func (ctx *Context) Scheme() string {

	return string(ctx.fastHttpRequestCtx.Request.URI().Scheme())
}

func (ctx *Context) Assert(i interface{}) error {
	if v, ok := i.(*websocket_context.IWebsocketContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (ctx *Context) Context() context.Context {
	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}
	return ctx.ctx
}

func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Context().Value(key)
}

func (ctx *Context) WithValue(key, val interface{}) {
	ctx.ctx = context.WithValue(ctx.Context(), key, val)
}

var upgrader = websocket.FastHTTPUpgrader{}

func (ctx *Context) Upgrade() {

	err := upgrader.Upgrade(ctx.fastHttpRequestCtx, func(serverConn *websocket.Conn) {
		defer serverConn.Close()
		clientConn, ok := <-ctx.connChan
		if !ok || clientConn == nil {
			return
		}
		defer clientConn.Close()

		wg := &sync.WaitGroup{}
		go func() {
			wg.Add(1)
			for {
				msgType, msg, err := serverConn.ReadMessage()
				if err != nil {
					log.Error("read:", err)
					break
				}
				err = clientConn.WriteMessage(msgType, msg)
				if err != nil {
					log.Error("write message error: ", err)
					break
				}
			}
			wg.Done()
		}()
		go func() {
			wg.Add(1)
			for {
				msgType, msg, err := clientConn.ReadMessage()
				if err != nil {
					log.Error("read upstream message err: ", err)
					return
				}
				err = serverConn.WriteMessage(msgType, msg)
				if err != nil {
					log.Error("write client message err: ", err)
					return
				}
			}
			wg.Done()
		}()

		wg.Wait()
	})
	if err != nil {
		log.Error("upgrade error: ", err)
		return
	}
}

//NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx, port int) *Context {
	id := uuid.NewV4()
	requestID := id.String()

	newCtx := &Context{
		fastHttpRequestCtx: ctx,
		header:             readHeader(ctx.Request),
		requestID:          requestID,
		port:               port,
		connChan:           make(chan *websocket.Conn),
	}

	//记录请求时间
	newCtx.WithValue("request_time", ctx.Time())
	go newCtx.Upgrade()
	return newCtx
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

func (ctx *Context) FastFinish() {
	close(ctx.connChan)
}

func readHeader(request fasthttp.Request) http.Header {
	header := make(http.Header)
	hs := strings.Split(request.Header.String(), "\r\n")
	for _, t := range hs {
		vs := strings.SplitN(t, ":", 2)
		if len(vs) < 2 {
			if vs[0] == "" {
				continue
			}
			header.Set(vs[0], "")
			continue
		}
		header.Set(vs[0], strings.TrimSpace(vs[1]))
	}
	return header
}

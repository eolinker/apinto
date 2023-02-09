package dubbo_context

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"fmt"
	"github.com/eolinker/eosc/eocontext"
	eoscContext "github.com/eolinker/eosc/eocontext"
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"github.com/eolinker/eosc/utils/config"
	"github.com/google/uuid"
	"net"
	"time"
)

var _ dubbo_context.IDubboContext = (*DubboContext)(nil)

type DubboContext struct {
	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	requestReader       dubbo_context.IRequestReader
	proxy               dubbo_context.IProxy
	labels              map[string]string
	port                int
	requestID           string
	conn                net.Conn
	acceptTime          time.Time
}

func NewContext(dubboPackage *impl.DubboPackage, port int, conn net.Conn) dubbo_context.IDubboContext {

	headerReader := &RequestHeaderReader{
		id:             dubboPackage.Header.ID,
		serialID:       dubboPackage.Header.SerialID,
		_type:          int(dubboPackage.Header.Type),
		bodyLen:        dubboPackage.GetBodyLen(),
		responseStatus: dubboPackage.Header.ResponseStatus,
	}

	headerWrite := &RequestHeaderWrite{
		id:             headerReader.id,
		serialID:       headerReader.serialID,
		_type:          headerReader._type,
		bodyLen:        headerReader.bodyLen,
		responseStatus: headerReader.responseStatus,
	}

	serviceReader := &RequestServiceReader{
		path:        dubboPackage.Service.Path,
		serviceName: dubboPackage.Service.Interface,
		group:       dubboPackage.Service.Group,
		version:     dubboPackage.Service.Version,
		method:      dubboPackage.Service.Method,
		timeout:     dubboPackage.Service.Timeout,
	}
	serviceWriter := &RequestServiceWrite{
		path:        serviceReader.path,
		serviceName: serviceReader.serviceName,
		group:       serviceReader.group,
		version:     serviceReader.version,
		method:      serviceReader.method,
		timeout:     serviceReader.timeout,
	}

	proxy := &Proxy{
		HeaderWriter:  headerWrite,
		serviceWriter: serviceWriter,
		body:          dubboPackage,
	}

	requestReader := &RequestReader{
		headerReader:  headerReader,
		serviceReader: serviceReader,
		body:          dubboPackage,
		attachments:   nil,
	}

	t := time.Now()
	dubboContext := &DubboContext{
		labels:        make(map[string]string),
		port:          port,
		requestID:     uuid.New().String(),
		proxy:         proxy,
		requestReader: requestReader,
		conn:          conn,
		acceptTime:    t,
	}
	dubboContext.ctx = context.Background()
	dubboContext.WithValue("request_time", t)

	return dubboContext
}

func (d *DubboContext) HeaderReader() dubbo_context.IRequestReader {
	return d.requestReader
}

func (d *DubboContext) Proxy() dubbo_context.IProxy {
	return d.proxy
}

func (d *DubboContext) SendTo(address string, timeout time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (d *DubboContext) RequestId() string {
	return d.requestID
}

func (d *DubboContext) AcceptTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *DubboContext) Context() context.Context {
	if d.ctx == nil {
		d.ctx = context.Background()
	}
	return d.ctx
}

func (d *DubboContext) Value(key interface{}) interface{} {
	return d.Context().Value(key)
}

func (d *DubboContext) WithValue(key, val interface{}) {
	d.ctx = context.WithValue(d.Context(), key, val)
}

func (d *DubboContext) Scheme() string {
	return "dubbo"
}

func (d *DubboContext) Assert(i interface{}) error {
	if v, ok := i.(*dubbo_context.IDubboContext); ok {
		*v = d
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (d *DubboContext) SetLabel(name, value string) {
	d.labels[name] = value
}

func (d *DubboContext) GetLabel(name string) string {
	return d.labels[name]
}

func (d *DubboContext) Labels() map[string]string {
	return d.labels
}

func (d *DubboContext) GetComplete() eocontext.CompleteHandler {
	return d.completeHandler
}

func (d *DubboContext) SetCompleteHandler(handler eocontext.CompleteHandler) {
	d.completeHandler = handler
}

func (d *DubboContext) GetFinish() eocontext.FinishHandler {
	return d.finishHandler
}

func (d *DubboContext) SetFinish(handler eocontext.FinishHandler) {
	d.finishHandler = handler
}

func (d *DubboContext) GetApp() eocontext.EoApp {
	return d.app
}

func (d *DubboContext) SetApp(app eocontext.EoApp) {
	d.app = app
}

func (d *DubboContext) GetBalance() eocontext.BalanceHandler {
	return d.balance
}

func (d *DubboContext) SetBalance(handler eocontext.BalanceHandler) {
	d.balance = handler
}

func (d *DubboContext) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	return d.upstreamHostHandler
}

func (d *DubboContext) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	d.upstreamHostHandler = handler
}

func (d *DubboContext) LocalIP() net.IP {
	return addrToIP(d.conn.LocalAddr())
}

func (d *DubboContext) LocalAddr() net.Addr {
	return d.conn.LocalAddr()
}

func (d *DubboContext) LocalPort() int {
	return d.port
}

func addrToIP(addr net.Addr) net.IP {
	x, ok := addr.(*net.TCPAddr)
	if !ok {
		return net.IPv4zero
	}
	return x.IP
}

package dubbo2_context

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc/eocontext"
	eoscContext "github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
	"github.com/google/uuid"
	"net"
	"net/netip"
	"reflect"
	"strings"
	"time"
)

func NewDubboParamBody(typesList []string, valuesList []hessian.Object) *dubbo2_context.Dubbo2ParamBody {

	valList := make([]interface{}, 0, len(valuesList))
	for _, v := range valuesList {
		valList = append(valList, v)
	}
	return &dubbo2_context.Dubbo2ParamBody{TypesList: typesList, ValuesList: valList}
}

var _ dubbo2_context.IDubbo2Context = (*DubboContext)(nil)

type DubboContext struct {
	netIP               net.IP
	localAddr           net.Addr
	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	response            dubbo2_context.IResponse
	requestReader       dubbo2_context.IRequestReader
	proxy               dubbo2_context.IProxy
	labels              map[string]string
	port                int
	requestID           string
	acceptTime          time.Time
}

func (d *DubboContext) Response() dubbo2_context.IResponse {

	return d.response
}

func NewContext(req *invocation.RPCInvocation, port int) dubbo2_context.IDubbo2Context {

	t := time.Now()

	method, typesList, valuesList := argumentsUnmarshal(req.Arguments())
	if method == "" || len(typesList) == 0 || len(valuesList) == 0 {
		log.Errorf("dubbo2 NewContext method=%s typesList=%v valuesList=%v req=%v", method, typesList, valuesList, req)
	}

	path := req.GetAttachmentWithDefaultValue(constant.PathKey, "")
	serviceName := req.GetAttachmentWithDefaultValue(constant.InterfaceKey, "")
	group := req.GetAttachmentWithDefaultValue(constant.GroupKey, "")
	version := req.GetAttachmentWithDefaultValue(constant.VersionKey, "")

	serviceReader := NewRequestServiceReader(path, serviceName, group, version, method)

	serviceWriter := NewRequestServiceWrite(path, serviceName, group, version, method)

	paramBody := NewDubboParamBody(typesList, valuesList)
	proxy := NewProxy(serviceWriter, paramBody, req.Attachments())

	copyMaps := utils.CopyMaps(req.Attachments())

	localAddr, _ := req.GetAttachment(constant.LocalAddr)

	remoteAddr, _ := req.GetAttachment(constant.RemoteAddr)
	remoteIp := remoteAddr[:strings.Index(remoteAddr, ":")]

	requestReader := NewRequestReader(serviceReader, localAddr, remoteIp, copyMaps)

	addr, _ := netip.ParseAddrPort(localAddr)

	addrPort := net.TCPAddrFromAddrPort(addr)

	dubboContext := &DubboContext{
		labels:        make(map[string]string),
		port:          port,
		requestID:     uuid.New().String(),
		proxy:         proxy,
		requestReader: requestReader,
		netIP:         addrPort.IP,
		localAddr:     addrPort,
		acceptTime:    t,
		response:      &Response{},
	}
	dubboContext.ctx = context.Background()
	dubboContext.WithValue("request_time", t)

	return dubboContext
}

func (d *DubboContext) HeaderReader() dubbo2_context.IRequestReader {
	return d.requestReader
}

func (d *DubboContext) Proxy() dubbo2_context.IProxy {
	return d.proxy
}

func (d *DubboContext) Invoke(address string, timeout time.Duration) error {
	return d.dial(address, timeout)
}

func (d *DubboContext) dial(addr string, timeout time.Duration) error {
	arguments := make([]interface{}, 3)
	parameterValues := make([]reflect.Value, 3)

	valuesList := make([]hessian.Object, 0, len(d.proxy.GetParam().ValuesList))

	for _, v := range d.proxy.GetParam().ValuesList {
		if object, ok := v.(hessian.Object); ok {
			valuesList = append(valuesList, object)
		}
	}

	arguments[0] = d.proxy.Service().Method()
	arguments[1] = d.proxy.GetParam().TypesList
	arguments[2] = valuesList

	parameterValues[0] = reflect.ValueOf(arguments[0])
	parameterValues[1] = reflect.ValueOf(arguments[1])
	parameterValues[2] = reflect.ValueOf(arguments[2])

	invoc := invocation.NewRPCInvocationWithOptions(invocation.WithMethodName("$invoke"),
		invocation.WithArguments(arguments),
		invocation.WithParameterValues(parameterValues))

	serviceName := d.proxy.Service().Interface()
	url, err := common.NewURL(addr,
		common.WithProtocol(dubbo.DUBBO), common.WithParamsValue(constant.SerializationKey, constant.Hessian2Serialization),
		common.WithParamsValue(constant.GenericFilterKey, "true"),
		common.WithParamsValue(constant.TimeoutKey, timeout.String()),
		common.WithParamsValue(constant.InterfaceKey, serviceName),
		common.WithParamsValue(constant.ReferenceFilterKey, "generic,filter"),
		common.WithPath(serviceName),
	)
	if err != nil {
		return err
	}

	for k, v := range d.proxy.Attachments() {
		invoc.SetAttachment(k, v)
	}

	//源码中已对连接做了缓存池
	dubboProtocol := dubbo.NewDubboProtocol()
	invoker := dubboProtocol.Refer(url)
	var resp interface{}
	invoc.SetReply(&resp)

	result := invoker.Invoke(d.Context(), invoc)
	if result.Error() != nil {
		return result.Error()
	}

	rpcResult := result.(*protocol.RPCResult)

	val := result.Result().(*interface{})

	d.response.SetBody(getResponse(formatData(*val), rpcResult.Err, rpcResult.Attachments()))

	return err
}

func getResponse(obj interface{}, err error, attachments map[string]interface{}) protocol.RPCResult {
	payload := impl.NewResponsePayload(obj, err, attachments)
	return protocol.RPCResult{
		Attrs: payload.Attachments,
		Err:   payload.Exception,
		Rest:  payload.RspObj,
	}
}

func (d *DubboContext) RequestId() string {
	return d.requestID
}

func (d *DubboContext) AcceptTime() time.Time {
	return d.acceptTime
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
	if v, ok := i.(*dubbo2_context.IDubbo2Context); ok {
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
	return d.netIP
}

func (d *DubboContext) LocalAddr() net.Addr {
	return d.localAddr
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

func argumentsUnmarshal(arguments []interface{}) (string, []string, []hessian.Object) {
	methodName := ""
	typeList := make([]string, 0)
	valueList := make([]hessian.Object, 0)

	if len(arguments) > 0 {
		if argsStr, sOk := arguments[0].(string); sOk {
			methodName = argsStr
		}
	}
	if len(arguments) > 1 {
		if argsTypeList, sOk := arguments[1].([]string); sOk {
			typeList = argsTypeList
		}
	}
	if len(arguments) > 2 {
		if argsValueList, sOk := arguments[2].([]hessian.Object); sOk {
			valueList = argsValueList
		}
	}

	return methodName, typeList, valueList

}

func formatData(value interface{}) interface{} {

	switch valueTemp := value.(type) {
	case map[interface{}]interface{}:
		maps := make(map[string]interface{})
		for k, v := range valueTemp {
			maps[utils.InterfaceToString(k)] = formatData(v)
		}
		return maps
	case []interface{}:
		values := make([]interface{}, 0)

		for _, v := range valueTemp {
			values = append(values, formatData(v))
		}
		return values
	default:
		return value
	}
}

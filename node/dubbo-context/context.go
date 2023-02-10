package dubbo_context

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc/eocontext"
	eoscContext "github.com/eolinker/eosc/eocontext"
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"
	"github.com/google/uuid"
	"net"
	"reflect"
	"time"
)

type DubboParamBody struct {
	typesList  []string
	valuesList []hessian.Object
}

var _ dubbo_context.IDubboContext = (*DubboContext)(nil)

type DubboContext struct {
	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	response            dubbo_context.IResponse
	requestReader       dubbo_context.IRequestReader
	proxy               dubbo_context.IProxy
	labels              map[string]string
	port                int
	requestID           string
	conn                net.Conn
	acceptTime          time.Time
}

func (d *DubboContext) Response() dubbo_context.IResponse {
	return d.response
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

	attachments, method, typesList, valuesList := packageUnmarshal(dubboPackage)

	serviceReader := &RequestServiceReader{
		path:        dubboPackage.Service.Path,
		serviceName: dubboPackage.Service.Interface,
		group:       dubboPackage.Service.Group,
		version:     dubboPackage.Service.Version,
		method:      method,
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
		param: &DubboParamBody{
			typesList:  typesList,
			valuesList: valuesList,
		},
		attachments: attachments,
	}

	copyMaps := utils.CopyMaps(attachments)

	requestReader := &RequestReader{
		headerReader:  headerReader,
		serviceReader: serviceReader,
		body:          dubboPackage,
		host:          conn.LocalAddr().String(),
		attachments:   copyMaps,
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
		response:      &Response{},
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

func (d *DubboContext) Invoke(address string, timeout time.Duration) error {

	t := time.Now()
	defer func() {
		d.response.SetResponseTime(time.Now().Sub(t))
	}()

	return d.dial(address, timeout)
}

func (d *DubboContext) dial(addr string, timeout time.Duration) error {
	if d.conn != nil {
		defer d.conn.Close()
	}
	arguments := make([]interface{}, 3)
	parameterValues := make([]reflect.Value, 3)

	typesList := make([]string, 0)
	valuesList := make([]hessian.Object, 0)
	if param, ok := d.proxy.GetParam().(*DubboParamBody); ok {
		typesList = param.typesList
		valuesList = param.valuesList
	}

	arguments[0] = d.proxy.Service().Method()
	arguments[1] = typesList
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

	dubboProtocol := dubbo.NewDubboProtocol()
	invoker := dubboProtocol.Refer(url)
	var resp interface{}
	invoc.SetReply(&resp)

	result := invoker.Invoke(context.Background(), invoc)
	if result.Error() != nil {
		return result.Error()
	}

	v := result.Result().(*interface{})

	bytes := formatData(*v)

	d.response.SetBody(bytes)

	by, err := d.packageMarshal(bytes)
	if err != nil {
		log.Errorf("dubbo-dial.packageMarshal err=%v", err)
		return err
	}

	_, err = d.conn.Write(by)

	return err
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

func (d *DubboContext) packageMarshal(body interface{}) ([]byte, error) {
	dubboPackage := impl.NewDubboPackage(nil)
	dubboPackage.Header = impl.DubboHeader{
		SerialID:       d.proxy.Header().SerialID(),
		Type:           impl.PackageResponse,
		ID:             d.proxy.Header().ID(),
		ResponseStatus: impl.Response_OK,
	}
	dubboPackage.SetBody(impl.EnsureResponsePayload(body))
	buf, err := dubboPackage.Marshal()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func packageUnmarshal(dubboPackage *impl.DubboPackage) (map[string]interface{}, string, []string, []hessian.Object) {
	attachments := make(map[string]interface{})
	methodName := ""
	typeList := make([]string, 0)
	valueList := make([]hessian.Object, 0)
	if bodyMap, bOk := dubboPackage.Body.(map[string]interface{}); bOk {
		if attachmentsInteface, aOk := bodyMap["attachments"]; aOk {
			if attachmentsTemp, ok := attachmentsInteface.(map[string]interface{}); ok {
				attachments = attachmentsTemp
			}

		}

		if argsMap, aOk := bodyMap["args"]; aOk {
			if argsList, lOk := argsMap.([]interface{}); lOk {

				if len(argsList) > 0 {
					if argsStr, sOk := argsList[0].(string); sOk {
						methodName = argsStr
					}
				}
				if len(argsList) > 1 {
					if argsTypeList, sOk := argsList[1].([]string); sOk {
						typeList = argsTypeList
					}
				}
				if len(argsList) > 2 {
					if argsValueList, sOk := argsList[2].([]hessian.Object); sOk {
						valueList = argsValueList
					}
				}

			}
		}
	}

	return attachments, methodName, typeList, valueList
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

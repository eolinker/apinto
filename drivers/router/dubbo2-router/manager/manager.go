package manager

import (
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"dubbo.apache.org/dubbo-go/v3/remoting"
	dubbo2_context "github.com/eolinker/apinto/node/dubbo2-context"
	"github.com/eolinker/apinto/router"
	eoscContext "github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"sync"
	"sync/atomic"
)

var _ IManger = (*dubboManger)(nil)

type IManger interface {
	Set(id string, port int, serviceName, methodName string, rule []AppendRule, handler router.IRouterHandler) error
	Delete(id string)
	Handler(port int)
}

func (d *dubboManger) SetGlobalFilters(globalFilters *eoscContext.IChainPro) {
	d.globalFilters.Store(globalFilters)
}

func NewManager() *dubboManger {
	return &dubboManger{
		matcher:       nil,
		routersData:   new(RouterData),
		connHandler:   NewConnHandler(),
		globalFilters: atomic.Pointer[eoscContext.IChainPro]{},
	}
}

type dubboManger struct {
	lock          sync.RWMutex
	matcher       router.IMatcher
	routersData   IRouterData
	connHandler   *ConnHandler
	globalFilters atomic.Pointer[eoscContext.IChainPro]
}

func (d *dubboManger) Set(id string, port int, serviceName, methodName string, rule []AppendRule, handler router.IRouterHandler) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	routersData := d.routersData.Set(id, port, serviceName, methodName, rule, handler)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Error("parse router data error: ", err)
		return err
	}
	d.matcher = matchers
	d.routersData = routersData
	return nil
}

func (d *dubboManger) Delete(id string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	routersData := d.routersData.Delete(id)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Errorf("delete router:%s %s", id, err.Error())
		return
	}

	d.matcher = matchers
	d.routersData = routersData
	return
}

func (d *dubboManger) Handler(port int) {
	for result := range d.connHandler.GetResult() {

		if result.Result().IsRequest {
			go func() {
				req := result.Result().Result.(*remoting.Request)

				dubboPackage := impl.NewDubboPackage(nil)

				dubboPackage.Header = impl.DubboHeader{
					SerialID: req.SerialID,
					Type:     impl.PackageRequest,
					ID:       req.ID,
				}

				if invoc, ok := req.Data.(*protocol.Invocation); ok {
					invocation := *invoc

					dubboPackage.Service.Path = invocation.GetAttachmentWithDefaultValue(constant.PathKey, "")
					dubboPackage.Service.Interface = invocation.GetAttachmentWithDefaultValue(constant.InterfaceKey, "")
					dubboPackage.Service.Version = invocation.GetAttachmentWithDefaultValue(constant.VersionKey, "")
					dubboPackage.Service.Group = invocation.GetAttachmentWithDefaultValue(constant.GroupKey, "")
					dubboPackage.Service.Method = invocation.MethodName()
				}

				context := dubbo2_context.NewContext(dubboPackage, port, result.Conn())

				match, has := d.matcher.Match(port, context.HeaderReader())
				if !has {
					//todo 怎样处理 conn.Write() ???
				} else {
					log.Debug("match has:", port)
					match.ServeHTTP(context)
				}
			}()

		}

	}

}

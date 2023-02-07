package manager

import (
	"bytes"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	dubbo_router "github.com/eolinker/apinto/router/dubbo-router"
	"github.com/eolinker/eosc/log"
	"net"
	"sync"
)

var _ IDubboManger = (*dubboManger)(nil)

type IDubboManger interface {
	//Set driver:http/dubbo/grpc
	Set(id string, port int, driver, host, path, httpMethod, serviceName, methodName string) error
	Delete(id string)
	// FastHandler func(port int, ln net.Listener)方法中调用
	FastHandler(port int, conn net.Conn)
}

func NewManager() IDubboManger {
	return &dubboManger{}
}

type dubboManger struct {
	lock    sync.RWMutex
	matcher dubbo_router.IMatcher
}

func (d *dubboManger) Set(id string, port int, driver, host, path, httpMethod, serviceName, methodName string) error {
	//TODO implement me
	panic("implement me")
}

func (d *dubboManger) Delete(id string) {
	//TODO implement me
	panic("implement me")
}

func (d *dubboManger) FastHandler(port int, conn net.Conn) {
	defer conn.Close()
	var info [128 * 1024]byte
	n, err := conn.Read(info[:])
	buf := bytes.NewBuffer(info[:n])
	dubboPackage := impl.NewDubboPackage(buf)
	if err = dubboPackage.ReadHeader(); err != nil {
		log.Errorf("dubboManger FastHandler err=%v", err)
		return
	}

	if err = dubboPackage.Unmarshal(); err != nil {
		log.Errorf("dubboManger FastHandler err=%v", err)
		return
	}

	match, has := d.matcher.Match(port, dubboPackage.Service)
	if !has {
		//todo 怎样处理 conn.Write() ???

	} else {
		log.Debug("match has:", port)
		match.DubboProxy(dubboPackage)
	}
}

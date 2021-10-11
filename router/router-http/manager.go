package router_http

import (
	"errors"
	"sync"

	"github.com/eolinker/eosc/env"

	"github.com/eolinker/eosc/log"

	traffic_http_fast "github.com/eolinker/eosc/traffic/traffic-http-fast"

	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/traffic"
)

var _ iManager = (*Manager)(nil)

var (
	errorCertificateNotExit = errors.New("not exist cert")
)

type iManager interface {
	Add(port int, id string, config *Config) error
	Del(port int, id string) error
	Cancel()
}

var manager iManager

func init() {
	var tf traffic.ITraffic
	bean.Autowired(&tf)

	bean.AddInitializingBeanFunc(func() {
		log.Debug("init router manager")

		manager = NewManager(tf)
	})
}

//Manager 路由管理器结构体
type Manager struct {
	locker  sync.Mutex
	routers IRouters

	tf traffic_http_fast.IHttpTraffic
}

//Cancel 关闭路由管理器
func (m *Manager) Cancel() {
	m.locker.Lock()
	defer m.locker.Unlock()

	m.tf.Close()
	m.tf = nil

}

//NewManager 创建路由管理器
func NewManager(tf traffic.ITraffic) *Manager {

	m := &Manager{
		routers: NewRouters(),
		tf:      traffic_http_fast.NewHttpTraffic(tf),
		locker:  sync.Mutex{},
	}
	return m
}

//Add 新增路由配置到路由管理器中
func (m *Manager) Add(port int, id string, config *Config) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	router, _, err := m.routers.Set(port, id, config)
	if err != nil {
		return err
	}
	if config.Protocol == "https" {
		certs := newCerts(config.Cert)
		m.tf.Get(port).SetHttps(router.Handler(), certs.certs)

	} else {
		m.tf.Get(port).SetHttp(router.Handler())
	}

	return nil
}

//Del 将某个路由配置从路由管理器中删去
func (m *Manager) Del(port int, id string) error {
	m.locker.Lock()
	defer m.locker.Unlock()
	if r, has := m.routers.Del(port, id); has {
		//若目标端口的http服务器已无路由配置，则关闭服务器及listener
		count := r.Count()

		log.Debug("after delete router,count of port:", port, " count:", count)
		if count == 0 {
			m.tf.ShutDown(port)
		} else if env.IsDebug() {

		}
	}

	return nil

}

//Add 将路由配置加入到路由管理器
func Add(port int, id string, config *Config) error {
	return manager.Add(port, id, config)
}

//Del 将路由配置从路由管理器中删去
func Del(port int, id string) error {
	return manager.Del(port, id)
}

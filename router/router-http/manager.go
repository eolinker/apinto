package router_http

import (
	"crypto/tls"
	"errors"
	"github.com/eolinker/apinto/plugin"
	"sync"

	"github.com/eolinker/eosc/config"

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
	var cfg *config.ListensMsg
	var pluginManager plugin.IPluginManager

	bean.Autowired(&tf)
	bean.Autowired(&cfg)
	bean.Autowired(&pluginManager)
	bean.AddInitializingBeanFunc(func() {
		log.Debug("init router manager")

		manager = NewManager(tf, cfg, pluginManager)
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
func NewManager(tf traffic.ITraffic, listenCfg *config.ListensMsg, pluginManager plugin.IPluginManager) *Manager {
	log.Debug("new router manager")
	m := &Manager{
		routers: NewRouters(pluginManager),
		tf:      traffic_http_fast.NewHttpTraffic(),
		locker:  sync.Mutex{},
	}
	if tf.IsStop() {
		return m
	}

	for _, cfg := range listenCfg.Listens {
		port := int(cfg.Port)

		l := tf.ListenTcp(port, traffic.Http1)
		if l == nil {
			continue
		}
		log.Debug("new http service ", port, cfg, l)
		if cfg.Scheme == "https" {
			cert, err := config.NewCert(cfg.Certificate, listenCfg.Dir)
			if err != nil {
				log.Warn("worker create certificate error:", err)
				continue
			}
			m.tf.Set(port, traffic_http_fast.NewHttpService(tls.NewListener(l, &tls.Config{GetCertificate: cert.GetCertificate})))
			continue
		}
		m.tf.Set(port, traffic_http_fast.NewHttpService(l))
	}
	return m
}

//Add 新增路由配置到路由管理器中
func (m *Manager) Add(port int, id string, config *Config) error {
	m.locker.Lock()
	defer m.locker.Unlock()
	if port == 0 {
		srv := m.tf.All()
		for p, s := range srv {
			router, _, err := m.routers.Set(p, id, config)
			if err != nil {
				return err
			}
			s.Set(router.Handler)
		}
		return nil
	}
	router, _, err := m.routers.Set(port, id, config)
	if err != nil {
		return err
	}
	serviceTF, has := m.tf.Get(port)
	if !has {
		log.Debug("not has port")
		return nil
	}
	serviceTF.Set(router.Handler)

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

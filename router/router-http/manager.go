package router

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"github.com/eolinker/eosc/listener"
	"net"
	"net/http"
	"sync"
	"time"
)

var _ iManager = (*Manager)(nil)
var (
	sign = ""
)

func init() {
	n := time.Now().UnixNano()
	data := make([]byte, 8)
	binary.PutVarint(data, n)
	sign = hex.EncodeToString(data)
}

type iManager interface {
	Add(port int, id string, config *Config) error
	Del(port int, id string) error
	Cancel()
}

var manager = NewManager()

type Manager struct {
	locker    sync.Mutex
	routers   IRouters
	servers   map[int]*http.Server
	listeners map[int]net.Listener


}

func (m *Manager) Cancel() {
	m.locker.Lock()
	defer m.locker.Unlock()
	ctx:=context.Background()
 	for p,s:=range m.servers{
 		s.Shutdown(ctx)
		delete(m.servers, p)
	}

	for k,l:=range m.listeners{
		l.Close()
		delete(m.listeners, k)
	}
}

func NewManager() *Manager {
	return &Manager{
		routers: NewRouters(),
	}
}

func (m *Manager) Add(port int, id string, config *Config) error {
	m.locker.Lock()
	defer m.locker.Unlock()
	router, isCreate, err := m.routers.Set(port, id, config)
	if err != nil {
		return err
	}
	if isCreate {
		s, has := m.servers[port]
		if !has {
			s = &http.Server{}

			s.Handler = router
			l, err := listener.ListenTCP(port, sign)
			if err != nil {
				return err
			}
			go s.Serve(l)

			m.servers[port] = s
			m.listeners[port] = l
		}
	}
	return nil
}

func (m *Manager) Del(port int, id string) error {
	m.locker.Lock()
	defer m.locker.Unlock()
	if r, has := m.routers.Del(port, id); has {
		if r.Count() == 0 {
			if s, has := m.servers[port]; has {
				ctx := context.Background()
				err := s.Shutdown(ctx)
				if err != nil {
					return err
				}
				delete(m.servers, port)
				m.listeners[port].Close()
				delete(m.listeners, port)
			}
		}
	}

	return nil

}

func Add(port int, id string, config *Config) error {
	return manager.Add(port, id, config)
}

func Del(port int, id string) error {
	return manager.Del(port, id)
}

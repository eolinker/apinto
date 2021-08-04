package router_http

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/eosc/listener"
)

var _ iManager = (*Manager)(nil)
var (
	sign                     = ""
	_ErrorCertificateNotExit = errors.New("not exit ca")
)

func init() {
	n := time.Now().UnixNano()
	data := make([]byte, 9)
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
	servers   map[int]*httpServer
	listeners map[int]net.Listener
}

type httpServer struct {
	tlsConfig *tls.Config
	port      int
	protocol  string
	srv       *http.Server
	certs     *Certs
}

func (s *httpServer) shutdown(ctx context.Context) {
	s.srv.Shutdown(ctx)
}

func (a *httpServer) GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if a.certs == nil {
		return nil, _ErrorCertificateNotExit
	}
	certificate, has := a.certs.Get(strings.ToLower(info.ServerName))
	if !has {
		return nil, _ErrorCertificateNotExit
	}

	return certificate, nil
}

func (m *Manager) Cancel() {
	m.locker.Lock()
	defer m.locker.Unlock()
	ctx := context.Background()
	for p, s := range m.servers {
		s.shutdown(ctx)
		delete(m.servers, p)
	}

	for k, l := range m.listeners {
		l.Close()
		delete(m.listeners, k)
	}
}

func NewManager() *Manager {
	return &Manager{
		routers:   NewRouters(),
		servers:   make(map[int]*httpServer),
		listeners: make(map[int]net.Listener),
		locker:    sync.Mutex{},
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
			s = &httpServer{srv: &http.Server{}}

			s.srv.Handler = router
			l, err := listener.ListenTCP(port, sign)
			if err != nil {
				return err
			}
			if config.Protocol == "https" {
				s.certs = newCerts(config.Cert)
				s.tlsConfig = &tls.Config{GetCertificate: s.GetCertificate}
				l = tls.NewListener(l, s.tlsConfig)
			}
			go s.srv.Serve(l)

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
				err := s.srv.Shutdown(ctx)
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

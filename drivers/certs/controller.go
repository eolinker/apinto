package certs

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/eolinker/eosc"
	"reflect"
	"strings"
	"sync"
)

var (
	controller                            = NewController()
	_                       eosc.ISetting = controller
	_                       IController   = controller
	configType                            = reflect.TypeOf((*Config)(nil))
	errorCertificateNotExit               = errors.New("not exist cert")
)

type IController interface {
	Store(id string)
	Del(id string)
	Save(name string, cert *tls.Certificate)
}
type Controller struct {
	profession string
	driver     string
	all        map[string]struct{}
	certs      map[string]*tls.Certificate
	lock       *sync.Mutex
}

func (c *Controller) Store(id string) {
	c.all[id] = struct{}{}

}

func (c *Controller) Del(id string) {
	delete(c.all, id)
}

func (c *Controller) Save(name string, cert *tls.Certificate) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.certs[name] = cert

	for _, dnsName := range cert.Leaf.DNSNames {
		c.certs[dnsName] = cert
	}

}

func (c *Controller) ConfigType() reflect.Type {
	return configType
}

func (c *Controller) Set(conf interface{}) (err error) {
	return eosc.ErrorUnsupportedKind
}

func (c *Controller) Get() interface{} {
	return nil
}

func (c *Controller) Mode() eosc.SettingMode {
	return eosc.SettingModeBatch
}

func (c *Controller) Check(cfg interface{}) (profession, name, driver, desc string, err error) {
	conf, ok := cfg.(*Config)
	if !ok {
		err = eosc.ErrorConfigType
		return
	}

	if empty(conf.Name) {
		err = eosc.ErrorConfigFieldUnknown
		return
	}

	_, err = parseCert(conf.Key, conf.Pem)
	if err != nil {
		return "", "", "", "", err
	}

	return c.profession, conf.Name, c.driver, "", nil

}
func empty(vs ...string) bool {
	for _, v := range vs {
		if len(v) == 0 {
			return true
		}
	}
	return false
}
func (c *Controller) AllWorkers() []string {
	ws := make([]string, 0, len(c.all))
	for id := range c.all {
		ws = append(ws, id)
	}
	return ws
}

// Get 获取证书
func (c *Controller) getCert(hostName string) (*tls.Certificate, bool) {
	if c == nil || len(c.certs) == 0 {
		return nil, true
	}
	cert, has := c.certs[hostName]
	if has {
		return cert, true
	}
	hs := strings.Split(hostName, ".")
	if len(hs) < 1 {
		return nil, false
	}

	cert, has = c.certs[fmt.Sprintf("*.%s", strings.Join(hs[1:], "."))]
	return cert, has
}

func NewController() *Controller {
	return &Controller{
		all:   map[string]struct{}{},
		certs: map[string]*tls.Certificate{},
		lock:  &sync.Mutex{},
	}
}

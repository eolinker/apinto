package certs

import (
	"crypto/tls"
	"errors"
	"github.com/eolinker/apinto/certs"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"reflect"
)

var (
	controller                            = NewController()
	_                       eosc.ISetting = controller
	_                       IController   = controller
	configType                            = reflect.TypeOf((*Config)(nil))
	errorCertificateNotExit               = errors.New("not exist cert")
)

func init() {
	bean.Injection(&controller)
}

type IController interface {
	Store(id string)
	Del(id string, cert *tls.Certificate)
	Save(cert *tls.Certificate)
}
type Controller struct {
	profession string
	driver     string
	all        map[string]struct{}
	iCerts     certs.ICert
}

func (c *Controller) Store(id string) {
	c.all[id] = struct{}{}
}

func (c *Controller) Del(id string, cert *tls.Certificate) {

	delete(c.all, id)

	c.iCerts.DelCert(cert)
}

func (c *Controller) Save(cert *tls.Certificate) {
	c.iCerts.SaveCert(cert)
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

func NewController() *Controller {

	c := &Controller{
		all: map[string]struct{}{},
	}

	bean.Autowired(&c.iCerts)
	return c
}

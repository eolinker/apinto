package certs

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/config"
	"strings"
	"sync"
)

func init() {
	iCert := NewCert()
	bean.Injection(&iCert)
}

var errorCertificateNotExit = errors.New("not exist cert")

type ICert interface {
	GetCertificateFunc(certs []*config.Certificate, dir string) func(info *tls.ClientHelloInfo) (*tls.Certificate, error)
	SaveCert(certificate *tls.Certificate)
	DelCert(cert *tls.Certificate)
}

type cert struct {
	certFunc func(info *tls.ClientHelloInfo) (*tls.Certificate, error)
	certs    map[string]*tls.Certificate
	lock     *sync.RWMutex
}

func (c *cert) DelCert(cert *tls.Certificate) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if cert != nil {
		delete(c.certs, cert.Leaf.Subject.CommonName)
		for _, name := range cert.Leaf.DNSNames {
			delete(c.certs, name)
		}
	}
}

func (c *cert) SaveCert(cert *tls.Certificate) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.certs[cert.Leaf.Subject.CommonName] = cert

	for _, dnsName := range cert.Leaf.DNSNames {
		c.certs[dnsName] = cert
	}
}

func (c *cert) GetCertificateFunc(certs []*config.Certificate, dir string) func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(info *tls.ClientHelloInfo) (certificate *tls.Certificate, err error) {

		newCert, _ := config.NewCert(certs, dir)

		if newCert != nil {
			certificate, err = newCert.GetCertificate(info)
			if certificate != nil {
				return
			}
		}

		if len(c.certs) == 0 {
			if err == nil {
				err = errorCertificateNotExit
			}
			return
		}

		var has = false
		certificate, has = c.getCert(strings.ToLower(info.ServerName))
		if !has {
			err = errorCertificateNotExit
			return
		}

		return
	}

}

// getCert 获取证书
func (c *cert) getCert(hostName string) (*tls.Certificate, bool) {
	if c == nil || len(c.certs) == 0 {
		return nil, true
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	certValue, has := c.certs[hostName]
	if has {
		return certValue, true
	}
	hs := strings.Split(hostName, ".")
	if len(hs) < 1 {
		return nil, false
	}

	certValue, has = c.certs[fmt.Sprintf("*.%s", strings.Join(hs[1:], "."))]
	return certValue, has
}

func NewCert() ICert {

	c := &cert{
		certs: map[string]*tls.Certificate{},
		lock:  &sync.RWMutex{},
	}

	return c
}

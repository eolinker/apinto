package certs

import (
	"crypto/tls"
	"crypto/x509"
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
	SaveCert(workerId string, cert *tls.Certificate, certificate *x509.Certificate)
	DelCert(workerId string)
}

type cert struct {
	certs      map[string]*tls.Certificate
	workerMaps map[string]map[string]*tls.Certificate
	lock       *sync.RWMutex
}

func (c *cert) DelCert(workerId string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delCerts := c.workerMaps[workerId]
	delete(c.workerMaps, workerId)

	//删除组装后的证书信息
	for key, _ := range delCerts {
		delete(c.certs, key)
	}

	//为防止把其他worker的证书也删除了，重新组装
	for _, certs := range c.workerMaps {
		for k, v := range certs {
			c.certs[k] = v
		}
	}
}

func (c *cert) SaveCert(workerId string, cert *tls.Certificate, certificate *x509.Certificate) {
	c.lock.Lock()
	defer c.lock.Unlock()

	//每次save都是覆盖操作
	certsMap := make(map[string]*tls.Certificate)

	certsMap[certificate.Subject.CommonName] = cert
	for _, dnsName := range certificate.DNSNames {
		certsMap[dnsName] = cert
	}

	c.workerMaps[workerId] = certsMap

	//将最新的证书组装好
	for k, v := range certsMap {
		c.certs[k] = v
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

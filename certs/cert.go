package certs

import (
	"crypto/tls"
	"errors"
	"github.com/eolinker/apinto/drivers/certs"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/config"
	"strings"
)

var certController = &certs.Controller{}

func init() {
	bean.Injection(&certController)
}

var errorCertificateNotExit = errors.New("not exist cert")

type ICert interface {
	GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error)
}

type cert struct {
	cert *config.Cert
}

func (c *cert) GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if c.cert != nil {
		certificate, _ := c.cert.GetCertificate(info)
		if certificate != nil {
			return certificate, nil
		}
	}

	if certController.Certs() == nil {
		return nil, errorCertificateNotExit
	}
	certificate, has := certController.GetCert(strings.ToLower(info.ServerName))
	if !has {
		return nil, errorCertificateNotExit
	}

	return certificate, nil

}

func NewCert(certs []*config.Certificate, dir string) ICert {
	newCert, _ := config.NewCert(certs, dir)

	c := &cert{
		cert: newCert,
	}

	return c
}

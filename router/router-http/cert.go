package router_http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/eolinker/eosc/log"
)

//Certs 证书集合结构体
type Certs struct {
	certs map[string]*tls.Certificate
}

//Get 获取证书
func (c *Certs) Get(hostName string) (*tls.Certificate, bool) {
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

func newCerts(certs []Cert) *Certs {
	cs := make(map[string]*tls.Certificate)
	for _, cert := range certs {
		x509KeyPair, err := tls.X509KeyPair([]byte(cert.Crt), []byte(cert.Key))
		if err != nil {
			log.Warn("parse ca error:", err)
			continue
		}
		certificate, err := x509.ParseCertificate(x509KeyPair.Certificate[0])
		if err != nil {
			log.Warn("parse cert error:", err)
			continue
		}
		cs[certificate.Subject.CommonName] = &x509KeyPair
		for _, dnsName := range certificate.DNSNames {
			cs[dnsName] = &x509KeyPair
		}
	}
	return &Certs{certs: cs}
}

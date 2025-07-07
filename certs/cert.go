package certs

import (
	"crypto/tls"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/eolinker/eosc/config"
	"github.com/tjfoc/gmsm/gmtls"
)

var errorCertificateNotExit = errors.New("not exist cert")

var (
	currentWorkers = make(map[string]*tls.Certificate)
	gmWorkers      = make(map[string][]*gmtls.Certificate)

	lock = sync.RWMutex{}
	// currentCert 普通TLS证书
	currentCert = atomic.Pointer[config.Cert[tls.Certificate]]{}
	// gmCert gmTLS证书
	gmCert = atomic.Pointer[config.Cert[gmtls.Certificate]]{}
	// gmEncCert gmTLS加密证书
	gmEncCert = atomic.Pointer[config.Cert[gmtls.Certificate]]{}
)

func init() {
	currentCert.Store(config.NewCert[tls.Certificate](nil))
	gmCert.Store(config.NewCert[gmtls.Certificate](nil))
	gmEncCert.Store(config.NewCert[gmtls.Certificate](nil))
}
func DelCert(workerId string) {
	lock.Lock()
	defer lock.Unlock()
	delete(currentWorkers, workerId)
	rebuild()
}

func SaveCert(workerId string, certs *tls.Certificate) {
	lock.Lock()
	defer lock.Unlock()
	currentWorkers[workerId] = certs
	rebuild()
}

func SaveGMCert(workerId string, certs []*gmtls.Certificate) {
	lock.Lock()
	defer lock.Unlock()
	gmWorkers[workerId] = certs
	gmRebuild()
}

func DelGMCert(workerId string) {
	lock.Lock()
	defer lock.Unlock()
	delete(gmWorkers, workerId)
	gmRebuild()
}

func gmRebuild() {
	gmMap := make(map[string]*gmtls.Certificate)
	gmEncMap := make(map[string]*gmtls.Certificate)
	for _, cs := range gmWorkers {
		l := len(cs)
		switch {
		case l == 2:
			i := cs[0]
			gmMap[i.Leaf.Subject.CommonName] = i
			for _, dnsName := range i.Leaf.DNSNames {
				gmMap[dnsName] = i
			}

			i = cs[1]
			gmEncMap[i.Leaf.Subject.CommonName] = i
			for _, dnsName := range i.Leaf.DNSNames {
				gmEncMap[dnsName] = i
			}
		default:
			continue
		}

	}
	gmCert.Swap(config.NewCert(gmMap))
	gmEncCert.Swap(config.NewCert(gmEncMap))
}
func rebuild() {
	currentMap := make(map[string]*tls.Certificate)
	for _, cs := range currentWorkers {
		i := cs
		currentMap[i.Leaf.Subject.CommonName] = i
		for _, dnsName := range i.Leaf.DNSNames {
			currentMap[dnsName] = i
		}
	}
	currentCert.Swap(config.NewCert(currentMap))
}

func GetCertificateFunc(certsLocal ...*config.Cert[tls.Certificate]) func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {

	if len(certsLocal) == 0 {

		return func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return currentCert.Load().GetCertificate(info)
		}
	}
	certList := make([]*config.Cert[tls.Certificate], 0, len(certsLocal))
	for _, c := range certsLocal {
		if c != nil {
			certList = append(certList, c)
		}
	}
	return func(info *tls.ClientHelloInfo) (certificate *tls.Certificate, err error) {
		certificate, err = currentCert.Load().GetCertificate(info)
		if certificate != nil {
			return
		}
		for _, cert := range certList {
			certificate, err = cert.GetCertificate(info)
			if certificate != nil {
				return
			}
		}
		if err == nil {
			err = errorCertificateNotExit
		}
		return
	}

}

func GetGMCertificateFunc() func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
	return func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
		return gmCert.Load().GetCertificate(info)
	}
}

func GetKECertificate() func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
	return func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
		return gmEncCert.Load().GetCertificate(info)
	}
}

package certs

import (
	"crypto/tls"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/eolinker/eosc/config"
)

var errorCertificateNotExit = errors.New("not exist cert")

type ICert interface {
	SaveCert(workerId string, cert *tls.Certificate)
	DelCert(workerId string)
}

var (
	workerMaps  = make(map[string]*tls.Certificate)
	lock        = sync.RWMutex{}
	currentCert = atomic.Pointer[config.Cert]{}
)

func init() {
	currentCert.Store(config.NewCert(nil))
}
func DelCert(workerId string) {
	lock.Lock()
	defer lock.Unlock()

	delete(workerMaps, workerId)
	rebuild()
}

func SaveCert(workerId string, cert *tls.Certificate) {
	lock.Lock()
	defer lock.Unlock()
	workerMaps[workerId] = cert
	rebuild()

}
func rebuild() {
	certsMap := make(map[string]*tls.Certificate)
	for _, i := range workerMaps {
		certsMap[i.Leaf.Subject.CommonName] = i
		for _, dnsName := range i.Leaf.DNSNames {
			certsMap[dnsName] = i
		}
	}
	currentCert.Swap(config.NewCert(certsMap))
}
func GetCertificateFunc(certsLocal ...*config.Cert) func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {

	if len(certsLocal) == 0 {

		return func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return currentCert.Load().GetCertificate(info)
		}
	}
	certList := make([]*config.Cert, 0, len(certsLocal))
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

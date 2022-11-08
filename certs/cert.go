package certs

import (
	"crypto/tls"
	"errors"
	"github.com/eolinker/eosc/config"
	"sync"
)

var errorCertificateNotExit = errors.New("not exist cert")

type ICert interface {
	SaveCert(workerId string, cert *tls.Certificate)
	DelCert(workerId string)
}

var (
	workerMaps               = make(map[string]*tls.Certificate)
	lock                     = sync.RWMutex{}
	currentCert *config.Cert = nil
)

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
	currentCert = config.NewCert(certsMap)
}
func GetCertificateFunc(certsLocal ...*config.Cert) func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	lock.RLock()
	certsCluster := currentCert
	lock.RUnlock()

	certList := make([]*config.Cert, 0, len(certsLocal)+1)
	for _, c := range certList {
		if c != nil {
			certList = append(certList, c)
		}
	}
	if certsCluster != nil {
		certList = append(certList, certsCluster)
	}
	if len(certList) == 0 {
		return func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return nil, errorCertificateNotExit
		}
	}
	if len(certList) == 1 {
		certs := certList[0]
		return func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certs.GetCertificate(info)
		}
	}

	return func(info *tls.ClientHelloInfo) (certificate *tls.Certificate, err error) {

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

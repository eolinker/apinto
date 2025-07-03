package certs

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/eolinker/eosc/config"
	"github.com/tjfoc/gmsm/gmtls"
)

var errorCertificateNotExit = errors.New("not exist cert")

type ICert interface {
	SaveCert(workerId string, cert []*gmtls.Certificate)
	DelCert(workerId string)
}

var (
	workerMaps = make(map[string][]*gmtls.Certificate)
	lock       = sync.RWMutex{}
	// currentCert 普通TLS证书
	currentCert = atomic.Pointer[config.Cert]{}
	// gmCert gmTLS证书
	gmCert = atomic.Pointer[config.Cert]{}
	// gmEncCert gmTLS加密证书
	gmEncCert = atomic.Pointer[config.Cert]{}
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

func SaveCert(workerId string, certs []*gmtls.Certificate) {
	lock.Lock()
	defer lock.Unlock()
	workerMaps[workerId] = certs
	rebuild()
}
func rebuild() {
	currentMap := make(map[string]*gmtls.Certificate)
	gmMap := make(map[string]*gmtls.Certificate)
	gmEncMap := make(map[string]*gmtls.Certificate)
	for _, cs := range workerMaps {
		l := len(cs)
		switch {
		case l == 1:
			i := cs[0]
			currentMap[i.Leaf.Subject.CommonName] = i
			for _, dnsName := range i.Leaf.DNSNames {
				currentMap[dnsName] = i
			}
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
	currentCert.Swap(config.NewCert(currentMap))
	gmCert.Swap(config.NewCert(gmMap))
	gmEncCert.Swap(config.NewCert(gmEncMap))
}

func GetCertificateFunc(certsLocal ...*config.Cert) func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {

	if len(certsLocal) == 0 {

		return func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
			return currentCert.Load().GetCertificate(info)
		}
	}
	certList := make([]*config.Cert, 0, len(certsLocal))
	for _, c := range certsLocal {
		if c != nil {
			certList = append(certList, c)
		}
	}
	return func(info *gmtls.ClientHelloInfo) (certificate *gmtls.Certificate, err error) {
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

func GetAutoCertificateFunc(certsLocal ...*config.Cert) func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
	return func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
		gmFlag := false
		// 检查支持协议中是否包含GMSSL
		for _, v := range info.SupportedVersions {
			if v == gmtls.VersionGMSSL {
				gmFlag = true
				break
			}
		}
		if !gmFlag {
			return GetCertificateFunc(certsLocal...)(info)
		}
		return gmCert.Load().GetCertificate(info)
	}
}

func GetKECertificate() func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
	return func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
		return gmEncCert.Load().GetCertificate(info)
	}
}

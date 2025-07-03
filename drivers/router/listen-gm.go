//go:build gm

package router

//
//import (
//	"crypto/tls"
//	"net"
//	"net/http"
//	"sync"
//
//	"github.com/tongsuo-project/tongsuo-go-sdk/crypto"
//
//	"github.com/eolinker/eosc/common/bean"
//	"github.com/eolinker/eosc/config"
//	"github.com/eolinker/eosc/log"
//	"github.com/eolinker/eosc/traffic"
//	"github.com/eolinker/eosc/traffic/mixl"
//	"github.com/soheilhy/cmux"
//
//	"github.com/eolinker/apinto/certs"
//
//	ts "github.com/tongsuo-project/tongsuo-go-sdk"
//)
//
//func init() {
//	matchWriters[AnyTCP] = matchersToMatchWriters(cmux.Any())
//
//	// 基本TLS
//	matchWriters[TlsTCP] = matchersToMatchWriters(cmux.TLS(
//		tls.VersionSSL30,
//		tls.VersionTLS10,
//		tls.VersionTLS11,
//		tls.VersionTLS12,
//		tls.VersionTLS13,
//		int(ts.NTLS),
//	))
//
//	//matchWriters[NTlsTCP] = matchersToMatchWriters(cmux.TLS(int(ts.NTLS)))
//
//	matchWriters[Http] = matchersToMatchWriters(cmux.HTTP1Fast(http.MethodPatch))
//	matchWriters[Dubbo2] = matchersToMatchWriters(cmux.PrefixMatcher(string([]byte{0xda, 0xbb})))
//	matchWriters[GRPC] = []cmux.MatchWriter{cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc")}
//	var tf traffic.ITraffic
//	var listenCfg *config.ListenUrl
//	bean.Autowired(&tf, &listenCfg)
//
//	bean.AddInitializingBeanFunc(func() {
//		initListener(tf, listenCfg)
//	})
//}
//func initListener(tf traffic.ITraffic, listenCfg *config.ListenUrl) {
//
//	if tf.IsStop() {
//		return
//	}
//
//	wg := sync.WaitGroup{}
//	tcp, ssl := tf.Listen(listenCfg.ListenUrls...)
//
//	listenerByPort := make(map[int][]net.Listener)
//	for _, l := range tcp {
//		port := readPort(l.Addr())
//		listenerByPort[port] = append(listenerByPort[port], l)
//	}
//	if len(ssl) > 0 {
//		tlsConfig := &g.Config{GetCertificate: certs.GetCertificateFunc()}
//
//		for _, l := range ssl {
//			//log.Debug("ssl listen: ", l.Addr().String())
//			port := readPort(l.Addr())
//			//listenerByPort[port] = append(listenerByPort[port], tls.NewListener(l, tlsConfig))
//			nTlsListen, err := newNTlsListen(l)
//			if err != nil {
//				log.Errorf("Error creating NTLS listener: %v\n", err)
//				continue
//			}
//			log.DebugF("NTLS listen: %s", nTlsListen.Addr().String())
//
//			listenerByPort[port] = append(listenerByPort[port], nTlsListen)
//
//		}
//	}
//	for port, lns := range listenerByPort {
//
//		var ln = mixl.NewMixListener(port, lns...)
//
//		wg.Add(1)
//		go func(ln net.Listener, p int) {
//			wg.Done()
//			cMux := cmux.New(ln)
//			for i, handler := range handlers {
//				log.Debug("i is ", i, " handler is ", handler)
//				if handler != nil {
//					go handler(p, cMux.MatchWithWriters(matchWriters[i]...))
//				}
//			}
//
//			cMux.Serve()
//		}(ln, port)
//
//	}
//	wg.Wait()
//	return
//}
//
//func newNTlsListen(l net.Listener) (net.Listener, error) {
//
//	ctx, err := ts.NewCtxWithVersion(ts.NTLS)
//	if err != nil {
//		log.Errorf("Error creating NTLS context: %v\n", err)
//		return nil, err
//	}
//	ctx.SetTLSExtServernameCallback(func(ssl *ts.SSL) ts.SSLTLSExtErr {
//		serverName := ssl.GetServername()
//		log.Infof("SNI: Client requested hostname: %s\n", serverName)
//		gmCert, has := certs.GetGMCertificate(serverName)
//		if !has {
//			return ts.SSLTLSExtErrAlertFatal
//		}
//		err = loadCertAndKeyForSSL(ssl, gmCert)
//		if err != nil {
//			log.Errorf("Error loading certificate for %s: %v\n", serverName, err)
//			return ts.SSLTLSExtErrAlertFatal
//		}
//
//		return ts.SSLTLSExtErrOK
//	})
//	err = loadDefaultCert(ctx)
//	if err != nil {
//		return nil, err
//	}
//	ctx.SetServerALPNProtos([]string{"h2", "http/1.1"})
//	ctx.SetCipherSuites("ECC-SM2-SM4-CBC-SM3")
//
//	return ts.NewListener(l, ctx), nil
//}
//
//// Load certificate and key for SSL
//func loadCertAndKeyForSSL(ssl *ts.SSL, gmCert *certs.GMCertificate) error {
//	ctx, err := ts.NewCtx()
//	if err != nil {
//		return err
//	}
//	err = ctx.UseSignCertificate(gmCert.SignCert)
//	if err != nil {
//		return err
//	}
//
//	err = ctx.UseEncryptCertificate(gmCert.EncCert)
//	if err != nil {
//		return err
//	}
//	err = ctx.UseSignPrivateKey(gmCert.SignKey)
//	if err != nil {
//		return err
//	}
//
//	err = ctx.UseEncryptPrivateKey(gmCert.EncKey)
//	if err != nil {
//		return err
//	}
//	ssl.SetSSLCtx(ctx)
//
//	return nil
//}
//
//func loadDefaultCert(ctx *ts.Ctx) error {
//	ec, err := crypto.LoadCertificateFromPEM([]byte(encCert))
//	if err != nil {
//		return err
//	}
//	err = ctx.UseEncryptCertificate(ec)
//	if err != nil {
//		return err
//	}
//	ek, err := crypto.LoadPrivateKeyFromPEM([]byte(encKey))
//	if err != nil {
//		return err
//	}
//	err = ctx.UseEncryptPrivateKey(ek)
//	if err != nil {
//		return err
//	}
//	sc, err := crypto.LoadCertificateFromPEM([]byte(signCert))
//	if err != nil {
//		return err
//	}
//	err = ctx.UseSignCertificate(sc)
//	if err != nil {
//		return err
//	}
//	sk, err := crypto.LoadPrivateKeyFromPEM([]byte(signKey))
//	if err != nil {
//		return err
//	}
//	return ctx.UseSignPrivateKey(sk)
//}
//
//const (
//	encCert = `-----BEGIN CERTIFICATE-----
//MIIB6zCCAZKgAwIBAgIUaNiS6WOsoEViDnmdb8Mdk3Qz5XwwCgYIKoEcz1UBg3Uw
//RTELMAkGA1UEBhMCQUExCzAJBgNVBAgMAkJCMQswCQYDVQQKDAJDQzELMAkGA1UE
//CwwCREQxDzANBgNVBAMMBnN1YiBjYTAgFw0yMzAyMjIwMjMwMTRaGA8yMTIzMDEy
//OTAyMzAxNFowSTELMAkGA1UEBhMCQUExCzAJBgNVBAgMAkJCMQswCQYDVQQKDAJD
//QzELMAkGA1UECwwCREQxEzARBgNVBAMMCnNlcnZlciBlbmMwWTATBgcqhkjOPQIB
//BggqgRzPVQGCLQNCAAR9vqVFQ0WBcr07aI5QnC31RYas4AtY7JQUmflKUKWMZ11v
//mtr/CJ6BN6djQ6zS81yjCopcz4G3zc5SZqAWueNko1owWDAJBgNVHRMEAjAAMAsG
//A1UdDwQEAwIDODAdBgNVHQ4EFgQUZ6Wt1ZR24FqcXla4hg/xOyju7FQwHwYDVR0j
//BBgwFoAUrGHrIoBiWQg+lsjRf850XAKvPJkwCgYIKoEcz1UBg3UDRwAwRAIgR1k1
//ecSt7I2335jEquFmHBE5pe8Sk/IqOqQS0Jvs1uYCIG5XMB0XeUaVb9OctaxgOQLN
//F8dRftiUHsyYXqfbaVjI
//-----END CERTIFICATE-----
//`
//	encKey = `-----BEGIN PRIVATE KEY-----
//MIGHAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBG0wawIBAQQgLrRk3CWTe+WZOFSf
//TMYwbOocLs3MSRpOO0/AvSmvH5mhRANCAAR9vqVFQ0WBcr07aI5QnC31RYas4AtY
//7JQUmflKUKWMZ11vmtr/CJ6BN6djQ6zS81yjCopcz4G3zc5SZqAWueNk
//-----END PRIVATE KEY-----
//`
//	signKey = `-----BEGIN PRIVATE KEY-----
//MIGHAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBG0wawIBAQQgeQTrKtO8mNXn/yvg
//R+pdbCgH5sl+WCFfXcqGl64soU2hRANCAAQFv/ruxAbI8/WApuOcUoR2wN8rYQZd
//SnT0dq8PtmiQ+JasxLIdiwNtE/F71NOCNJCL7bd/jj6uhwZU/G+oBI0M
//-----END PRIVATE KEY-----
//`
//	signCert = `-----BEGIN CERTIFICATE-----
//MIIB7jCCAZOgAwIBAgIUcbKTlc6+CNoHglmEk+xm+WIqZcAwCgYIKoEcz1UBg3Uw
//RTELMAkGA1UEBhMCQUExCzAJBgNVBAgMAkJCMQswCQYDVQQKDAJDQzELMAkGA1UE
//CwwCREQxDzANBgNVBAMMBnN1YiBjYTAgFw0yMzAyMjIwMjMwMTRaGA8yMTIzMDEy
//OTAyMzAxNFowSjELMAkGA1UEBhMCQUExCzAJBgNVBAgMAkJCMQswCQYDVQQKDAJD
//QzELMAkGA1UECwwCREQxFDASBgNVBAMMC3NlcnZlciBzaWduMFkwEwYHKoZIzj0C
//AQYIKoEcz1UBgi0DQgAEBb/67sQGyPP1gKbjnFKEdsDfK2EGXUp09HavD7ZokPiW
//rMSyHYsDbRPxe9TTgjSQi+23f44+rocGVPxvqASNDKNaMFgwCQYDVR0TBAIwADAL
//BgNVHQ8EBAMCBsAwHQYDVR0OBBYEFH3uBqkdowIvk//P7n5UtnpV9TR6MB8GA1Ud
//IwQYMBaAFKxh6yKAYlkIPpbI0X/OdFwCrzyZMAoGCCqBHM9VAYN1A0kAMEYCIQCz
//W/6Z/d/IJUTrO0o8nCxNle6R0AkRCKUFhW9zbIRlNwIhAJZxg4gs2cV2QF37oHs6
//9TD+MkRbql4Yb47+jLf8f247
//-----END CERTIFICATE-----
//`
//)

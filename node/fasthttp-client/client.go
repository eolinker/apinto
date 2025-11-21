package fasthttp_client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/eosc/eocontext"
	"github.com/valyala/fasthttp"
)

func ProxyTimeout(scheme string, host string, node eocontext.INode, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
	err := defaultClient.ProxyTimeout(addr, host, req, resp, timeout)
	if err != nil {
		node.Down()
	}
	return err
}

var defaultClient = NewClient()

const (
	DefaultMaxConns           = 10240
	DefaultMaxConnWaitTimeout = time.Second * 60
	DefaultMaxRedirectCount   = 2
)

// Client implements http client.
//
// Copying Client by value is prohibited. Create new instance instead.
//
// It is safe calling Client methods from concurrently running goroutines.
//
// The fields of a Client should not be changed while it is in use.
type Client struct {
	mLock  sync.RWMutex
	m      map[string]*fasthttp.HostClient
	ms     map[string]*fasthttp.HostClient
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient() *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		m:      make(map[string]*fasthttp.HostClient),
		ms:     make(map[string]*fasthttp.HostClient),
		ctx:    ctx,
		cancel: cancel,
	}
}

func readAddress(addr string) (scheme, host string) {
	if i := strings.Index(addr, "://"); i > 0 {
		return strings.ToLower(addr[:i]), addr[i+3:]
	}
	return "http", addr
}

func GenDialFunc(isTls bool) (fasthttp.DialFunc, error) {
	proxy := os.Getenv("http_proxy")
	if isTls {
		proxy = os.Getenv("https_proxy")
	}
	if proxy != "" {
		uri, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		return proxyDial(fmt.Sprintf("%s:%s", uri.Hostname(), uri.Port())), nil
	}
	return Dial, nil
}

func (c *Client) getHostClient(addr string, rewriteHost string) (*fasthttp.HostClient, string, error) {

	scheme, nodeAddr := readAddress(addr)
	host := nodeAddr
	isTLS := strings.EqualFold(scheme, "https")

	if !strings.EqualFold(scheme, "http") && !isTLS {
		return nil, "", fmt.Errorf("unsupported protocol %q. http and https are supported", scheme)
	}

	c.mLock.RLock()
	m := c.m
	if isTLS {
		m = c.ms
	}
	key := host
	hc := m[key]
	c.mLock.RUnlock()
	if hc != nil {
		return hc, scheme, nil
	}
	c.mLock.Lock()
	defer c.mLock.Unlock()

	if isTLS {
		m = c.ms
	} else {
		m = c.m
	}
	if hc == nil {
		dial, err := GenDialFunc(isTLS)
		if err != nil {
			return nil, "", err
		}
		dialAddr := addMissingPort(nodeAddr, isTLS)

		httpAddr := dialAddr
		if isTLS {
			if rewriteHost != "" && rewriteHost != nodeAddr {
				httpAddr = rewriteHost
				dial = func(addr string) (net.Conn, error) {
					proxy := os.Getenv("https_proxy")
					if proxy != "" {
						uri, err := url.Parse(proxy)
						if err != nil {
							return nil, err
						}
						return proxyDial(fmt.Sprintf("%s:%s", uri.Hostname(), uri.Port()))(addr)
					}
					return Dial(dialAddr)
				}
			}
		}

		hc = &fasthttp.HostClient{
			Addr:  httpAddr,
			IsTLS: isTLS,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},

			Dial:                dial,
			StreamResponseBody:  true,
			MaxConns:            DefaultMaxConns,
			MaxIdleConnDuration: 0,

			// 重试配置：针对 ErrConnectionClosed 自动重试
			MaxIdemponentCallAttempts: 3, // 最大重试次数（默认 3），适用于幂等请求（如 GET）
			RetryIfErr: func(req *fasthttp.Request, attempts int, err error) (resetTimeout bool, retry bool) {
				if errors.Is(err, io.EOF) { // 针对你的错误重试
					return true, true // 重试并重置超时
				}
				return false, false
			},

			ConnPoolStrategy: fasthttp.LIFO,
		}
		//http2.ConfigureClient(hc, http2.ClientOpts{})
		m[key] = hc
		if len(m) == 1 {
			go c.startCleaner(m)
		}
	}

	return hc, scheme, nil
}

// ProxyTimeout performs the given request and waits for response during
// the given timeout duration.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// Response is ignored if resp is nil.
//
// ErrTimeout is returned if the response wasn't returned during
// the given timeout.
//
// ErrNoFreeConns is returned if all Client.MaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
//
// Warning: ProxyTimeout does not terminate the request itself. The request will
// continue in the background and the response will be discarded.
// If requests take too long and the connection pool gets filled up please
// try setting a ReadTimeout.
func (c *Client) ProxyTimeout(addr string, host string, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	request := req
	request.Header.ResetConnectionClose()
	request.Header.Set("Connection", "keep-alive")
	connectionClose := resp.ConnectionClose()
	defer func() {
		if connectionClose {
			resp.SetConnectionClose()
		}
	}()

	client, scheme, err := c.getHostClient(addr, host)
	if err != nil {
		return err
	}
	request.URI().SetScheme(scheme)

	return client.DoTimeout(req, resp, timeout)

}

func (c *Client) startCleaner(m map[string]*fasthttp.HostClient) {
	sleep := time.Second * 10
	mustStop := false
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(sleep):
			c.mLock.Lock()
			for k, v := range m {
				if v.ConnsCount() == 0 {
					v.CloseIdleConnections()
					delete(m, k)
				}
			}

			if len(m) == 0 {
				mustStop = true
			}
			c.mLock.Unlock()
			if mustStop {
				return
			}
		}
	}
}

func (c *Client) mCleaner(m map[string]*fasthttp.HostClient) {
	mustStop := false

	//sleep := c.MaxIdleConnDuration
	//if sleep < time.Second {
	//	sleep = time.Second
	//} else if sleep > 10*time.Second {
	//	sleep = 10 * time.Second
	//}
	sleep := time.Second * 10
	for {
		c.mLock.Lock()
		for k, v := range m {
			shouldRemove := v.ConnsCount() == 0
			if shouldRemove {
				delete(m, k)
			}
		}
		if len(m) == 0 {
			mustStop = true
		}

		c.mLock.Unlock()

		if mustStop {
			break
		}
		time.Sleep(sleep)
	}
}

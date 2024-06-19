package fasthttp_client

import (
	"fmt"
	"net"
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

var defaultClient Client

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
	mLock sync.Mutex
	m     map[string]*fasthttp.HostClient
	ms    map[string]*fasthttp.HostClient
}

func readAddress(addr string) (scheme, host string) {
	if i := strings.Index(addr, "://"); i > 0 {
		return strings.ToLower(addr[:i]), addr[i+3:]
	}
	return "http", addr
}

func (c *Client) getHostClient(addr string, rewriteHost string) (*fasthttp.HostClient, string, error) {

	scheme, nodeAddr := readAddress(addr)
	host := nodeAddr
	isTLS := false
	if strings.EqualFold(scheme, "https") {
		isTLS = true
		host = fmt.Sprintf("%s-%s", rewriteHost, nodeAddr)

	} else if !strings.EqualFold(scheme, "http") {
		return nil, "", fmt.Errorf("unsupported protocol %q. http and https are supported", scheme)
	}

	startCleaner := false

	c.mLock.Lock()
	m := c.m
	if isTLS {
		m = c.ms
	}
	if m == nil {
		m = make(map[string]*fasthttp.HostClient)
		if isTLS {
			c.ms = m
		} else {
			c.m = m
		}
	}
	hc := m[host]
	if hc == nil {
		dial := Dial
		dialAddr := addMissingPort(nodeAddr, isTLS)
		httpAddr := dialAddr
		if isTLS {
			if rewriteHost != "" && rewriteHost != nodeAddr {

				httpAddr = rewriteHost

				dial = func(addr string) (net.Conn, error) {
					return Dial(dialAddr)
				}
			}
		}

		hc = &fasthttp.HostClient{
			Addr:               httpAddr,
			IsTLS:              isTLS,
			Dial:               dial,
			MaxConns:           DefaultMaxConns,
			MaxConnWaitTimeout: DefaultMaxConnWaitTimeout,
			RetryIf: func(request *fasthttp.Request) bool {
				return false
			},
		}
		m[host] = hc
		if len(m) == 1 {
			startCleaner = true
		}
	}
	c.mLock.Unlock()

	if startCleaner {
		go c.mCleaner(m)
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
	//var deadline time.Time
	//var requestURI string
	//redirectCount := 0
	//for {
	client, scheme, err := c.getHostClient(addr, host)
	if err != nil {
		return err
	}

	request.URI().SetScheme(scheme)

	//if redirectCount == 0 {
	//	deadline = time.Now().Add(timeout)
	//} else {
	//	request.SetRequestURI(requestURI)
	//}

	return client.DoTimeout(req, resp, timeout)
	//if err != nil {
	//	return err
	//}
	//if !fasthttp.StatusCodeIsRedirect(resp.StatusCode()) || redirectCount >= DefaultMaxRedirectCount {
	//	break
	//}
	//redirectCount++
	//location := resp.Header.Peek("Location")
	//if len(location) == 0 {
	//	return fasthttp.ErrMissingLocation
	//}
	//addr, requestURI = getRedirectURL(req.URI().String(), location)
	//}

	//return nil
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

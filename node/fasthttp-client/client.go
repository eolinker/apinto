package fasthttp_client

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

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
// ErrNoFreeConns is returned if all DefaultMaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
//
// Warning: ProxyTimeout does not terminate the request itself. The request will
// continue in the background and the response will be discarded.
// If requests take too long and the connection pool gets filled up please
// try using a Client and setting a ReadTimeout.
func ProxyTimeout(addr string, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return defaultClient.ProxyTimeout(addr, req, resp, timeout)
}

var defaultClient Client

// Client implements http client.
//
// Copying Client by value is prohibited. Create new instance instead.
//
// It is safe calling Client methods from concurrently running goroutines.
//
// The fields of a Client should not be changed while it is in use.
type Client struct {

	// Client name. Used in User-Agent request header.
	//
	// Default client name is used if not set.
	Name string

	// NoDefaultUserAgentHeader when set to true, causes the default
	// User-Agent header to be excluded from the Request.
	NoDefaultUserAgentHeader bool

	// Callback for establishing new connections to hosts.
	//
	// Default Dial is used if not set.
	Dial fasthttp.DialFunc

	// Attempt to connect to both ipv4 and ipv6 addresses if set to true.
	//
	// This option is used only if default TCP dialer is used,
	// i.e. if Dial is blank.
	//
	// By default client connects only to ipv4 addresses,
	// since unfortunately ipv6 remains broken in many networks worldwide :)
	DialDualStack bool

	// TLS config for https connections.
	//
	// Default TLS config is used if not set.
	TLSConfig *tls.Config

	// Maximum number of connections per each host which may be established.
	//
	// DefaultMaxConnsPerHost is used if not set.
	MaxConnsPerHost int

	// Idle keep-alive connections are closed after this duration.
	//
	// By default idle connections are closed
	// after DefaultMaxIdleConnDuration.
	MaxIdleConnDuration time.Duration

	// Keep-alive connections are closed after this duration.
	//
	// By default connection duration is unlimited.
	MaxConnDuration time.Duration

	// Maximum number of attempts for idempotent calls
	//
	// DefaultMaxIdemponentCallAttempts is used if not set.
	MaxIdemponentCallAttempts int

	// Per-connection buffer size for responses' reading.
	// This also limits the maximum header size.
	//
	// Default buffer size is used if 0.
	ReadBufferSize int

	// Per-connection buffer size for requests' writing.
	//
	// Default buffer size is used if 0.
	WriteBufferSize int

	// Maximum duration for full response reading (including body).
	//
	// By default response read timeout is unlimited.
	ReadTimeout time.Duration

	// Maximum duration for full request writing (including body).
	//
	// By default request write timeout is unlimited.
	WriteTimeout time.Duration

	// Maximum response body size.
	//
	// The client returns ErrBodyTooLarge if this limit is greater than 0
	// and response body is greater than the limit.
	//
	// By default response body size is unlimited.
	MaxResponseBodySize int

	// Header names are passed as-is without normalization
	// if this option is set.
	//
	// Disabled header names' normalization may be useful only for proxying
	// responses to other clients expecting case-sensitive
	// header names. See https://github.com/valyala/fasthttp/issues/57
	// for details.
	//
	// By default request and response header names are normalized, i.e.
	// The first letter and the first letters following dashes
	// are uppercased, while all the other letters are lowercased.
	// Examples:
	//
	//     * HOST -> Host
	//     * content-type -> Content-Type
	//     * cONTENT-lenGTH -> Content-Length
	DisableHeaderNamesNormalizing bool

	// Path values are sent as-is without normalization
	//
	// Disabled path normalization may be useful for proxying incoming requests
	// to servers that are expecting paths to be forwarded as-is.
	//
	// By default path values are normalized, i.e.
	// extra slashes are removed, special characters are encoded.
	DisablePathNormalizing bool

	// Maximum duration for waiting for a free connection.
	//
	// By default will not waiting, return ErrNoFreeConns immediately
	MaxConnWaitTimeout time.Duration

	// RetryIf controls whether a retry should be attempted after an error.
	//
	// By default will use isIdempotent function
	RetryIf fasthttp.RetryIfFunc

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

func (c *Client) getHostClient(addr string) (*fasthttp.HostClient, string, error) {

	scheme, host := readAddress(addr)

	isTLS := false
	if strings.EqualFold(scheme, "https") {
		isTLS = true
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
		hc = &fasthttp.HostClient{
			Addr:                          addMissingPort(host, isTLS),
			Name:                          c.Name,
			NoDefaultUserAgentHeader:      c.NoDefaultUserAgentHeader,
			Dial:                          c.Dial,
			DialDualStack:                 c.DialDualStack,
			IsTLS:                         isTLS,
			TLSConfig:                     c.TLSConfig,
			MaxConns:                      c.MaxConnsPerHost,
			MaxIdleConnDuration:           c.MaxIdleConnDuration,
			MaxConnDuration:               c.MaxConnDuration,
			MaxIdemponentCallAttempts:     c.MaxIdemponentCallAttempts,
			ReadBufferSize:                c.ReadBufferSize,
			WriteBufferSize:               c.WriteBufferSize,
			ReadTimeout:                   c.ReadTimeout,
			WriteTimeout:                  c.WriteTimeout,
			MaxResponseBodySize:           c.MaxResponseBodySize,
			DisableHeaderNamesNormalizing: c.DisableHeaderNamesNormalizing,
			DisablePathNormalizing:        c.DisablePathNormalizing,
			MaxConnWaitTimeout:            c.MaxConnWaitTimeout,
			RetryIf:                       c.RetryIf,
		}
		m[string(host)] = hc
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
func (c *Client) ProxyTimeout(addr string, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	client, scheme, err := c.getHostClient(addr)
	if err != nil {
		return err
	}

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	req.CopyTo(request)
	request.URI().SetScheme(scheme)
	request.Header.ResetConnectionClose()
	return client.DoTimeout(req, resp, timeout)

}

func (c *Client) mCleaner(m map[string]*fasthttp.HostClient) {
	mustStop := false

	sleep := c.MaxIdleConnDuration
	if sleep < time.Second {
		sleep = time.Second
	} else if sleep > 10*time.Second {
		sleep = 10 * time.Second
	}

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

func addMissingPort(addr string, isTLS bool) string {
	n := strings.Index(addr, ":")
	if n >= 0 {
		return addr
	}
	port := 80
	if isTLS {
		port = 443
	}
	return net.JoinHostPort(addr, strconv.Itoa(port))
}

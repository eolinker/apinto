package websocket

import (
	"net/http"
	"net/url"
	"time"

	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"

	"github.com/fasthttp/websocket"
)

var dialer = &websocket.Dialer{
	Proxy:            http.ProxyFromEnvironment,
	HandshakeTimeout: 45 * time.Second,
}

var skipHeaders = []string{
	"Upgrade",
	"Connection",
	"Sec-Websocket-Key",
	"Sec-Websocket-Version",
	"Sec-Websocket-Extensions",
	"Sec-Websocket-Protocol",
}

func DialWithTimeout(node eocontext.INode, scheme, path string, query string, header http.Header, timeout time.Duration) (*websocket.Conn, *http.Response, error) {
	log.Debug("node: ", node.Addr())
	u := url.URL{Scheme: scheme, Host: node.Addr(), Path: path, RawQuery: query}
	dialer.HandshakeTimeout = timeout
	for _, key := range skipHeaders {
		header.Del(key)
	}
	dial, h, err := dialer.Dial(u.String(), header)
	if err != nil {
		node.Down()
	}
	return dial, h, err
}

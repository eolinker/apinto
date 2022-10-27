package websocket

import (
	"net/http"
	"time"

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

func DialWithTimeout(urlStr string, header http.Header, timeout time.Duration) (*websocket.Conn, *http.Response, error) {
	dialer.HandshakeTimeout = timeout
	for _, key := range skipHeaders {
		header.Del(key)
	}
	return dialer.Dial(urlStr, header)
}

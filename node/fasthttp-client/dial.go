package fasthttp_client

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eolinker/eosc/debug"
	"github.com/valyala/fasthttp"
)

var (
	tcpDial = &fasthttp.TCPDialer{
		Concurrency: 1000,
	}

	lock       sync.Mutex
	dialCount  int64
	closeCount int64
	lists      = make([]CountItem, 0, 100)
)

type CountItem struct {
	Time       string `json:"time"`
	DialCount  int64  `json:"dial_count"`
	CloseCount int64  `json:"close_count"`
}

func init() {
	debug.Register("/debug/dial", DebugHandleFun)
	go reset()
}

func reset() {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			dv := atomic.SwapInt64(&dialCount, 0)
			cv := atomic.SwapInt64(&closeCount, 0)
			if dv == 0 && cv == 0 {
				continue
			}
			lock.Lock()
			c := len(lists)
			if c > 99 {
				copy(lists, lists[1:])
				lists = lists[:c-1]
			}
			lists = append(lists, CountItem{
				DialCount:  dv,
				CloseCount: cv,
				Time:       time.Now().Format("22006-01-02 15:04:05"),
			})
			lock.Unlock()
		}
	}
}
func DebugHandleFun(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	fmt.Fprintf(w, "proxy dial:")
	for _, v := range lists {
		fmt.Fprintf(w, "%s %d : %d\n", v.Time, v.DialCount, v.CloseCount)
	}
}

const DefaultDialTimeout = time.Second * 10

func Dial(addr string) (net.Conn, error) {
	atomic.AddInt64(&dialCount, 1)
	conn, err := tcpDial.DialTimeout(addr, DefaultDialTimeout)
	if err != nil {
		return nil, err
	}
	//return conn, nil
	return &debugConn{Conn: conn}, nil
}

type debugConn struct {
	net.Conn
}

func (c *debugConn) Close() error {
	atomic.AddInt64(&closeCount, 1)
	return c.Conn.Close()
}
func addMissingPort(addr string, isTLS bool) string {

	n := strings.LastIndex(addr, ":")
	if n >= 0 {
		return addr
	}
	port := 80
	if isTLS {
		port = 443
	}
	return net.JoinHostPort(addr, strconv.Itoa(port))
}
func readPort(addr string) int {
	n := strings.LastIndex(addr, ":")
	if n >= 0 {
		p, e := strconv.Atoi(addr[n+1:])
		if e != nil {
			return p
		}
	}
	return 0
}
func getRedirectURL(baseURL string, location []byte) (string, string) {
	u := fasthttp.AcquireURI()
	u.Update(baseURL)
	u.UpdateBytes(location)
	u.RequestURI()
	defer fasthttp.ReleaseURI(u)
	return fmt.Sprintf("%s://%s", u.Scheme(), u.Host()), u.String()
}

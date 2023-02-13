package manager

import (
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/remoting"
	gxbytes "github.com/dubbogo/gost/bytes"
	"github.com/eolinker/eosc/log"
	perrors "github.com/pkg/errors"
	"io"
	"net"
	"sync"
)

const maxReadBufLen = 4 * 1024

type ConnResult struct {
	conn   net.Conn
	result *remoting.DecodeResult
}

func (c *ConnResult) Conn() net.Conn {
	return c.conn
}

func (c *ConnResult) Result() *remoting.DecodeResult {
	return c.result
}

type ConnHandler struct {
	lock     *sync.Mutex
	connMaps map[string]net.Conn
	result   chan *ConnResult
}

func NewConnHandler() *ConnHandler {
	return &ConnHandler{
		lock:     new(sync.Mutex),
		connMaps: make(map[string]net.Conn),
		result:   make(chan *ConnResult, 0),
	}
}

func (p *ConnHandler) Handler(conn net.Conn) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.connMaps[conn.RemoteAddr().String()] = conn
	p.task(conn)
}

func (p *ConnHandler) GetResult() <-chan *ConnResult {
	return p.result
}

func (p *ConnHandler) setResult(v *ConnResult) {
	p.result <- v
}

func (p *ConnHandler) task(conn net.Conn) {
	go func() {
		var (
			pkgLen   int
			buf      []byte
			err      error
			ok       bool
			netError net.Error
			pkg      *remoting.DecodeResult
		)
		for {
			pktBuf := gxbytes.NewBuffer(nil)
			var bufLen = 0
			for {
				reader := io.Reader(conn)
				buf = pktBuf.WriteNextBegin(maxReadBufLen)
				bufLen, err = reader.Read(buf)
				if err != nil {
					if netError, ok = perrors.Cause(err).(net.Error); ok && netError.Timeout() {
						break
					}
					if perrors.Cause(err) == io.EOF {
						log.Infof("session.conn read EOF, client send over, session exit")
						err = nil
						if bufLen != 0 {
							log.Infof("session.conn read EOF, while the bufLen(%d) is non-zero.")
						}
						break
					}

					log.Errorf("[session.conn.read] = error:%+v", perrors.WithStack(err))
					return

				}
				break
			}

			if 0 != bufLen {
				pktBuf.WriteNextEnd(bufLen)
				for {
					if pktBuf.Len() <= 0 {
						break
					}
					codec := dubbo.DubboCodec{}

					pkg, pkgLen, err = codec.Decode(pktBuf.Bytes())

					if err == nil && pkgLen > maxReadBufLen {
						err = perrors.Errorf("pkgLen %d > session max message len %d", pkgLen, maxReadBufLen)
					}
					if err != nil {
						break
					}
					if pkg == nil {
						break
					}
					result := &ConnResult{
						conn:   conn,
						result: pkg,
					}
					p.setResult(result)
					pktBuf.Next(pkgLen)
				}
			}
		}
	}()

}

package manager

import (
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/remoting"
	gxbytes "github.com/dubbogo/gost/bytes"
	"github.com/eolinker/eosc/log"
	"github.com/pkg/errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

const maxReadBufLen = 4 * 1024

type dubbo2Server struct {
	lock *sync.Mutex
	conn []net.Conn
	stop int32
}

func NewDubbo2Server() *dubbo2Server {
	return &dubbo2Server{
		lock: new(sync.Mutex),
	}
}

func (d *dubbo2Server) Handler(port int, conn net.Conn) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if atomic.LoadInt32(&d.stop) == 1 {
		conn.Close()
		return
	}

	d.conn = append(d.conn, conn)

	d.task(port, conn)
}

func (d *dubbo2Server) ShutDown() {
	d.lock.Lock()
	defer d.lock.Unlock()

	atomic.StoreInt32(&d.stop, 1)
	defer atomic.StoreInt32(&d.stop, 0)

	for _, conn := range d.conn {
		conn.Close()
	}

	d.conn = nil
}

func (d *dubbo2Server) task(port int, conn net.Conn) {
	go func() {
		var (
			pkgLen   int
			buf      []byte
			err      error
			ok       bool
			netError net.Error
		)
		pktBuf := gxbytes.NewBuffer(nil)

		for {

			var bufLen = 0
			for {
				reader := io.Reader(conn)
				buf = pktBuf.WriteNextBegin(maxReadBufLen)
				bufLen, err = reader.Read(buf)
				if err != nil {
					if netError, ok = errors.Cause(err).(net.Error); ok && netError.Timeout() {
						break
					}
					if errors.Cause(err) == io.EOF {
						log.Infof("session.conn read EOF, client send over, session exit")
						err = nil
						if bufLen != 0 {
							log.Infof("session.conn read EOF, while the bufLen(%d) is non-zero.", bufLen)
							break
						}
						return
					}

					log.Errorf("[session.conn.read] = error:%+v", errors.WithStack(err))
					return

				}
				break
			}

			if 0 != bufLen {
				go func() {
					pktBuf.WriteNextEnd(bufLen)
					for {
						if pktBuf.Len() <= 0 {
							break
						}
						codec := dubbo.DubboCodec{}
						var pkg *remoting.DecodeResult

						pkg, pkgLen, err = codec.Decode(pktBuf.Bytes())

						if err == nil && pkgLen > maxReadBufLen {
							err = errors.Errorf("pkgLen %d > session max message len %d", pkgLen, maxReadBufLen)
						}
						pktBuf.Next(pkgLen)
						if err != nil {
							break
						}
						if pkg == nil {
							break
						}

						manager.Handler(port, conn, pkg)

					}
				}()
			}
		}
	}()

}

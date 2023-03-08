package conn_clone

import (
	"container/list"
	"io"
	"net"
	"sync"
	"time"
)

type ConnPip struct {
	net.Conn
	reader io.Reader
}

func (c *ConnPip) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

type ConnReader struct {
	conn   net.Conn
	reader io.Reader
}

func (c *ConnReader) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (c *ConnReader) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (c *ConnReader) Close() error {
	return nil
}

func (c *ConnReader) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *ConnReader) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ConnReader) SetDeadline(t time.Time) error {
	return nil
}

func (c *ConnReader) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *ConnReader) SetWriteDeadline(t time.Time) error {
	return nil
}

func Clone(conn net.Conn) (rw net.Conn, r net.Conn) {
	pipeRw, writerRw := io.Pipe()
	pipeR, writerR := io.Pipe()

	go copyTo(conn, writerRw, writerR)
	return &ConnPip{
			Conn:   conn,
			reader: pipeRw,
		}, &ConnReader{
			conn:   conn,
			reader: pipeR,
		}
}

type blockChan struct {
	ls    *list.List
	lock  sync.Mutex
	w     *io.PipeWriter
	isRun bool
}

func newBlockChan(w *io.PipeWriter) *blockChan {
	bc := &blockChan{w: w, ls: list.New()}
	go rc(bc)
	return bc
}
func rc(bc *blockChan) {

	for {
		bc.lock.Lock()

		e := bc.ls.Front()
		if e == nil {
			bc.isRun = false
			bc.lock.Unlock()
			return
		}
		d := bc.ls.Remove(e).(*block)
		bc.lock.Unlock()
		data, n, err := d.Data()

		for i := 0; i < n; {
			nw, _ := bc.w.Write(data[i:])
			i += nw
		}
		d.Release()
		if err != nil {
			bc.w.CloseWithError(err)
			return
		}

	}

}
func (b *blockChan) write(d *block) {
	b.lock.Lock()

	b.ls.PushBack(d)
	if !b.isRun {
		b.isRun = true
		go rc(b)
	}
	b.lock.Unlock()

}

func copyTo(conn net.Conn, ws ...*io.PipeWriter) {

	cs := make([]*blockChan, 0, len(ws))
	for _, w := range ws {
		cs = append(cs, newBlockChan(w))
	}
	for {

		bc := readBlock(conn)
		for _, c := range cs {
			c.write(bc.Clone())
		}
		bc.Release()
	}
}

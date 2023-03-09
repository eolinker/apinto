package clone

import (
	"io"
	"net"
	"time"
)

type connPip struct {
	net.Conn
	reader io.Reader
}

func (c *connPip) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

type connReader struct {
	conn   net.Conn
	reader io.Reader
}

func (c *connReader) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (c *connReader) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (c *connReader) Close() error {
	return nil
}

func (c *connReader) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *connReader) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connReader) SetDeadline(t time.Time) error {
	return nil
}

func (c *connReader) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *connReader) SetWriteDeadline(t time.Time) error {
	return nil
}

func CloneConn(conn net.Conn, count int) (rw net.Conn, r []net.Conn) {

	if count < 1 {
		count = 1
	}

	rs := Clone(conn, count+1)

	cw := &connPip{
		Conn:   conn,
		reader: rs[0],
	}
	rs = rs[1:]
	crs := make([]net.Conn, count)
	for _, r := range rs {
		crs = append(crs, &connReader{
			conn:   conn,
			reader: r,
		})
	}
	return cw, crs

}

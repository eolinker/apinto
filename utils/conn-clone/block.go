package conn_clone

import (
	"io"
	"sync"
)

const BuffSize = 4096

var (
	poolBuff = sync.Pool{
		New: func() any {
			return &block{
				buf: make([]byte, BuffSize),
			}
		},
	}
)

type block struct {
	buf []byte
	n   int
	err error
}

func readBlock(r io.Reader) *block {
	b := acquireBlock()
	b.n, b.err = r.Read(b.buf)
	return b
}
func (b *block) Release() {
	b.n = 0
	b.err = nil
	poolBuff.Put(b)
}
func (b *block) Clone() *block {
	c := acquireBlock()
	copy(b.buf, c.buf)
	c.n, c.err = b.n, b.err
	return c
}
func (b *block) Data() ([]byte, int, error) {

	return b.buf[:b.n], b.n, b.err
}
func acquireBlock() *block {
	return poolBuff.Get().(*block)
}

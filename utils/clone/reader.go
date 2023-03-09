package clone

import (
	"container/list"
	"io"
	"sync"
)

func Clone(r io.Reader, size int) []io.Reader {

	if size <= 1 {
		return []io.Reader{r}
	}

	rs := make([]io.Reader, size)
	ws := make([]*io.PipeWriter, size)
	for i := 0; i < size; i++ {
		reader, writer := io.Pipe()
		rs = append(rs, reader)
		ws = append(ws, writer)
	}
	go copyTo(r, ws...)
	return rs
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

func copyTo(in io.Reader, ws ...*io.PipeWriter) {

	cs := make([]*blockChan, 0, len(ws))
	for _, w := range ws {
		cs = append(cs, newBlockChan(w))
	}
	for {

		bc := readBlock(in)
		for _, c := range cs {
			c.write(bc.Clone())
		}
		bc.Release()
	}
}

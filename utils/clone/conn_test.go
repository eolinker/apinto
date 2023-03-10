package clone

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func TestCloneTcp(t *testing.T) {

	wg := sync.WaitGroup{}

	listen, err := net.Listen("tcp", ":9988")
	if err != nil {
		panic(err)
	}
	go func() {

		index := 0
		for {
			index++
			accept, err := listen.Accept()
			if err != nil {

				return
			}
			wg.Add(1)

			if index%2 == 0 {
				go func() {
					read(accept, "org..", t)
					wg.Done()
				}()
			} else {

				go func() {
					rw, r := CloneConn(accept, 1)
					go read(r[0], "clone", t)
					read(rw, "main.", t)

					wg.Done()
				}()

			}
		}
	}()
	writeTcp("cle 1", "127.0.0.1:9988", t)
	//write("cle 2", t)
	wg.Wait()
	listen.Close()
}
func writeTcp(name string, addr string, t *testing.T) {
	buf := bytes.Buffer{}
	for i := 0; i < 4096; i++ {
		buf.WriteString(fmt.Sprint(rand.Int(), ","))
	}
	data := buf.Bytes()
	conn, err := net.Dial("tcp", addr)
	if err != nil {

		return
	}

	go doWrite(name, conn, data, t)

}
func doWrite(name string, w io.WriteCloser, data []byte, testing *testing.T) {

	tc := time.NewTicker(time.Second)
	te := time.NewTimer(time.Second * 5)
	total := int64(0)
	last := int64(0)
	for {
		select {
		case t := <-tc.C:
			testing.Logf("\t[%s] %s\twrite %dk/s\n", t.Format(time.RFC3339), name, (total-last)/1024)
			last = total
		case <-te.C:
			w.Close()
			return
		default:
			n, err := w.Write(data)
			if err != nil {
				return
			}
			total += int64(n)
		}

	}
}
func read(r io.Reader, name string, testing *testing.T) {

	rb := bufio.NewReader(r)
	buf := make([]byte, 4096)
	total := int64(0)
	last := int64(0)
	tc := time.NewTicker(time.Second)
	defer tc.Stop()
	for {

		select {
		case t := <-tc.C:
			testing.Logf("\t[%s] %s\tread %dk/s\n", t.Format(time.RFC3339), name, (total-last)/1024)
			last = total
		default:
			r, err := rb.Read(buf)
			if err != nil {
				return
			}
			total += int64(r)
		}
	}
}

func TestClonePipe(t *testing.T) {
	buf := bytes.Buffer{}
	for i := 0; i < 8192; i++ {
		buf.WriteString(fmt.Sprint(rand.Int(), ","))
	}
	data := buf.Bytes()
	reader, writer := io.Pipe()

	readers := Clone(reader, 3)
	wg := sync.WaitGroup{}

	for i, r := range readers {
		wg.Add(1)
		go func(i int, r io.Reader) {
			read(r, fmt.Sprintf("reader-%d", i+1), t)
			wg.Done()
		}(i, r)
	}

	doWrite("write", writer, data, t)
}

package conn_clone

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func TestClone(t *testing.T) {

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
					rw, r := Clone(accept)
					go read(r, "clone", t)
					read(rw, "main.", t)

					wg.Done()
				}()

			}
		}
	}()
	write("cle 1", t)
	write("cle 2", t)
	wg.Wait()
	listen.Close()
}
func write(name string, testing *testing.T) {
	buf := bytes.Buffer{}
	for i := 0; i < 4096; i++ {
		buf.WriteString(fmt.Sprint(rand.Int(), ","))
	}

	conn, err := net.Dial("tcp", "127.0.0.1:9988")
	if err != nil {

		return
	}

	go func() {
		data := buf.Bytes()
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
				conn.Close()
				return
			default:
				n, err := conn.Write(data)
				if err != nil {
					return
				}
				total += int64(n)
			}

		}
	}()

}
func read(conn net.Conn, name string, testing *testing.T) {

	rb := bufio.NewReader(conn)
	buf := make([]byte, 1024)
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

package httplog

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type _HttpWriter struct {
	cancelFunc context.CancelFunc

	locker  sync.Mutex
	client  *http.Client
	chanOut chan []byte
}

func (h *_HttpWriter) reset(c *Config) {

	h.locker.Lock()
	h.stop()
	ctx, cancelFunc := context.WithCancel(context.Background())
	h.cancelFunc = cancelFunc
	handlerCount := c.HandlerCount

	if h.chanOut == nil {
		h.chanOut = make(chan []byte, handlerCount*5)
	}
	if handlerCount <= 0 {
		handlerCount = 2
	}
	for i := 0; i < handlerCount; i++ {
		go h.do(c.Method, c.Url, c.Headers, ctx)
	}

	h.locker.Unlock()
}
func (h *_HttpWriter) stop() error {

	if h.cancelFunc != nil {
		h.cancelFunc()
		h.cancelFunc = nil
	}

	return nil
}
func (h *_HttpWriter) Close() error {
	h.locker.Lock()
	h.stop()
	close(h.chanOut)
	h.chanOut = nil
	h.locker.Unlock()
	return nil
}

func (h *_HttpWriter) do(method, url string, header http.Header, ctx context.Context) {
	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(error); ok {
				fmt.Println("[httplog] do send error:", err)
				debug.PrintStack()
			}
			go h.do(method, url, header, ctx)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-h.chanOut:
			if ok {
				h.send(method, url, header, p)
			}
		}
	}
}
func newHttpWriter() *_HttpWriter {

	return &_HttpWriter{

		locker: sync.Mutex{},
		client: &http.Client{

			Timeout: time.Second * 30,
		},
		chanOut: nil,
	}
}

func (h *_HttpWriter) Write(p []byte) (n int, err error) {
	h.chanOut <- p
	return len(p), nil
}

func (h *_HttpWriter) send(method, url string, header http.Header, p []byte) {
	request, err := http.NewRequest(method, url, bytes.NewReader(p))

	if err != nil {
		fmt.Printf("[httplog]:create requeset error: %s %s header<%v> body<%s> error<%s>\n", method, url, header, string(p), err)
		return
	}
	for k, v := range header {
		request.Header[k] = v
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := h.client.Do(request)
	if err != nil {
		fmt.Printf("[httplog]:send error: %s %s header<%v> body<%s> error<%s>\n", method, url, header, string(p), err)
		return
	}
	if response.StatusCode != 200 {
		body, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			fmt.Printf("send to httplog error:%s %s status<%d,%s> :%s\n", method, url, response.StatusCode, response.Status, err.Error())
		} else {
			fmt.Printf("[httplog] response error:%s %s status<%d,%s> body:%s\n", method, url, response.StatusCode, response.Status, string(body))

		}
	}

}

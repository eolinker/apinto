package http_context

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

//ResponseReader 响应结构体
type ResponseReader struct {
	*CookiesHandler
	*Header
	*BodyHandler
	*StatusHandler
}

func newResponseReader(response *http.Response) *ResponseReader {
	if response == nil {
		return nil
	}
	r := new(ResponseReader)
	r.Header = NewHeader(response.Header)
	r.CookiesHandler = newCookieHandle(response.Header)
	r.StatusHandler = NewStatusHandler()
	r.SetStatus(response.StatusCode, response.Status)
	// if response.ContentLength > 0 {
	// 	body, _ := ioutil.ReadAll(response.body)
	// 	r.BodyHandler = NewBodyHandler(body)
	// } else {
	// 	r.BodyHandler = NewBodyHandler(nil)
	// }
	body, _ := ioutil.ReadAll(response.Body)
	r.BodyHandler = NewBodyHandler(body)

	return r
}

//NewResponseReader 新增ResponseReader
func NewResponseReader(header *fasthttp.ResponseHeader, statusCode int, body []byte) *ResponseReader {
	r := new(ResponseReader)
	tmpHeader := http.Header{}

	hs := strings.Split(string(header.Header()), "\r\n")
	for i, h := range hs {
		if i == 0 {
			continue
		}
		values := strings.Split(h, ":")
		vLen := len(values)
		if vLen < 2 {
			if values[0] != "" {
				tmpHeader.Set(values[0], "")
			}
		} else {
			tmpHeader.Set(values[0], values[1])
		}
	}
	r.Header = &Header{header: tmpHeader}
	r.StatusHandler = NewStatusHandler()
	r.SetStatus(statusCode, strconv.Itoa(statusCode))

	r.BodyHandler = NewBodyHandler(body)
	return r
}

package http

import (
	"github.com/eolinker/eosc/log"
	"net/http"
	"net/textproto"
)


type LocationReader string

func (l LocationReader) Read(sources interface{}) (string, bool) {
	 if req,ok:=sources.(*http.Request);ok{
	 	return l.read(req)
	 }
	 return "",false
}

func (l LocationReader) read(req *http.Request) (string, bool) {
	return req.URL.Path,true
}

type HeaderReader string
func (h HeaderReader) Read(sources interface{}) (string, bool) {
	if req,ok:=sources.(*http.Request);ok{
		return h.read(req)
	}
	return "",false
}

func (h HeaderReader) read(req *http.Request) (string, bool) {
	v := req.Header[textproto.CanonicalMIMEHeaderKey(string(h))]
 	if len(v) == 0{
 		return  "",false
	}
	return v[0],true
}

type CookieReader string
func (c CookieReader) Read(sources interface{}) (string, bool) {
	if req,ok:=sources.(*http.Request);ok{
		return c.read(req)
	}
	return "",false
}
func (c CookieReader) read(req *http.Request) (string, bool) {
	cookie,err:=req.Cookie(string(c))
	if err!= nil{
		log.Debugf("read cookie %s:%s", c,err)
		return "",false
	}

	return cookie.Value,true
}

type HostReader string
func (h HostReader) Read(sources interface{}) (string, bool) {
	if req,ok:=sources.(*http.Request);ok{
		return h.read(req)
	}
	return "",false
}
func (h HostReader) read(req *http.Request) (string, bool) {
	return req.Host,true
}

type QueryReader string
func (q QueryReader) Read(sources interface{}) (string, bool) {
	if req,ok:=sources.(*http.Request);ok{
		return q.read(req)
	}
	return "",false
}
func (q QueryReader) read(req *http.Request) (string, bool) {

	vs := req.URL.Query()[string(q)]
	if len(vs) == 0 {
		return "",false
	}
	return vs[0],true
}



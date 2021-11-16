package http_context

//type Value map[string]string
//
//func (h Value) Get(key string) (string, bool) {
//	v, ok := h[key]
//	return v, ok
//}
//
//type IRequest interface {
//	Host() string
//	Method() string
//	Path() string
//	ContentType() string
//	RequestHeader() Value
//	Query() Value
//	RawQuery() string
//	RawBody() []byte
//}
//
//type ProxyRequest struct {
//	req         *fasthttp.ProxyRequest
//	path        string
//	host        string
//	method      string
//	header      Value
//	query       Value
//	rawQuery    string
//	rawBody     []byte
//	contentType string
//}
//
//func (r *ProxyRequest) Host() string {
//	if r.host == "" {
//		r.host = strings.Split(string(r.req.RequestHeader.Host()), ":")[0]
//	}
//	return r.host
//}
//
//func (r *ProxyRequest) Method() string {
//	if r.method == "" {
//		r.method = string(r.req.RequestHeader.Method())
//	}
//	return r.method
//}
//
//func (r *ProxyRequest) Path() string {
//	if r.path == "" {
//		r.path = string(r.req.URIRequest().Path())
//	}
//	return r.path
//}
//
//func (r *ProxyRequest) RequestHeader() Value {
//	if r.header == nil {
//		r.header = make(Value)
//		hs := strings.Split(r.req.RequestHeader.String(), "\r\n")
//		for _, h := range hs {
//			vs := strings.Split(h, ":")
//			if len(vs) < 2 {
//				if vs[0] == "" {
//					continue
//				}
//				r.header[vs[0]] = ""
//				continue
//			}
//			r.header[vs[0]] = strings.TrimSpace(vs[1])
//
//		}
//	}
//	return r.header
//}
//
//func (r *ProxyRequest) Query() Value {
//	if r.rawQuery == "" {
//		r.rawQuery = string(r.req.URIRequest().QueryString())
//	}
//	if r.query == nil {
//		r.query = make(Value)
//		qs := strings.Split(r.rawQuery, "&")
//		for _, q := range qs {
//			vs := strings.Split(q, "=")
//			if len(vs) < 2 {
//				if vs[0] == "" {
//					continue
//				}
//				r.query[vs[0]] = ""
//				continue
//			}
//			r.query[vs[0]] = strings.TrimSpace(vs[1])
//		}
//	}
//	return r.query
//}
//
//func (r *ProxyRequest) RawQuery() string {
//	if r.rawQuery == "" {
//		r.rawQuery = string(r.req.URIRequest().QueryString())
//	}
//	return r.rawQuery
//}
//
//func (r *ProxyRequest) RawBody() []byte {
//	if r.rawBody == nil {
//		r.rawBody = r.req.Body()
//	}
//	return r.rawBody
//}
//
//func (r *ProxyRequest) ContentType() string {
//	if r.contentType == "" {
//		r.contentType = string(r.req.RequestHeader.ContentType())
//	}
//	return r.contentType
//}
//
//func newRequest(req *fasthttp.ProxyRequest) IRequest {
//	newReq := &ProxyRequest{
//		req: req,
//	}
//	return newReq
//}

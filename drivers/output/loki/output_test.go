package loki

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type Context struct {
	ctx    context.Context
	labels map[string]string
}

func NewContext(labels map[string]string) *Context {
	return &Context{labels: labels, ctx: context.Background()}
}

func (c *Context) RequestId() string {
	return uuid.NewString()
}

func (c *Context) AcceptTime() time.Time {
	return time.Now()
}

func (c *Context) Context() context.Context {
	return c.ctx
}

func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *Context) WithValue(key, val interface{}) {
	c.ctx = context.WithValue(c.Context(), key, val)
}

func (c *Context) Scheme() string {
	return "http"
}

func (c *Context) Assert(i interface{}) error {
	return nil
}

func (c *Context) SetLabel(name, value string) {
	c.labels[name] = value
}

func (c *Context) GetLabel(name string) string {
	return c.labels[name]
}

func (c *Context) Labels() map[string]string {
	return c.labels
}

func (c *Context) GetComplete() eocontext.CompleteHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetCompleteHandler(handler eocontext.CompleteHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetFinish() eocontext.FinishHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetFinish(handler eocontext.FinishHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetBalance() eocontext.BalanceHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetBalance(handler eocontext.BalanceHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) RealIP() string {
	return "127.0.0.1"
}

func (c *Context) LocalIP() net.IP {
	return net.ParseIP("127.0.0.1")
}

func (c *Context) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 80,
	}
}

func (c *Context) LocalPort() int {
	return 80
}

func (c *Context) IsCloneable() bool {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Clone() (eocontext.EoContext, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Request() http_service.IRequestReader {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Proxy() http_service.IRequest {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Response() http_service.IResponse {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SendTo(scheme string, node eocontext.INode, timeout time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Proxies() []http_service.IProxy {
	//TODO implement me
	panic("implement me")
}

func (c *Context) FastFinish() {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetEntry() eosc.IEntry {
	//TODO implement me
	panic("implement me")
}

//func TestOutput(t *testing.T) {
//	ctx := NewContext(map[string]string{
//		"api":            "7c50261c-5e41-5446-4f72-9e89fcc0c04b",
//		"block_name":     "",
//		"cluster":        "apinto",
//		"node":           "apinto-1",
//		"proxy_addr":     "127.0.0.1:28080",
//		"proxy_header":   "Accept=application%2Fjson%2C+text%2Fplain%2C+%2A%2F%2A&Accept-Encoding=gzip%2C+deflate%2C+br%2C+zstd&Accept-Language=zh-CN%2Czh%3Bq%3D0.9%2Cen%3Bq%3D0.8&Connection=keep-alive&Cookie=uid%3D1%3B+Session%3Dfa2102f5-332f-6b4e-cf29-f2cd9945d4a4%3B+i18next%3Dzh-CN%3B+namespace%3Ddefault%3B+SESSIONID%3D0ed12865-9c6d-1522-3b4d-32e57cf027f5%3B+sid%3DFe26.2%2A%2Ae08199ce07f0eeb1e2b1788e7f3d3d9435bc27405ad8fcca669cab9267256f95%2AsdRhZ2ijCcJ7AaoYlwVRpg%2AbUcOwtHIB2Wtss9U9NL2euGfYcsVJ3hxS1mjUqafO1dwYQZ04OlZlg1hQMozfqWT3jn7ogW0DM3I2LMNUXy8QOyLLZ_xeUW7e7W7FMPyUeQfEGBYmHbd7Zw-cLu6F4TAIjAtfw-Uw_5sA-Ujp4eRL0VENeMKdvQidHnhMTwvInuzXnorDvPiDFfdtT2Eq7Kz6Wj5Y-I6Zr6Yle9Fh0fJIszFbR92EqXHHF-QLQeJe3H4eDnStQv4uSTA3F0sc3zS%2A%2Ad0cab7160d6c4a94d06a40fef1659fae17236afa1fe1cbcf76a1c7bead54319e%2AMDHDJRSWuDAzhv7azz_Dbkxy7t4XM5foqHMWgiMxd4A&Host=127.0.0.1%3A28080&Referer=http%3A%2F%2F127.0.0.1%3A8099%2Fapplication&Sec-Ch-Ua=%22Google+Chrome%22%3Bv%3D%22131%22%2C+%22Chromium%22%3Bv%3D%22131%22%2C+%22Not_A+Brand%22%3Bv%3D%2224%22&Sec-Ch-Ua-Mobile=%3F0&Sec-Ch-Ua-Platform=%22macOS%22&Sec-Fetch-Dest=empty&Sec-Fetch-Mode=cors&Sec-Fetch-Site=same-origin&User-Agent=Mozilla%2F5.0+%28Macintosh%3B+Intel+Mac+OS+X+10_15_7%29+AppleWebKit%2F537.36+%28KHTML%2C+like+Gecko%29+Chrome%2F131.0.0.0+Safari%2F537.36&X-Forwarded-For=127.0.0.1&X-Real-Ip=127.0.0.1",
//		"proxy_host":     "127.0.0.1:28080",
//		"proxy_method":   "GET",
//		"proxy_scheme":   "http",
//		"proxy_status":   "200",
//		"proxy_uri":      "/_system/activation/check?namespace=default",
//		"remote_addr":    "127.0.0.1",
//		"request_header": "Accept=application%2Fjson%2C+text%2Fplain%2C+%2A%2F%2A&Accept-Encoding=gzip%2C+deflate%2C+br%2C+zstd&Accept-Language=zh-CN%2Czh%3Bq%3D0.9%2Cen%3Bq%3D0.8&Connection=keep-alive&Cookie=uid%3D1%3B+Session%3Dfa2102f5-332f-6b4e-cf29-f2cd9945d4a4%3B+i18next%3Dzh-CN%3B+namespace%3Ddefault%3B+SESSIONID%3D0ed12865-9c6d-1522-3b4d-32e57cf027f5%3B+sid%3DFe26.2%2A%2Ae08199ce07f0eeb1e2b1788e7f3d3d9435bc27405ad8fcca669cab9267256f95%2AsdRhZ2ijCcJ7AaoYlwVRpg%2AbUcOwtHIB2Wtss9U9NL2euGfYcsVJ3hxS1mjUqafO1dwYQZ04OlZlg1hQMozfqWT3jn7ogW0DM3I2LMNUXy8QOyLLZ_xeUW7e7W7FMPyUeQfEGBYmHbd7Zw-cLu6F4TAIjAtfw-Uw_5sA-Ujp4eRL0VENeMKdvQidHnhMTwvInuzXnorDvPiDFfdtT2Eq7Kz6Wj5Y-I6Zr6Yle9Fh0fJIszFbR92EqXHHF-QLQeJe3H4eDnStQv4uSTA3F0sc3zS%2A%2Ad0cab7160d6c4a94d06a40fef1659fae17236afa1fe1cbcf76a1c7bead54319e%2AMDHDJRSWuDAzhv7azz_Dbkxy7t4XM5foqHMWgiMxd4A&Host=127.0.0.1%3A8099&Referer=http%3A%2F%2F127.0.0.1%3A8099%2Fapplication&Sec-Ch-Ua=%22Google+Chrome%22%3Bv%3D%22131%22%2C+%22Chromium%22%3Bv%3D%22131%22%2C+%22Not_A+Brand%22%3Bv%3D%2224%22&Sec-Ch-Ua-Mobile=%3F0&Sec-Ch-Ua-Platform=%22macOS%22&Sec-Fetch-Dest=empty&Sec-Fetch-Mode=cors&Sec-Fetch-Site=same-origin&User-Agent=Mozilla%2F5.0+%28Macintosh%3B+Intel+Mac+OS+X+10_15_7%29+AppleWebKit%2F537.36+%28KHTML%2C+like+Gecko%29+Chrome%2F131.0.0.0+Safari%2F537.36",
//		"request_host":   "127.0.0.1:8099",
//		"request_id":     "f3fa9deb-ee0b-4c40-ad32-70939defaea0",
//		"request_method": "GET",
//		"request_scheme": "http",
//		"request_time":   "25",
//		"request_uri":    "/_system/activation/check?namespace=default",
//		"response_time":  "25190416",
//		"service":        "local_dashboard",
//		"src_ip":         "127.0.0.1",
//		"status":         "200",
//	})
//}

package filter

import (
	"context"
	http2 "net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/eolinker/eosc/http"
)

type Out struct {
	output []string
}

func NewOut() *Out {
	return &Out{}
}

func (o *Out) String() string {
	return strings.Join(o.output, ",")
}

func (o *Out) reset() {
	o.output = o.output[:0]
}
func (o *Out) out(v string) {
	o.output = append(o.output, v)
}

type TestFilter struct {
	o    *Out
	name string
}

func NewTestFilter(o *Out, name string) *TestFilter {
	return &TestFilter{o: o, name: name}
}

func (t *TestFilter) DoFilter(ctx http.IHttpContext, next http.IChain) (err error) {
	t.o.out(t.name)

	return next.DoChain(ctx)
}

var out = NewOut()

func TestIFilter(t *testing.T) {

	filterOrg := make([]http.IFilter, 2)
	for i := range filterOrg {
		filterOrg[i] = NewTestFilter(out, strconv.Itoa(i+1))
	}

	chainOrg := NewChainHandler(filterOrg)

	do(chainOrg, "org", t)
	chainTest := chainOrg.Append(NewTestFilter(out, "append1.1"), NewTestFilter(out, "append1.2"))
	do(chainTest, "append", t)
	//chainOrg.Reset([]http.IFilter{NewTestFilter(out, "reset")})
	////do(chainTest, "test", t)
	//
	////do(chainOrg, "org", t)
	//
	chainTest2 := chainOrg.Append(NewTestFilter(out, "append 2"))
	do(chainTest2, "append2", t)
	//do(chainTest, "test1", t)
	t2 := chainTest.Append(NewTestFilter(out, "append3"))
	do(t2, "append to append", t)

	chainOrg.Reset(NewTestFilter(out, "reset"))
	do(t2, "append to append after reset:", t)
}
func do(chain IChain, name string, t *testing.T) {
	out.reset()
	chain.DoChain(new(TestContext))
	t.Log(name, ":", out)
}

type TestContext struct {
	ctx context.Context
}

func (t *TestContext) Context() context.Context {
	if t.ctx == nil {
		t.ctx = context.Background()
	}
	return t.ctx
}

func (t *TestContext) Value(key interface{}) interface{} {
	return t.Context().Value(key)
}

func (t *TestContext) WithValue(key, val interface{}) {
	t.ctx = context.WithValue(t.Context(), key, val)
}

func (t *TestContext) GetHeader(name string) string {
	panic("implement me")
}

func (t *TestContext) Headers() http2.Header {
	panic("implement me")
}

func (t *TestContext) SetHeader(key, value string) {
	panic("implement me")
}

func (t *TestContext) AddHeader(key, value string) {
	panic("implement me")
}

func (t *TestContext) DelHeader(key string) {
	panic("implement me")
}

func (t *TestContext) Set() http.Header {
	panic("implement me")
}

func (t *TestContext) Append() http.Header {
	panic("implement me")
}

func (t *TestContext) Cookie(name string) (*http2.Cookie, error) {
	panic("implement me")
}

func (t *TestContext) Cookies() []*http2.Cookie {
	panic("implement me")
}

func (t *TestContext) AddCookie(c *http2.Cookie) {
	panic("implement me")
}

func (t *TestContext) StatusCode() int {
	panic("implement me")
}

func (t *TestContext) Status() string {
	panic("implement me")
}

func (t *TestContext) SetStatus(code int, status string) {
	panic("implement me")
}

func (t *TestContext) SetBody(bytes []byte) {
	panic("implement me")
}

func (t *TestContext) GetBody() []byte {
	panic("implement me")
}

func (t *TestContext) RequestId() string {
	panic("implement me")
}

func (t *TestContext) Request() http.RequestReader {
	panic("implement me")
}

func (t *TestContext) Proxy() http.Request {
	panic("implement me")
}

func (t *TestContext) Labels() map[string]string {
	panic("implement me")
}

func (t *TestContext) ProxyResponse() http.ResponseReader {
	panic("implement me")
}

func (t *TestContext) SetStoreValue(key string, value interface{}) error {
	panic("implement me")
}

func (t *TestContext) GetStoreValue(key string) (interface{}, bool) {
	panic("implement me")
}

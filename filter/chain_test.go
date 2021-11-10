package filter

import (
	"context"
	"fmt"
	http2 "net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc/http"
)

type Out struct {
	output []string
	t      *testing.T
}

func NewOut(t *testing.T) *Out {
	return &Out{t: t}
}
func (o *Out) Test(chain http.IChain, name string) {
	o.reset()
	if err := chain.DoChain(new(TestContext)); err != nil {
		log.Error(err)
	}

	o.t.Log(name, "\t:\t", o.String())
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

func TestIFilter(t *testing.T) {
	out := NewOut(t)
	filterOrg := make([]http.IFilter, 2)
	for i := range filterOrg {
		filterOrg[i] = NewTestFilter(out, strconv.Itoa(i+1))
	}

	chainOrg := NewChain(filterOrg)

	out.Test(chainOrg, "ort")
	chainTest := chainOrg.Append(NewTestFilter(out, "append1.1"), NewTestFilter(out, "append1.2"))
	out.Test(chainTest, "append")

	chainTest2 := chainOrg.Append(NewTestFilter(out, "append 2"))
	out.Test(chainTest2, "append2")

	t2 := chainTest.Append(NewTestFilter(out, "append3"))
	out.Test(t2, "append to append")

	chainOrg.Reset(NewTestFilter(out, "reset"))
	out.Test(t2, "append to append after reset:")
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

func TestAppendCell(t *testing.T) {
	out := NewOut(t)
	filters1 := make([]http.IFilter, 5)
	for i := range filters1 {

		filters1[i] = NewTestFilter(out, fmt.Sprint("org-", i+1))
	}
	filters2 := make([]http.IFilter, 5)
	for i := range filters2 {

		filters2[i] = NewTestFilter(out, fmt.Sprint("append-", i+1))
	}

	org := NewChain(filters1)
	app := NewChain(filters2)

	target := org.Append(app.ToFilter())

	out.Test(org, "org")
	out.Test(app, "app")
	out.Test(target, "target")

	org.Reset(NewTestFilter(out, "org"))
	out.Test(target, "target-reset org")

	app.Reset(NewTestFilter(out, "app"))
	out.Test(target, "target-reset app")

}

func TestReset(t *testing.T) {
	out := NewOut(t)
	filters1 := make([]http.IFilter, 5)
	for i := range filters1 {
		filters1[i] = NewTestFilter(out, fmt.Sprint("org-", i+1))
	}
	filters2 := make([]http.IFilter, 5)
	for i := range filters2 {

		filters2[i] = NewTestFilter(out, fmt.Sprint("append-", i+1))
	}

	org := NewChain(filters1)
	app := NewChain(filters2)

	target1 := org.Append(app.ToFilter())
	target2 := app.Append(org.ToFilter())

	out.Test(target1, "target1")
	out.Test(target2, "target2")

}

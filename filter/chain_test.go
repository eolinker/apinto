package filter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/context/http-context"
)

type Out struct {
	output []string
	t      *testing.T
}

func NewOut(t *testing.T) *Out {
	return &Out{t: t}
}
func (o *Out) Test(chain context.IChain, name string) {
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

func (t *TestFilter) DoHttpFilter(ctx http_service.IHttpContext, next context.IChain) error {
	t.o.out(t.name)

	return next.DoChain(ctx)
}

func (t *TestFilter) Destroy() {
	return
}

func TestIFilter(t *testing.T) {
	out := NewOut(t)
	filterOrg := make([]http_service.HttpFilter, 2)
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

func (t *TestContext) Response() http_service.IResponse {
	panic("implement me")
}

func (t *TestContext) Proxies() []http_service.IRequest {
	panic("implement me")
}

func (t *TestContext) SendTo(address string, timeout time.Duration) error {
	panic("implement me")
}

func (t *TestContext) RequestId() string {
	panic("implement me")
}

func (t *TestContext) Request() http_service.IRequestReader {
	panic("implement me")
}

func (t *TestContext) Proxy() http_service.IRequest {
	panic("implement me")
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

func TestReset(t *testing.T) {
	out := NewOut(t)
	filters1 := make([]http_service.HttpFilter, 5)
	for i := range filters1 {
		filters1[i] = NewTestFilter(out, fmt.Sprint("org-", i+1))
	}
	filters2 := make([]http_service.HttpFilter, 5)
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

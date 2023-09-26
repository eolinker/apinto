package limiting_strategy

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/eolinker/apinto/strategy"

	"github.com/eolinker/eosc/eocontext"
)

var maxID = 1000

type EmptyContext struct {
	labels map[string]string
}

func NewEmptyContext() *EmptyContext {
	e := &EmptyContext{
		labels: map[string]string{
			//"api": strconv.Itoa(rand.Intn(maxID)),
			"api": strconv.Itoa(1),
		},
	}
	return e
}

func (e *EmptyContext) RequestId() string {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) AcceptTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) Value(key interface{}) interface{} {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) WithValue(key, val interface{}) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) Scheme() string {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) Assert(i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) SetLabel(name, value string) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) GetLabel(name string) string {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) Labels() map[string]string {
	return e.labels
}

func (e *EmptyContext) GetComplete() eocontext.CompleteHandler {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) SetCompleteHandler(handler eocontext.CompleteHandler) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) GetFinish() eocontext.FinishHandler {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) SetFinish(handler eocontext.FinishHandler) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) GetBalance() eocontext.BalanceHandler {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) SetBalance(handler eocontext.BalanceHandler) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) RealIP() string {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) LocalIP() net.IP {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) LocalPort() int {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) IsCloneable() bool {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyContext) Clone() (eocontext.EoContext, error) {
	//TODO implement me
	panic("implement me")
}

func BenchmarkLimiting(b *testing.B) {
	handlers := make([]*LimitingHandler, 0, maxID)
	for i := 0; i < maxID; i++ {
		name := strconv.Itoa(i + 1)
		handler, _ := NewLimitingHandler(name, &Config{
			Stop:     false,
			Priority: 0,
			Filters: strategy.FilterConfig{
				"api": []string{name},
			},
		})
		handlers = append(handlers, handler)
	}
	b.ResetTimer()
	//begin := time.Now()
	for i := 0; i < b.N; i++ {
		ctx := NewEmptyContext()
		//begin := time.Now()
		for _, h := range handlers {
			if h.Filter().Check(ctx) {
				//fmt.Printf("match %s\n", h.name)
				break
			}
		}
		//fmt.Println("spend time:", time.Now().Sub(begin))
	}

}

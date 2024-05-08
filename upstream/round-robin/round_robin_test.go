package round_robin

import (
	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/eosc/eocontext"
	"runtime"
	"sync"
	"testing"
	"time"
)

type demoNode struct {
}

func (d *demoNode) GetAttrs() eocontext.Attrs {
	return make(eocontext.Attrs)
}

func (d *demoNode) GetAttrByName(name string) (string, bool) {
	return "", false
}

func (d *demoNode) ID() string {
	return "127.0.0.1:8080"
}

func (d *demoNode) IP() string {
	return "127.0.0.1"
}

func (d *demoNode) Port() int {
	return 8080
}

func (d *demoNode) Addr() string {
	return "127.0.0.1:8080"
}

func (d *demoNode) Status() eocontext.NodeStatus {
	return eocontext.Running
}

func (d *demoNode) Up() {
}

func (d *demoNode) Down() {
}

func (d *demoNode) Leave() {
}

type demo struct {
	nodeSing demoNode
}

func (d *demo) Nodes() []eocontext.INode {
	return []eocontext.INode{&d.nodeSing}
}

func Test_roundRobin_Next_Retry_demo(t *testing.T) {
	app := new(demo)
	robin := newRoundRobin(app, "http", time.Second)
	testDoRetry(robin, t)

}
func testDoRetry(robin *roundRobin, t *testing.T) {

	timer := time.NewTimer(time.Second * 60)
	for {
		select {
		case <-timer.C:
			return
		default:

		}
		node, _, err := robin.Select(nil)
		if err != nil {
			t.Error(err)
			return
		}

		//t.Log(i, next.Addr())
		node.Down()
	}
}
func Test_roundRobin_Next_Retry_Status(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	discovery := static.CreateAnonymous(&static.Config{
		HealthOn: false,
		Health:   nil,
	})

	app, err := discovery.GetApp("127.0.0.1:8080;127.0.0.1:8081")
	if err != nil {
		return
	}

	robin := newRoundRobin(app, "http", time.Second)
	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			testDoRetry(robin, t)
		}()

	}
	wg.Wait()
}

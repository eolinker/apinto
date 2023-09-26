package counter

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eolinker/eosc"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers/counter"

	"github.com/eolinker/eosc/log"
)

const (
	localMode = "local"
)

var _ counter.ICounter = (*LocalCounter)(nil)

func NewLocalCounter(key string, variables eosc.Untyped[string, string], client scope_manager.IProxyOutput[counter.IClient], counterPusher scope_manager.IProxyOutput[counter.ICountPusher]) *LocalCounter {
	return &LocalCounter{key: key, client: client, variables: variables, counterPusher: counterPusher}
}

// LocalCounter 本地计数器
type LocalCounter struct {
	key string
	// 剩余次数
	remain int64
	// 锁定次数
	lock int64

	locker sync.Mutex

	counterPusher scope_manager.IProxyOutput[counter.ICountPusher]
	variables     eosc.Untyped[string, string]

	resetTime time.Time

	client scope_manager.IProxyOutput[counter.IClient]
}

func (c *LocalCounter) Lock(count int64) error {

	remain := atomic.AddInt64(&c.remain, -count)
	if remain < 0 {
		c.locker.Lock()
		defer c.locker.Unlock()
		now := time.Now()
		if now.Sub(c.resetTime) < 10*time.Second {
			return fmt.Errorf("no enough, key:%s, remain:%d, count:%d", c.key, c.remain, count)
		}

		var err error
		c.resetTime = now
		variables := c.variables.All()
		for _, client := range c.client.List() {
			// 获取最新的次数
			remain, err = counter.GetRemainCount(client, c.key, count, variables)
			if err != nil {
				log.Errorf("get remain count error: %s", err.Error())
				continue
			}
			break
		}
		if err != nil {
			return fmt.Errorf("no enough, key:%s, remain:%d, count:%d", c.key, c.remain, count)
		}
	}
	atomic.StoreInt64(&c.remain, remain)
	atomic.AddInt64(&c.lock, count)
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "lock", "remain:", c.remain, ",lock:", c.lock, ",count:", count)
	log.DebugF("lock now: %s,key: %s,remain: %d,lock: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), c.key, c.remain, c.lock, count)
	return nil
}

func (c *LocalCounter) Complete(count int64) error {
	// 需要解除已经锁定的部分次数
	atomic.AddInt64(&c.lock, -count)
	variables := c.variables.All()
	for _, p := range c.counterPusher.List() {
		p.Push(c.key, count, variables)
	}
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "complete", "remain:", c.remain, ",lock:", c.lock, ",count:", count)
	log.DebugF("complete now: %s,key: %s,remain: %d,lock: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), c.key, c.remain, c.lock, count)
	return nil
}

func (c *LocalCounter) RollBack(count int64) error {
	// 需要解除已经锁定的部分次数,并且增加剩余次数
	atomic.AddInt64(&c.remain, count)
	atomic.AddInt64(&c.lock, -count)
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "rollback", "remain:", c.remain, ",lock:", c.lock, ",count:", count)
	log.DebugF("rollback now: %s,key: %s,remain: %d,lock: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), c.key, c.remain, c.lock, count)
	return nil
}

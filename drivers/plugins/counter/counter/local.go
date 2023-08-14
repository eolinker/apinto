package counter

import (
	"fmt"
	"sync"
	"time"
)

var _ ICounter = (*LocalCounter)(nil)

func NewLocalCounter(key string, client IClient) *LocalCounter {
	return &LocalCounter{key: key, client: client}
}

// LocalCounter 本地计数器
type LocalCounter struct {
	key string
	// 剩余次数
	remain int64
	// 锁定次数
	lock int64

	locker sync.Mutex

	resetTime time.Time
	client    IClient
}

func (c *LocalCounter) Lock(count int64) error {
	c.locker.Lock()
	defer c.locker.Unlock()
	remain := c.remain - count
	if remain < 0 {
		now := time.Now()
		if now.Sub(c.resetTime) < 10*time.Second {
			return fmt.Errorf("no enough, key:%s, remain:%d, count:%d", c.key, c.remain, count)
		}

		var err error
		c.resetTime = now
		// 获取最新的次数
		remain, err = getRemainCount(c.client, c.key, count)
		if err != nil {
			return err
		}
	}
	c.remain = remain
	c.lock += count
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "lock", "remain:", c.remain, ",lock:", c.lock, ",count:", count)
	return nil
}

func (c *LocalCounter) Complete(count int64) error {
	c.locker.Lock()
	defer c.locker.Unlock()
	// 需要解除已经锁定的部分次数
	c.lock -= count
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "complete", "remain:", c.remain, ",lock:", c.lock, ",count:", count)
	return nil
}

func (c *LocalCounter) RollBack(count int64) error {
	c.locker.Lock()
	defer c.locker.Unlock()
	// 需要解除已经锁定的部分次数,并且增加剩余次数
	c.remain += c.lock
	c.lock -= count
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "rollback", "remain:", c.remain, ",lock:", c.lock, ",count:", count)
	return nil
}

func (c *LocalCounter) ResetClient(client IClient) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.client = client
}

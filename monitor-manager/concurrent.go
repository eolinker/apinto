package monitor_manager

import "sync/atomic"

type concurrency struct {
	count int32
}

func (c *concurrency) Add(count int32) {
	atomic.AddInt32(&c.count, count)
}

func (c *concurrency) Get() int32 {
	return atomic.LoadInt32(&c.count)
}

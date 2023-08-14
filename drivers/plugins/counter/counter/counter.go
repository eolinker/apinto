package counter

import "fmt"

type ICounter interface {
	// Lock 锁定次数
	Lock(count int64) error
	// Complete 完成扣次操作
	Complete(count int64) error
	// RollBack 回滚
	RollBack(count int64) error
	// ResetClient 重置客户端
	ResetClient(client IClient)
}

func getRemainCount(client IClient, key string, count int64) (int64, error) {
	remain, err := client.Get(key)
	if err != nil {
		return 0, err
	}
	remain -= count
	if remain < 0 {
		return 0, fmt.Errorf("no enough, key:%s, remain:%d, count:%d", key, remain, count)
	}
	return remain, nil
}

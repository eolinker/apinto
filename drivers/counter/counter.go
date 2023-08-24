package counter

import (
	"fmt"
	"reflect"

	"github.com/eolinker/eosc"

	"github.com/eolinker/eosc/utils/config"
)

var (
	FilterSkillName = config.TypeName(reflect.TypeOf((*IClient)(nil)).Elem())
)

type IClient interface {
	Get(variables eosc.Untyped[string, string]) (int64, error)
}

type ICounter interface {
	// Lock 锁定次数
	Lock(count int64) error
	// Complete 完成扣次操作
	Complete(count int64) error
	// RollBack 回滚
	RollBack(count int64) error
}

func GetRemainCount(client IClient, key string, count int64, variables eosc.Untyped[string, string]) (int64, error) {
	remain, err := client.Get(variables)
	if err != nil {
		return 0, err
	}
	remain -= count
	if remain < 0 {
		return 0, fmt.Errorf("no enough, key:%s, remain:%d, count:%d", key, remain, count)
	}
	return remain, nil
}

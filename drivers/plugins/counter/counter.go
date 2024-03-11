package counter

type ICounter interface {
	// Lock 锁定次数
	Lock(count int64) error
	// Complete 完成扣次操作
	Complete(count int64) error
	// RollBack 回滚
	RollBack(count int64) error
}

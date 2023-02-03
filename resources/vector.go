package resources

import "time"

// 通过滑动窗口实现的平滑计数器
const VectorsSkill = "github.com/eolinker/apinto/resources.resources.IVectors"

type IVectors interface {
	BuildVector(name string, uni, step time.Duration) (Vector, error)
}

type Vector interface {
	Add(key string, delta int64)
	CompareAndAdd(key string, threshold, delta int64) bool
	Get(key string) int64
}

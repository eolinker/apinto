package resources

import (
	"context"
	"time"
)

// 通过滑动窗口实现的平滑计数器
const VectorsSkill = "github.com/eolinker/apinto/resources.resources.IVectors"

type IVectors interface {
	BuildVector(name string, uni, step time.Duration) (Vector, error)
}

type Vector interface {
	Add(ctx context.Context, key string, delta int64) int64
	CompareAndAdd(ctx context.Context, key string, threshold, delta int64) (int64, bool)
	Get(ctx context.Context, key string) int64
}

package scalar

import (
	"sync/atomic"
	"time"
)

type Vectors interface {
	CompareAndAdd(threshold, delta uint64) (added bool)
}

type _Vectors struct {
	vectors   []uint64
	size      uint64
	lastIndex uint64
	step      uint64
}

//newVectors 统计时长，滑动窗口步进，单位都是 毫秒，
func newVectors(uni, step uint64) *_Vectors {

	if uni < 1000 {
		uni = 1000
	}
	if step < 500 {
		step = 500
	}

	size := uni / step

	if size > 20 {
		size = 20
		step = uni / size
	}

	return &_Vectors{size: size, step: step, vectors: make([]uint64, size)}
}

func (s *_Vectors) CompareAndAdd(threshold, delta uint64) bool {
	index := s.refresh()
	value := uint64(0)
	for i := range s.vectors {
		value += atomic.LoadUint64(&s.vectors[i])
	}

	if 0 == threshold || value < threshold {
		atomic.AddUint64(&s.vectors[index%s.size], delta)
		return true
	}
	return false
}

func (s *_Vectors) refresh() uint64 {
	seconds := uint64(time.Now().Unix())
	index := seconds / s.step
	last := atomic.SwapUint64(&s.lastIndex, index)

	if index > last {
		if index-last > s.step {
			last = index - s.step - 1
		}
		for i := index; i > last; i-- {
			atomic.StoreUint64(&s.vectors[i%s.size], 0)
		}
	}
	return index
}

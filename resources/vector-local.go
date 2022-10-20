package resources

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	localVector IVectors = (*VectorsLocalBuild)(nil)
)
var (
	onceVector  sync.Once
	LocalVector func() IVectors
)

func init() {
	LocalVector = func() IVectors {

		onceVector.Do(func() {
			localVector = NewVectorsLocalBuild()
			LocalVector = func() IVectors {
				return localVector
			}
		})
		return localVector
	}

}

type VectorsLocalBuild struct {
	lock sync.Mutex

	vectors map[string]Vector
}

func NewVectorsLocalBuild() *VectorsLocalBuild {
	return &VectorsLocalBuild{
		vectors: make(map[string]Vector),
	}
}

func (v *VectorsLocalBuild) BuildVector(name string, uni, step time.Duration) (Vector, error) {

	if uni < time.Second {
		uni = time.Second
	}
	if step < 500*time.Millisecond {
		step = 500 * time.Millisecond
	}

	size := uni / step
	if size > 20 {
		size = 20
	}
	step = uni / size

	key := fmt.Sprintf("%s:%d:%d", name, uni, step)
	v.lock.Lock()
	defer v.lock.Unlock()

	vector, has := v.vectors[key]
	if has {
		return vector, nil
	}

	vector = newVectorLocal(key, uni, step)
	v.vectors[key] = vector
	return vector, nil
}

type vectorLocal struct {
	name      string
	step      int64
	lastIndex int64
	size      int64
	lock      sync.RWMutex
	vm        map[string]*vectorValues
}

func (v *vectorLocal) CompareAndAdd(key string, threshold, delta int64) bool {
	index, vector := v.refresh(key)
	value := v.read(vector)
	if value < threshold {
		atomic.AddInt64(&vector.vectors[index%v.size], delta)
		return true
	}
	return false
}

type vectorValues struct {
	vectors []int64
}

func (v *vectorLocal) Add(key string, delta int64) int64 {
	index, vector := v.refresh(key)
	atomic.AddInt64(&vector.vectors[index%v.size], delta)
	return v.read(vector)
}

func (v *vectorLocal) Get(key string) int64 {
	_, values := v.refresh(key)
	return v.read(values)
}
func (v *vectorLocal) vector(key string) *vectorValues {
	token := fmt.Sprint(v.name, ":", key)
	v.lock.RLock()
	values, has := v.vm[token]
	v.lock.RUnlock()
	if has {
		return values
	}

	v.lock.Lock()
	defer v.lock.Unlock()
	values, has = v.vm[token]
	if has {
		return values
	}
	values = &vectorValues{vectors: make([]int64, v.size)}
	return values
}
func (v *vectorLocal) read(vectors *vectorValues) int64 {

	value := int64(0)
	for i := range vectors.vectors {
		value += atomic.LoadInt64(&vectors.vectors[i])
	}
	return value
}
func (v *vectorLocal) refresh(key string) (int64, *vectorValues) {
	vectors := v.vector(key)
	seconds := time.Now().UnixNano()
	index := seconds / v.step
	last := atomic.SwapInt64(&v.lastIndex, index)

	if index > last {
		if index-last > v.step {

			for i := int64(0); i < v.size; i++ {
				atomic.StoreInt64(&vectors.vectors[i], 0)
			}

		} else {
			for i := last; i < index; i++ {
				atomic.StoreInt64(&vectors.vectors[i%v.size], 0)
			}
		}
	}
	return index, vectors
}
func newVectorLocal(name string, uin, step time.Duration) *vectorLocal {
	v := &vectorLocal{name: name, step: int64(step), size: int64(uin / step), vm: make(map[string]*vectorValues)}

	index := time.Now().UnixNano() / v.step
	atomic.StoreInt64(&v.lastIndex, index)
	return v
}

package queue

type Entry[T any] struct {
	v    *T
	next *Entry[T]
}

func (entry *Entry[T]) Value() *T {
	return entry.v
}
func NewEntry[T any](v *T) *Entry[T] {
	return &Entry[T]{
		v: v,
	}
}

type Queue[T any] interface {
	Pop() *Entry[T]
	Push(entry *Entry[T])
	Empty() bool
}

var (
	_ Queue[any] = (*imlQueue[any])(nil)
)

type imlQueue[T any] struct {
	head *Entry[T]
	tail *Entry[T]
}

func (q *imlQueue[T]) Empty() bool {
	return q.head == nil
}

func NewQueue[T any](vs ...*T) Queue[T] {
	q := &imlQueue[T]{}
	for _, v := range vs {
		q.Push(NewEntry(v))
	}
	return q
}

func (q *imlQueue[T]) Pop() *Entry[T] {
	if q.head == nil {
		return nil
	}
	head := q.head
	next := head.next
	head.next = nil
	q.head = next
	if next == nil {
		q.tail = nil
	}
	return head
}

func (q *imlQueue[T]) Push(entry *Entry[T]) {

	tail := q.tail
	q.tail = entry
	if tail == nil {
		q.head = entry
	} else {
		tail.next = entry
	}

}

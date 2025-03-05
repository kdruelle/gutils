package containers

type Container interface {
	Len() int
	IsEmpty() bool
	Clear()
}

type SequenceContainer[T any] interface {
	Get(int) (T, error)
	GetPointer(int)(*T, error)
	Set(int, T) error
	Insert(int, T) error
	Remove(int) error
}

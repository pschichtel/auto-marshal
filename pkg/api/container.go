package api

type Container[T interface{}] interface {
	ContainedValue() T
}

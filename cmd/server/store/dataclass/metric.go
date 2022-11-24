package dataclass

type Metric[T any] struct {
	Name  string
	Value T
}

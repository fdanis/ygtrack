package dataclass

import "github.com/fdanis/ygtrack/internal/constraints"

type Metric[T constraints.Number] struct {
	Name  string
	Value T
}

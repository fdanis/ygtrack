package constraints

type Number interface {
	~int64 | ~float64
}

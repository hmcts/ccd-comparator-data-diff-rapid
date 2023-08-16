package comparator

type Pair[T any, U any] struct {
	Left  T
	Right U
}

func NewPair[T any, U any](left T, right U) *Pair[T, U] {
	return &Pair[T, U]{
		Left:  left,
		Right: right,
	}
}

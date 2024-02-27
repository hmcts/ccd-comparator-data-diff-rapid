package comparator

type Either[T any, U any] struct {
	Left  T
	Right U
}

func NewEither[T any, U any](left T, right U) Either[T, U] {
	return Either[T, U]{
		Left:  left,
		Right: right,
	}
}

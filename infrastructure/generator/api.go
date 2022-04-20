package generator

type Generator[T comparable] interface {
	Generate() (T, error)
}

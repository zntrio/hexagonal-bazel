package types

func AsRef[T any](v T) *T {
	return &v
}

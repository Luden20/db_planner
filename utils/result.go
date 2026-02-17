package utils

type Result[T any] struct {
	Success bool
	Data    *T
	Message string
}

func ResultFromError[T any](err error) Result[T] {
	return Result[T]{
		Success: false,
		Data:    nil,
		Message: err.Error(),
	}
}

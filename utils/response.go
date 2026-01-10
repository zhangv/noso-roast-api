package utils

type Response map[string]interface{}

type Response2[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func Success[T any](data T) Response2[T] {
	return Response2[T]{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

func Error[T any](code int, message string) Response2[T] {
	return Response2[T]{
		Code:    code,
		Message: message,
	}
}

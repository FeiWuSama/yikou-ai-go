package common

import (
	"errors"
	"workspace-yikou-ai-go/pkg/errors"
)

type BaseResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func NewSuccessResponse[T any](data T) *BaseResponse[T] {
	return &BaseResponse[T]{
		Code:    20000,
		Message: "success",
		Data:    data,
	}
}

func NewErrorResponse[T any](err error) *BaseResponse[T] {
	newError := pkg.ErrorNo{}
	if errors.As(err, &newError) {
		return &BaseResponse[T]{
			Code:    newError.Code,
			Message: newError.Message,
		}
	} else {
		newError = pkg.ConvertError(err)
		return &BaseResponse[T]{
			Code:    newError.Code,
			Message: newError.Message,
		}
	}
}

func NewResponse[T any](code int, message string, data T) *BaseResponse[T] {
	return &BaseResponse[T]{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

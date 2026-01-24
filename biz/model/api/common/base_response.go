package common

import "workspace-yikou-ai-go/biz/model/errors"

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

func NewErrorResponse[T any](err errors.ErrorNo) *BaseResponse[T] {
	return &BaseResponse[T]{
		Code:    err.Code,
		Message: err.Message,
	}
}

func NewResponse[T any](code int, message string, data T) *BaseResponse[T] {
	return &BaseResponse[T]{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

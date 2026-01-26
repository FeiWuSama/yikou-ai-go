package pkg

import (
	"errors"
	"fmt"
)

type ErrorNo struct {
	Code    int
	Message string
}

func (e ErrorNo) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func NewErrorNo(code int, message string) ErrorNo {
	return ErrorNo{
		Code:    code,
		Message: message,
	}
}

func NewErrorNoByCode(code int) ErrorNo {
	message, _ := ErrMessageMap[code]
	return ErrorNo{
		Code:    code,
		Message: message,
	}
}

func (e ErrorNo) WithMessage(message string) ErrorNo {
	e.Message = message
	return e
}

func ConvertError(err error) ErrorNo {
	newErr := ErrorNo{}
	if errors.As(err, &newErr) {
		return newErr
	}

	newErr = SystemError
	newErr.Message = err.Error()
	return newErr
}

const (
	SuccessErrCode        = 0
	ParamErrCode          = 40000
	NotLoginErrCode       = 40100
	NotAuthErrCode        = 40101
	ForbiddenErrorCode    = 40400
	TooManyRequestErrCode = 40300
	SystemErrorCode       = 50000
	OperationErrorCode    = 51001
)

var ErrMessageMap = map[int]string{
	SuccessErrCode:        "ok",
	ParamErrCode:          "请求参数错误",
	NotLoginErrCode:       "未登录",
	NotAuthErrCode:        "无权限",
	ForbiddenErrorCode:    "禁止访问",
	TooManyRequestErrCode: "请求过于频繁",
	SystemErrorCode:       "系统内部异常",
	OperationErrorCode:    "操作失败",
}

var (
	Success        = NewErrorNoByCode(SuccessErrCode)
	ParamsError    = NewErrorNoByCode(ParamErrCode)
	NotLoginError  = NewErrorNoByCode(NotLoginErrCode)
	NotAuthError   = NewErrorNoByCode(NotAuthErrCode)
	ForbiddenError = NewErrorNoByCode(ForbiddenErrorCode)
	InternalError  = NewErrorNoByCode(TooManyRequestErrCode)
	SystemError    = NewErrorNoByCode(SystemErrorCode)
	OperationError = NewErrorNoByCode(OperationErrorCode)
)

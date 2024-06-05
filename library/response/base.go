package response

import (
	"collection-center/internal/ecode"
	"strings"
)

type Result struct {
	Code    ecode.Code  `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Page struct {
	Total      int64 `json:"total"`
	PageNumber int   `json:"pageNumber" binding:"required"`
	PageSize   int   `json:"pageSize" binding:"required"`
}
type List struct {
	Page
	List interface{} `json:"list"`
}

func (r *Result) IsSuccess() bool {
	if r.Code == ecode.OK {
		return true
	} else {
		return false
	}
}
func Res(err error, data interface{}, language string) *Result {
	if err != nil {
		return ResErr(err, language)
	} else {
		return ResOk(data)
	}
}

func ResOk(data interface{}) *Result {
	res := &Result{
		Code:    ecode.OK,
		Message: "Success",
		Data:    data,
	}
	return res
}

func ResErr(err error, language string) *Result {
	code := ecode.ServerErr
	res := &Result{
		Code:    code,
		Message: err.Error(),
	}
	if ec, ok := err.(ecode.Code); ok {
		res = &Result{
			Code:    ec,
			Message: ec.Message(language),
		}
	} else if strings.HasPrefix(err.Error(), "rpc error: code = Unknown desc = ") {
		//rpc error: code = Unknown desc
		ec = ecode.String(err.Error()[33:])
		msg := ec.Message(language)
		if ec == ecode.ServerErr {
			msg = err.Error()[33:]
		}
		res = &Result{
			Code:    ec,
			Message: msg,
		}
	}
	return res
}

func ResWrapErr(err error, err1 error, language string) *Result {
	code := ecode.ServerErr
	res := &Result{
		Code:    code,
		Message: err.Error(),
	}
	if ec, ok := err.(ecode.Code); ok {
		res = &Result{
			Code:    ec,
			Message: err1.Error(),
		}
	} else if strings.HasPrefix(err.Error(), "rpc error: code = Unknown desc = ") {
		//rpc error: code = Unknown desc
		ec = ecode.String(err.Error()[33:])
		msg := ec.Message(language)
		if ec == ecode.ServerErr {
			msg = err.Error()[33:]
		}
		res = &Result{
			Code:    ec,
			Message: msg,
		}
	}
	return res
}

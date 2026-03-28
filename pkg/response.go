package pkg

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func OK(data any) Response {
	return Response{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	}
}

func Fail(err *BizError) Response {
	if err == nil {
		return Response{
			Code:    CodeInternal,
			Message: "内部错误",
			Data:    struct{}{},
		}
	}

	return Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    struct{}{},
	}
}

func FailWith(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    struct{}{},
	}
}

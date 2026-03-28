package pkg

import "fmt"

const (
	CodeOK            = 0
	CodeInvalidParam  = 10001
	CodeUnauthorized  = 10002
	CodeRateLimited   = 10003
	CodeGoodsNotFound = 20001
	CodeSoldOut       = 20002
	CodeAlreadyBought = 20003
	CodeInvalidToken  = 20004
	CodeOrderNotFound = 20005
	CodeOrderTimeout  = 20006
	CodeInternal      = 50001
)

type BizError struct {
	Code    int
	Message string
}

func (e *BizError) Error() string {
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

func NewBizError(code int, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrInvalidParam  = NewBizError(CodeInvalidParam, "参数错误")
	ErrUnauthorized  = NewBizError(CodeUnauthorized, "未授权")
	ErrRateLimited   = NewBizError(CodeRateLimited, "限流")
	ErrGoodsNotFound = NewBizError(CodeGoodsNotFound, "商品不存在")
	ErrSoldOut       = NewBizError(CodeSoldOut, "已售罄")
	ErrAlreadyBought = NewBizError(CodeAlreadyBought, "已购")
	ErrInvalidToken  = NewBizError(CodeInvalidToken, "Token无效")
	ErrOrderNotFound = NewBizError(CodeOrderNotFound, "订单不存在")
	ErrOrderTimeout  = NewBizError(CodeOrderTimeout, "订单超时")
	ErrInternal      = NewBizError(CodeInternal, "内部错误")
)

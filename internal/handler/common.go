package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"seckill/pkg"
)

func writeOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, pkg.OK(data))
}

func writeError(c *gin.Context, err error) {
	if err == nil {
		c.JSON(http.StatusInternalServerError, pkg.Fail(pkg.ErrInternal))
		return
	}
	var bizErr *pkg.BizError
	if errors.As(err, &bizErr) {
		status := http.StatusOK
		switch bizErr.Code {
		case pkg.CodeUnauthorized:
			status = http.StatusUnauthorized
		case pkg.CodeRateLimited:
			status = http.StatusTooManyRequests
		case pkg.CodeInvalidParam:
			status = http.StatusBadRequest
		}
		c.JSON(status, pkg.Fail(bizErr))
		return
	}
	c.JSON(http.StatusInternalServerError, pkg.Fail(pkg.ErrInternal))
}

func parseInt64Param(c *gin.Context, name string) (int64, error) {
	v := c.Param(name)
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, pkg.ErrInvalidParam
	}
	return id, nil
}

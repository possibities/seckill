package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"seckill/internal/middleware"
	"seckill/internal/model"
	"seckill/pkg"
)

type orderService interface {
	PayOrder(ctx context.Context, userID, orderID int64) (*model.PayOrderResp, error)
}

type OrderHandler struct {
	svc orderService
}

func NewOrderHandler(svc orderService) (*OrderHandler, error) {
	if svc == nil {
		return nil, pkg.ErrInternal
	}
	return &OrderHandler{svc: svc}, nil
}

func (h *OrderHandler) Pay(c *gin.Context) {
	uid, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, pkg.ErrUnauthorized)
		return
	}
	orderID, err := parseInt64Param(c, "orderId")
	if err != nil {
		writeError(c, err)
		return
	}

	resp, err := h.svc.PayOrder(c.Request.Context(), uid, orderID)
	if err != nil {
		writeError(c, err)
		return
	}
	writeOK(c, resp)
}

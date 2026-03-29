package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"seckill/internal/middleware"
	"seckill/internal/model"
	"seckill/pkg"
)

type seckillService interface {
	IssueSeckillToken(ctx context.Context, userID, goodsID int64) (*model.SeckillTokenResp, error)
	DoSeckill(ctx context.Context, userID, goodsID int64, token string) (*model.QueueResp, error)
	GetResult(ctx context.Context, userID, goodsID int64) (*model.SeckillResultResp, error)
}

type SeckillHandler struct {
	svc seckillService
}

func NewSeckillHandler(svc seckillService) (*SeckillHandler, error) {
	if svc == nil {
		return nil, pkg.ErrInternal
	}
	return &SeckillHandler{svc: svc}, nil
}

func (h *SeckillHandler) IssueToken(c *gin.Context) {
	uid, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, pkg.ErrUnauthorized)
		return
	}
	goodsID, err := parseInt64Param(c, "goodsId")
	if err != nil {
		writeError(c, err)
		return
	}

	resp, err := h.svc.IssueSeckillToken(c.Request.Context(), uid, goodsID)
	if err != nil {
		writeError(c, err)
		return
	}
	writeOK(c, resp)
}

func (h *SeckillHandler) Do(c *gin.Context) {
	uid, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, pkg.ErrUnauthorized)
		return
	}
	goodsID, err := parseInt64Param(c, "goodsId")
	if err != nil {
		writeError(c, err)
		return
	}

	urlToken := c.Param("token")
	if urlToken == "" {
		writeError(c, pkg.ErrInvalidParam)
		return
	}

	resp, err := h.svc.DoSeckill(c.Request.Context(), uid, goodsID, urlToken)
	if err != nil {
		writeError(c, err)
		return
	}
	writeOK(c, resp)
}

func (h *SeckillHandler) Result(c *gin.Context) {
	uid, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, pkg.ErrUnauthorized)
		return
	}
	goodsID, err := parseInt64Param(c, "goodsId")
	if err != nil {
		writeError(c, err)
		return
	}

	resp, err := h.svc.GetResult(c.Request.Context(), uid, goodsID)
	if err != nil {
		writeError(c, err)
		return
	}
	writeOK(c, resp)
}

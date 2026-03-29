package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"seckill/internal/model"
	"seckill/pkg"
)

type goodsGetter interface {
	GetByID(ctx context.Context, goodsID int64) (*model.SeckillGoods, error)
}

type GoodsHandler struct {
	goodsRepo goodsGetter
}

func NewGoodsHandler(goodsRepo goodsGetter) (*GoodsHandler, error) {
	if goodsRepo == nil {
		return nil, pkg.ErrInternal
	}
	return &GoodsHandler{goodsRepo: goodsRepo}, nil
}

func (h *GoodsHandler) GetGoods(c *gin.Context) {
	goodsID, err := parseInt64Param(c, "goodsId")
	if err != nil {
		writeError(c, err)
		return
	}

	goods, err := h.goodsRepo.GetByID(c.Request.Context(), goodsID)
	if err != nil {
		writeError(c, err)
		return
	}
	writeOK(c, goods)
}

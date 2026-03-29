package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"seckill/internal/middleware"
)

type RouterDeps struct {
	JWT     *middleware.JWTManager
	User    *UserHandler
	Goods   *GoodsHandler
	Seckill *SeckillHandler
	Order   *OrderHandler
}

func RegisterRoutes(engine *gin.Engine, deps RouterDeps) {
	v1 := engine.Group("/api/v1")

	v1.POST("/user/login", deps.User.Login)
	v1.GET("/goods/:goodsId", deps.Goods.GetGoods)

	authGroup := v1.Group("")
	authGroup.Use(deps.JWT.Middleware())

	tokenLimiter := middleware.NewUserGoodsWindowLimiter(3, time.Minute)
	globalSeckillLimiter := middleware.NewGlobalRateLimiter(rate.Limit(10000), 10000)
	userSeckillLimiter := middleware.NewUserRateLimiter(rate.Limit(1), 1)

	authGroup.GET("/seckill/token/:goodsId", tokenLimiter, deps.Seckill.IssueToken)
	authGroup.POST("/seckill/do/:goodsId/:token", globalSeckillLimiter, userSeckillLimiter, deps.Seckill.Do)
	authGroup.GET("/seckill/result/:goodsId", deps.Seckill.Result)
	authGroup.POST("/order/:orderId/pay", deps.Order.Pay)
}

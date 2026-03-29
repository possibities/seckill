package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"seckill/pkg"
)

type userLimiter struct {
	mu      sync.Mutex
	limit   rate.Limit
	burst   int
	entries map[int64]*rate.Limiter
}

type userGoodsCounter struct {
	count   int
	resetAt time.Time
}

type userGoodsLimiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	entries map[string]*userGoodsCounter
}

func NewGlobalRateLimiter(limit rate.Limit, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(limit, burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, pkg.Fail(pkg.ErrRateLimited))
			return
		}
		c.Next()
	}
}

func NewUserRateLimiter(limit rate.Limit, burst int) gin.HandlerFunc {
	ul := &userLimiter{limit: limit, burst: burst, entries: map[int64]*rate.Limiter{}}
	return func(c *gin.Context) {
		uid, ok := UserIDFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.Fail(pkg.ErrUnauthorized))
			return
		}
		if !ul.allow(uid) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, pkg.Fail(pkg.ErrRateLimited))
			return
		}
		c.Next()
	}
}

func (u *userLimiter) allow(userID int64) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	l, ok := u.entries[userID]
	if !ok {
		l = rate.NewLimiter(u.limit, u.burst)
		u.entries[userID] = l
	}
	return l.Allow()
}

func NewUserGoodsWindowLimiter(max int, window time.Duration) gin.HandlerFunc {
	ugl := &userGoodsLimiter{max: max, window: window, entries: map[string]*userGoodsCounter{}}
	return func(c *gin.Context) {
		uid, ok := UserIDFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.Fail(pkg.ErrUnauthorized))
			return
		}
		goodsID, err := strconv.ParseInt(c.Param("goodsId"), 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, pkg.Fail(pkg.ErrInvalidParam))
			return
		}

		if !ugl.allow(uid, goodsID) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, pkg.Fail(pkg.ErrRateLimited))
			return
		}
		c.Next()
	}
}

func (u *userGoodsLimiter) allow(userID, goodsID int64) bool {
	key := fmt.Sprintf("%d:%d", userID, goodsID)
	now := time.Now()

	u.mu.Lock()
	defer u.mu.Unlock()

	entry, ok := u.entries[key]
	if !ok || now.After(entry.resetAt) {
		u.entries[key] = &userGoodsCounter{count: 1, resetAt: now.Add(u.window)}
		return true
	}
	if entry.count >= u.max {
		return false
	}
	entry.count++
	return true
}

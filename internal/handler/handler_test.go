package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"seckill/config"
	"seckill/internal/middleware"
	"seckill/internal/model"
	"seckill/pkg"
)

type fakeAuth struct{}

func (f *fakeAuth) Authenticate(ctx context.Context, username, password string) (int64, error) {
	if username == "u" && password == "p" {
		return 1, nil
	}
	return 0, pkg.ErrUnauthorized
}

type fakeGoods struct{}

func (f *fakeGoods) GetByID(ctx context.Context, goodsID int64) (*model.SeckillGoods, error) {
	if goodsID == 1 {
		return &model.SeckillGoods{ID: 1, Name: "g"}, nil
	}
	return nil, pkg.ErrGoodsNotFound
}

type fakeSeckill struct{}

func (f *fakeSeckill) IssueSeckillToken(ctx context.Context, userID, goodsID int64) (*model.SeckillTokenResp, error) {
	return &model.SeckillTokenResp{SeckillToken: "t", ExpireAt: time.Now().Unix()}, nil
}
func (f *fakeSeckill) DoSeckill(ctx context.Context, userID, goodsID int64, token string) (*model.QueueResp, error) {
	return &model.QueueResp{QueueStatus: "queuing"}, nil
}
func (f *fakeSeckill) GetResult(ctx context.Context, userID, goodsID int64) (*model.SeckillResultResp, error) {
	return &model.SeckillResultResp{Status: "queuing"}, nil
}

type fakeOrder struct{}

func (f *fakeOrder) PayOrder(ctx context.Context, userID, orderID int64) (*model.PayOrderResp, error) {
	if orderID <= 0 {
		return nil, errors.New("bad")
	}
	return &model.PayOrderResp{OrderID: orderID, Status: "paid", PaidAt: time.Now().Unix()}, nil
}

func testRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	jwt := middleware.NewJWTManager(config.JWTConfig{Secret: "s", Issuer: "i", Expire: time.Hour})
	uh, _ := NewUserHandler(&fakeAuth{}, jwt)
	gh, _ := NewGoodsHandler(&fakeGoods{})
	sh, _ := NewSeckillHandler(&fakeSeckill{})
	oh, _ := NewOrderHandler(&fakeOrder{})
	r := gin.New()
	RegisterRoutes(r, RouterDeps{JWT: jwt, User: uh, Goods: gh, Seckill: sh, Order: oh})
	return r
}

func TestLogin(t *testing.T) {
	r := testRouter(t)
	body, _ := json.Marshal(map[string]string{"username": "u", "password": "p"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected code: %d", w.Code)
	}
}

func TestGetGoods(t *testing.T) {
	r := testRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/goods/1", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected code: %d", w.Code)
	}
}

func TestAuthRequired(t *testing.T) {
	r := testRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/seckill/token/1", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected code: %d", w.Code)
	}
}

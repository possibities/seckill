package mq

import (
	"context"
	"errors"
	"testing"
	"time"

	"seckill/internal/cache"
	"seckill/internal/model"
	"seckill/pkg"
)

type fakeOrderRepo struct {
	getErr       error
	createErr    error
	createdOrder *model.SeckillOrder
}

func (f *fakeOrderRepo) Create(ctx context.Context, order *model.SeckillOrder) error {
	f.createdOrder = order
	return f.createErr
}

func (f *fakeOrderRepo) GetByIdempotentKey(ctx context.Context, key string) (*model.SeckillOrder, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &model.SeckillOrder{ID: 1}, nil
}

type fakeGoodsRepo struct {
	goods       *model.SeckillGoods
	getErr      error
	decreaseErr error
}

func (f *fakeGoodsRepo) GetByID(ctx context.Context, goodsID int64) (*model.SeckillGoods, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.goods, nil
}

func (f *fakeGoodsRepo) DecreaseAvailableStock(ctx context.Context, goodsID int64) error {
	return f.decreaseErr
}

type fakeResultCache struct {
	lastResult string
}

func (f *fakeResultCache) SetResult(ctx context.Context, userID, goodsID int64, result string, ttl time.Duration) error {
	f.lastResult = result
	return nil
}

type fakeIDGen struct{}

func (f *fakeIDGen) NextID() (int64, error) {
	return 123, nil
}

func TestHandleMessageDuplicateConsumed(t *testing.T) {
	c := &Consumer{
		orderRepo:  &fakeOrderRepo{},
		goodsRepo:  &fakeGoodsRepo{},
		stockCache: &fakeResultCache{},
		idGen:      &fakeIDGen{},
		resultTTL:  time.Minute,
	}
	body := []byte(`{"user_id":1,"goods_id":2,"idempotent_key":"k1","created_at":1}`)
	if err := c.handleMessage(context.Background(), body); err != nil {
		t.Fatalf("handle message: %v", err)
	}
}

func TestHandleMessageCreateSuccess(t *testing.T) {
	orderRepo := &fakeOrderRepo{getErr: pkg.ErrOrderNotFound}
	resultCache := &fakeResultCache{}
	c := &Consumer{
		orderRepo:  orderRepo,
		goodsRepo:  &fakeGoodsRepo{goods: &model.SeckillGoods{ID: 2, SeckillPrice: 99.9}},
		stockCache: resultCache,
		idGen:      &fakeIDGen{},
		payTTL:     time.Minute,
		resultTTL:  time.Minute,
	}
	body := []byte(`{"user_id":1,"goods_id":2,"idempotent_key":"k2","created_at":1}`)
	if err := c.handleMessage(context.Background(), body); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	if orderRepo.createdOrder == nil {
		t.Fatal("order should be created")
	}
	if resultCache.lastResult != cache.ResultSuccess {
		t.Fatalf("unexpected result status: %s", resultCache.lastResult)
	}
}

func TestHandleMessageSoldOut(t *testing.T) {
	resultCache := &fakeResultCache{}
	c := &Consumer{
		orderRepo:  &fakeOrderRepo{getErr: pkg.ErrOrderNotFound},
		goodsRepo:  &fakeGoodsRepo{goods: &model.SeckillGoods{ID: 2}, decreaseErr: pkg.ErrSoldOut},
		stockCache: resultCache,
		idGen:      &fakeIDGen{},
		resultTTL:  time.Minute,
	}
	body := []byte(`{"user_id":1,"goods_id":2,"idempotent_key":"k3","created_at":1}`)
	if err := c.handleMessage(context.Background(), body); err != nil {
		t.Fatalf("sold out should not retry: %v", err)
	}
	if resultCache.lastResult != cache.ResultFailed {
		t.Fatalf("expected failed result, got: %s", resultCache.lastResult)
	}
}

func TestHandleMessageUnexpectedError(t *testing.T) {
	expected := errors.New("db down")
	c := &Consumer{
		orderRepo:  &fakeOrderRepo{getErr: expected},
		goodsRepo:  &fakeGoodsRepo{},
		stockCache: &fakeResultCache{},
		idGen:      &fakeIDGen{},
	}
	body := []byte(`{"user_id":1,"goods_id":2,"idempotent_key":"k4","created_at":1}`)
	if err := c.handleMessage(context.Background(), body); !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

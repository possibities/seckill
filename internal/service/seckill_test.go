package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"seckill/internal/cache"
	"seckill/internal/model"
	"seckill/pkg"
)

type fakeGoodsRepo struct {
	goods *model.SeckillGoods
	err   error
}

func (f *fakeGoodsRepo) GetByID(ctx context.Context, goodsID int64) (*model.SeckillGoods, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.goods, nil
}

type fakeStockCache struct {
	decreaseResult cache.StockResult
	decreaseErr    error
	setResultCalls int
	lastResult     string
}

func (f *fakeStockCache) DecreaseStock(ctx context.Context, goodsID, userID int64) (cache.StockResult, error) {
	return f.decreaseResult, f.decreaseErr
}

func (f *fakeStockCache) SetResult(ctx context.Context, userID, goodsID int64, result string, ttl time.Duration) error {
	f.setResultCalls++
	f.lastResult = result
	return nil
}

type fakeTokenCache struct {
	verifyOK  bool
	verifyErr error
	setErr    error
}

func (f *fakeTokenCache) SetURLToken(ctx context.Context, userID, goodsID int64, token string, ttl time.Duration) error {
	return f.setErr
}

func (f *fakeTokenCache) VerifyURLToken(ctx context.Context, userID, goodsID int64, token string) (bool, error) {
	return f.verifyOK, f.verifyErr
}

type fakeProducer struct {
	err       error
	published bool
}

func (f *fakeProducer) PublishSeckillOrder(ctx context.Context, msg SeckillOrderMessage) error {
	if f.err != nil {
		return f.err
	}
	f.published = true
	return nil
}

func TestIssueSeckillToken(t *testing.T) {
	svc, err := NewSeckillService(
		&fakeGoodsRepo{goods: &model.SeckillGoods{ID: 1}},
		&fakeStockCache{},
		&fakeTokenCache{},
		&fakeProducer{},
	)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	resp, err := svc.IssueSeckillToken(context.Background(), 10, 1)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if resp.SeckillToken == "" || resp.ExpireAt == 0 {
		t.Fatal("unexpected token response")
	}
}

func TestDoSeckillInvalidToken(t *testing.T) {
	svc, _ := NewSeckillService(
		&fakeGoodsRepo{goods: &model.SeckillGoods{ID: 1}},
		&fakeStockCache{},
		&fakeTokenCache{verifyOK: false},
		&fakeProducer{},
	)

	_, err := svc.DoSeckill(context.Background(), 1, 1, "bad")
	if !errors.Is(err, pkg.ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got: %v", err)
	}
}

func TestDoSeckillSoldOut(t *testing.T) {
	stock := &fakeStockCache{decreaseResult: cache.StockNoInventory}
	svc, _ := NewSeckillService(
		&fakeGoodsRepo{goods: &model.SeckillGoods{ID: 1}},
		stock,
		&fakeTokenCache{verifyOK: true},
		&fakeProducer{},
	)

	_, err := svc.DoSeckill(context.Background(), 1, 1, "ok")
	if !errors.Is(err, pkg.ErrSoldOut) {
		t.Fatalf("expected sold out error, got: %v", err)
	}
	if stock.lastResult != cache.ResultFailed {
		t.Fatalf("expected failed result set, got: %s", stock.lastResult)
	}
}

func TestDoSeckillSuccess(t *testing.T) {
	producer := &fakeProducer{}
	stock := &fakeStockCache{decreaseResult: cache.StockSuccess}
	svc, _ := NewSeckillService(
		&fakeGoodsRepo{goods: &model.SeckillGoods{ID: 1}},
		stock,
		&fakeTokenCache{verifyOK: true},
		producer,
	)

	resp, err := svc.DoSeckill(context.Background(), 1, 1, "ok")
	if err != nil {
		t.Fatalf("do seckill: %v", err)
	}
	if resp.QueueStatus != "queuing" {
		t.Fatalf("unexpected queue status: %s", resp.QueueStatus)
	}
	if !producer.published {
		t.Fatal("message should be published")
	}
	if stock.lastResult != cache.ResultQueueing {
		t.Fatalf("expected queueing result, got: %s", stock.lastResult)
	}
}

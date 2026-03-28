package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"seckill/internal/cache"
	"seckill/internal/model"
	"seckill/pkg"
)

type goodsReader interface {
	GetByID(ctx context.Context, goodsID int64) (*model.SeckillGoods, error)
}

type stockOperator interface {
	DecreaseStock(ctx context.Context, goodsID, userID int64) (cache.StockResult, error)
	SetResult(ctx context.Context, userID, goodsID int64, result string, ttl time.Duration) error
}

type tokenVerifier interface {
	SetURLToken(ctx context.Context, userID, goodsID int64, token string, ttl time.Duration) error
	VerifyURLToken(ctx context.Context, userID, goodsID int64, token string) (bool, error)
}

type orderMessageProducer interface {
	PublishSeckillOrder(ctx context.Context, msg SeckillOrderMessage) error
}

type SeckillOrderMessage struct {
	UserID        int64
	GoodsID       int64
	IdempotentKey string
	CreatedAt     time.Time
}

type SeckillService struct {
	goodsRepo    goodsReader
	stockCache   stockOperator
	tokenCache   tokenVerifier
	producer     orderMessageProducer
	queueTTL     time.Duration
	urlTokenTTL  time.Duration
	soldOutFlags sync.Map
}

func NewSeckillService(
	goodsRepo goodsReader,
	stockCache stockOperator,
	tokenCache tokenVerifier,
	producer orderMessageProducer,
) (*SeckillService, error) {
	if goodsRepo == nil || stockCache == nil || tokenCache == nil || producer == nil {
		return nil, fmt.Errorf("service dependencies must not be nil")
	}
	return &SeckillService{
		goodsRepo:   goodsRepo,
		stockCache:  stockCache,
		tokenCache:  tokenCache,
		producer:    producer,
		queueTTL:    10 * time.Minute,
		urlTokenTTL: 5 * time.Minute,
	}, nil
}

func (s *SeckillService) IssueSeckillToken(ctx context.Context, userID, goodsID int64) (*model.SeckillTokenResp, error) {
	if _, err := s.goodsRepo.GetByID(ctx, goodsID); err != nil {
		return nil, err
	}
	token, err := cache.GenerateToken()
	if err != nil {
		return nil, err
	}
	if err = s.tokenCache.SetURLToken(ctx, userID, goodsID, token, s.urlTokenTTL); err != nil {
		return nil, err
	}
	return &model.SeckillTokenResp{
		SeckillToken: token,
		ExpireAt:     time.Now().Add(s.urlTokenTTL).Unix(),
	}, nil
}

func (s *SeckillService) DoSeckill(ctx context.Context, userID, goodsID int64, urlToken string) (*model.QueueResp, error) {
	ok, err := s.tokenCache.VerifyURLToken(ctx, userID, goodsID, urlToken)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, pkg.ErrInvalidToken
	}

	if s.isLocalSoldOut(goodsID) {
		return nil, pkg.ErrSoldOut
	}

	result, err := s.stockCache.DecreaseStock(ctx, goodsID, userID)
	if err != nil {
		return nil, err
	}

	switch result {
	case cache.StockNoInventory:
		s.markLocalSoldOut(goodsID)
		_ = s.stockCache.SetResult(ctx, userID, goodsID, cache.ResultFailed, s.queueTTL)
		return nil, pkg.ErrSoldOut
	case cache.StockAlreadyBought:
		_ = s.stockCache.SetResult(ctx, userID, goodsID, cache.ResultFailed, s.queueTTL)
		return nil, pkg.ErrAlreadyBought
	case cache.StockSuccess:
		// continue
	default:
		return nil, fmt.Errorf("unknown stock result: %d", result)
	}

	idempotentKey, err := cache.GenerateToken()
	if err != nil {
		return nil, err
	}

	msg := SeckillOrderMessage{
		UserID:        userID,
		GoodsID:       goodsID,
		IdempotentKey: idempotentKey,
		CreatedAt:     time.Now(),
	}
	if err = s.producer.PublishSeckillOrder(ctx, msg); err != nil {
		_ = s.stockCache.SetResult(ctx, userID, goodsID, cache.ResultFailed, s.queueTTL)
		return nil, err
	}

	if err = s.stockCache.SetResult(ctx, userID, goodsID, cache.ResultQueueing, s.queueTTL); err != nil {
		return nil, err
	}

	return &model.QueueResp{QueueStatus: "queuing"}, nil
}

func (s *SeckillService) isLocalSoldOut(goodsID int64) bool {
	v, ok := s.soldOutFlags.Load(goodsID)
	if !ok {
		return false
	}
	flag, _ := v.(bool)
	return flag
}

func (s *SeckillService) markLocalSoldOut(goodsID int64) {
	s.soldOutFlags.Store(goodsID, true)
}

package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/redis/go-redis/v9"

	"seckill/config"
	"seckill/internal/cache"
	"seckill/internal/model"
	"seckill/pkg"
)

type orderRepo interface {
	Create(ctx context.Context, order *model.SeckillOrder) error
	GetByIdempotentKey(ctx context.Context, key string) (*model.SeckillOrder, error)
}

type goodsRepo interface {
	GetByID(ctx context.Context, goodsID int64) (*model.SeckillGoods, error)
	DecreaseAvailableStock(ctx context.Context, goodsID int64) error
}

type resultWriter interface {
	SetResult(ctx context.Context, userID, goodsID int64, result string, ttl time.Duration) error
}

type idGenerator interface {
	NextID() (int64, error)
}

type Consumer struct {
	consumer   rocketmq.PushConsumer
	topic      string
	orderRepo  orderRepo
	goodsRepo  goodsRepo
	stockCache resultWriter
	idGen      idGenerator
	payTTL     time.Duration
	resultTTL  time.Duration
}

func NewConsumer(
	cfg config.RocketMQConfig,
	orderRepo orderRepo,
	goodsRepo goodsRepo,
	stockCache resultWriter,
	idGen idGenerator,
) (*Consumer, error) {
	if len(cfg.NameServers) == 0 {
		return nil, fmt.Errorf("rocketmq name servers are empty")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("rocketmq topic is empty")
	}
	if orderRepo == nil || goodsRepo == nil || stockCache == nil || idGen == nil {
		return nil, fmt.Errorf("consumer dependencies must not be nil")
	}

	nameSrv, err := primitive.NewNamesrvAddr(cfg.NameServers...)
	if err != nil {
		return nil, fmt.Errorf("build namesrv: %w", err)
	}

	c, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(cfg.GroupPrefix+"_order_consumer"),
		consumer.WithNameServer(nameSrv),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
	)
	if err != nil {
		return nil, err
	}

	mc := &Consumer{
		consumer:   c,
		topic:      cfg.Topic,
		orderRepo:  orderRepo,
		goodsRepo:  goodsRepo,
		stockCache: stockCache,
		idGen:      idGen,
		payTTL:     15 * time.Minute,
		resultTTL:  10 * time.Minute,
	}

	err = c.Subscribe(cfg.Topic, consumer.MessageSelector{Type: consumer.TAG, Expression: "*"}, mc.consume)
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func (c *Consumer) Start() error {
	return c.consumer.Start()
}

func (c *Consumer) Shutdown() error {
	return c.consumer.Shutdown()
}

func (c *Consumer) consume(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		if err := c.handleMessage(ctx, msg.Body); err != nil {
			return consumer.ConsumeRetryLater, err
		}
	}
	return consumer.ConsumeSuccess, nil
}

func (c *Consumer) handleMessage(ctx context.Context, body []byte) error {
	var payload seckillOrderPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	order, err := c.orderRepo.GetByIdempotentKey(ctx, payload.IdempotentKey)
	if err == nil {
		_ = c.stockCache.SetResult(ctx, payload.UserID, payload.GoodsID, cache.BuildResultValue(cache.ResultSuccess, order.ID), c.resultTTL)
		return nil
	}
	if !errors.Is(err, pkg.ErrOrderNotFound) {
		return err
	}

	goods, err := c.goodsRepo.GetByID(ctx, payload.GoodsID)
	if err != nil {
		_ = c.stockCache.SetResult(ctx, payload.UserID, payload.GoodsID, cache.BuildResultValue(cache.ResultFailed, 0), c.resultTTL)
		if errors.Is(err, pkg.ErrGoodsNotFound) {
			return nil
		}
		return err
	}

	if err = c.goodsRepo.DecreaseAvailableStock(ctx, payload.GoodsID); err != nil {
		_ = c.stockCache.SetResult(ctx, payload.UserID, payload.GoodsID, cache.BuildResultValue(cache.ResultFailed, 0), c.resultTTL)
		if errors.Is(err, pkg.ErrSoldOut) {
			return nil
		}
		return err
	}

	orderID, err := c.idGen.NextID()
	if err != nil {
		return err
	}

	newOrder := &model.SeckillOrder{
		ID:            orderID,
		UserID:        payload.UserID,
		GoodsID:       payload.GoodsID,
		SeckillPrice:  goods.SeckillPrice,
		Status:        model.OrderStatusPending,
		IdempotentKey: payload.IdempotentKey,
		PayExpireAt:   time.Now().Add(c.payTTL),
	}
	if err = c.orderRepo.Create(ctx, newOrder); err != nil {
		if isDuplicateError(err) {
			dupOrder, getErr := c.orderRepo.GetByIdempotentKey(ctx, payload.IdempotentKey)
			if getErr == nil {
				_ = c.stockCache.SetResult(ctx, payload.UserID, payload.GoodsID, cache.BuildResultValue(cache.ResultSuccess, dupOrder.ID), c.resultTTL)
				return nil
			}
			_ = c.stockCache.SetResult(ctx, payload.UserID, payload.GoodsID, cache.BuildResultValue(cache.ResultSuccess, 0), c.resultTTL)
			return nil
		}
		return err
	}

	return c.stockCache.SetResult(ctx, payload.UserID, payload.GoodsID, cache.BuildResultValue(cache.ResultSuccess, orderID), c.resultTTL)
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, redis.TxFailedErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}

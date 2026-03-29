package mq

import (
	"context"
	"encoding/json"
	"fmt"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"

	"seckill/config"
	"seckill/internal/service"
)

type Producer struct {
	producer rocketmq.Producer
	topic    string
}

type seckillOrderPayload struct {
	UserID        int64  `json:"user_id"`
	GoodsID       int64  `json:"goods_id"`
	IdempotentKey string `json:"idempotent_key"`
	CreatedAt     int64  `json:"created_at"`
}

func NewProducer(cfg config.RocketMQConfig) (*Producer, error) {
	if len(cfg.NameServers) == 0 {
		return nil, fmt.Errorf("rocketmq name servers are empty")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("rocketmq topic is empty")
	}

	nameSrv, err := primitive.NewNamesrvAddr(cfg.NameServers...)
	if err != nil {
		return nil, fmt.Errorf("build namesrv: %w", err)
	}

	p, err := rocketmq.NewProducer(
		producer.WithNameServer(nameSrv),
		producer.WithGroupName(cfg.GroupPrefix+"_producer"),
		producer.WithRetry(2),
	)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: p,
		topic:    cfg.Topic,
	}, nil
}

func (p *Producer) Start() error {
	return p.producer.Start()
}

func (p *Producer) Shutdown() error {
	return p.producer.Shutdown()
}

func (p *Producer) PublishSeckillOrder(ctx context.Context, msg service.SeckillOrderMessage) error {
	payload := seckillOrderPayload{
		UserID:        msg.UserID,
		GoodsID:       msg.GoodsID,
		IdempotentKey: msg.IdempotentKey,
		CreatedAt:     msg.CreatedAt.Unix(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	rmqMsg := primitive.NewMessage(p.topic, body)
	_, err = p.producer.SendSync(ctx, rmqMsg)
	return err
}

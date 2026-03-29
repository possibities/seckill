package mq

import (
	"testing"

	"seckill/config"
)

func TestNewProducerConfigValidation(t *testing.T) {
	if _, err := NewProducer(config.RocketMQConfig{}); err == nil {
		t.Fatal("expected error for empty config")
	}
}

package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultTokenTTL = 5 * time.Minute

type TokenCache struct {
	rdb *redis.Client
}

func NewTokenCache(rdb *redis.Client) (*TokenCache, error) {
	if rdb == nil {
		return nil, errors.New("nil redis client")
	}
	return &TokenCache{rdb: rdb}, nil
}

func GenerateToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (c *TokenCache) SetIdempotentToken(ctx context.Context, userID, goodsID int64, token string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	ok, err := c.rdb.SetNX(ctx, IdempotentTokenKey(userID, goodsID), token, ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (c *TokenCache) VerifyIdempotentToken(ctx context.Context, userID, goodsID int64, token string) (bool, error) {
	stored, err := c.rdb.Get(ctx, IdempotentTokenKey(userID, goodsID)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return stored == token, nil
}

func (c *TokenCache) SetURLToken(ctx context.Context, userID, goodsID int64, token string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	return c.rdb.Set(ctx, URLTokenKey(userID, goodsID), token, ttl).Err()
}

func (c *TokenCache) VerifyURLToken(ctx context.Context, userID, goodsID int64, token string) (bool, error) {
	stored, err := c.rdb.Get(ctx, URLTokenKey(userID, goodsID)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return stored == token, nil
}

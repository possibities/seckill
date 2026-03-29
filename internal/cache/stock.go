package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type StockResult int

const (
	StockSuccess       StockResult = 1
	StockNoInventory   StockResult = -1
	StockAlreadyBought StockResult = -2
)

const (
	ResultQueueing = "0"
	ResultSuccess  = "1"
	ResultFailed   = "2"
)

var seckillStockScript = redis.NewScript(`
if redis.call('SISMEMBER', KEYS[3], ARGV[1]) == 1 then
  return -2
end

local stock = tonumber(redis.call('GET', KEYS[1]))
if (not stock) or stock <= 0 then
  redis.call('SET', KEYS[2], '1')
  return -1
end

redis.call('DECR', KEYS[1])
redis.call('SADD', KEYS[3], ARGV[1])

local left = tonumber(redis.call('GET', KEYS[1]))
if left <= 0 then
  redis.call('SET', KEYS[2], '1')
end

return 1
`)

type StockCache struct {
	rdb *redis.Client
}

func NewStockCache(rdb *redis.Client) (*StockCache, error) {
	if rdb == nil {
		return nil, errors.New("nil redis client")
	}
	return &StockCache{rdb: rdb}, nil
}

func (c *StockCache) InitStock(ctx context.Context, goodsID int64, stock int) error {
	if stock < 0 {
		return fmt.Errorf("invalid stock: %d", stock)
	}
	pipe := c.rdb.TxPipeline()
	pipe.Set(ctx, StockKey(goodsID), stock, 0)
	pipe.Del(ctx, SoldOutKey(goodsID))
	pipe.Del(ctx, UsersKey(goodsID))
	_, err := pipe.Exec(ctx)
	return err
}

func (c *StockCache) DecreaseStock(ctx context.Context, goodsID, userID int64) (StockResult, error) {
	res, err := seckillStockScript.Run(ctx, c.rdb, []string{
		StockKey(goodsID),
		SoldOutKey(goodsID),
		UsersKey(goodsID),
	}, userID).Result()
	if err != nil {
		return 0, err
	}

	result, parseErr := parseScriptResult(res)
	if parseErr != nil {
		return 0, parseErr
	}
	return result, nil
}

func (c *StockCache) IsSoldOut(ctx context.Context, goodsID int64) (bool, error) {
	val, err := c.rdb.Get(ctx, SoldOutKey(goodsID)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

func (c *StockCache) SetResult(ctx context.Context, userID, goodsID int64, result string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return c.rdb.Set(ctx, ResultKey(userID, goodsID), result, ttl).Err()
}

func (c *StockCache) GetResult(ctx context.Context, userID, goodsID int64) (string, error) {
	val, err := c.rdb.Get(ctx, ResultKey(userID, goodsID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func BuildResultValue(status string, orderID int64) string {
	if status == ResultSuccess && orderID > 0 {
		return fmt.Sprintf("%s:%d", ResultSuccess, orderID)
	}
	return status
}

func ParseResultValue(value string) (status string, orderID int64, err error) {
	if value == "" {
		return ResultQueueing, 0, nil
	}
	parts := strings.Split(value, ":")
	if len(parts) == 1 {
		switch parts[0] {
		case ResultQueueing, ResultSuccess, ResultFailed:
			return parts[0], 0, nil
		default:
			return "", 0, fmt.Errorf("unknown result status: %s", parts[0])
		}
	}
	if len(parts) == 2 && parts[0] == ResultSuccess {
		id, parseErr := strconv.ParseInt(parts[1], 10, 64)
		if parseErr != nil {
			return "", 0, fmt.Errorf("parse result order id: %w", parseErr)
		}
		return ResultSuccess, id, nil
	}
	return "", 0, fmt.Errorf("invalid result value: %s", value)
}

func parseScriptResult(v any) (StockResult, error) {
	switch x := v.(type) {
	case int64:
		return StockResult(x), nil
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse script result: %w", err)
		}
		return StockResult(n), nil
	default:
		return 0, fmt.Errorf("unexpected lua result type: %T", v)
	}
}

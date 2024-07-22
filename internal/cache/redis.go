package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache struct {
	redis *redis.Client
}

//::: NEW redis

func NewRedis(connStr string) (*Cache, error) {
	opt, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	return &Cache{client}, nil
}

//::: get from redis

func (c *Cache) GetFrom(ctx context.Context, key string) (string, error) {
	getValue := c.redis.Get(ctx, key)
	return getValue.Val(), getValue.Err()
}

// ::: set to redis
func (c *Cache) SetTo(ctx context.Context, key string, value string) error {
	return c.redis.Set(ctx, key, value, time.Minute*5).Err()
}

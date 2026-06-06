package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/trace/trace/config"
)

// Client wraps the go-redis client.
// All repositories embed this to get the underlying connection.
type Client struct {
	rdb *redis.Client
}

func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// RDB exposes the underlying redis.Client for repository use.
func (c *Client) RDB() *redis.Client {
	return c.rdb
}

package config

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func (cfg Config) NewRedisClient() (*redis.Client, error) {

	connect := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	client := redis.NewClient(&redis.Options{
		Addr: connect,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return client, nil
}

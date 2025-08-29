package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"product-service/internal/core/domain/entity"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/gommon/log"
)

type CartRedisRepositoryInterface interface {
	AddToCart(ctx context.Context, userID int64, items []entity.CartItem) error
	GetCart(ctx context.Context, userID int64) ([]entity.CartItem, error)
}

type CartRedisRepository struct {
	Client *redis.Client
}

// GetCart implements CartRedisRepositoryInterface.
func (c *CartRedisRepository) GetCart(ctx context.Context, userID int64) ([]entity.CartItem, error) {

	val, err := c.Client.Get(ctx, fmt.Sprintf("cart:%v", userID)).Result()

	if err == redis.Nil {
		log.Infof("[CartRedisRepository-1] GetCart: Cart not found")
		return nil, nil
	}

	if err != nil {
		log.Errorf("[CartRedisRepository-2] GetCart: %v", err)
		return nil, err
	}

	var items []entity.CartItem
	err = json.Unmarshal([]byte(val), &items)
	if err != nil {
		log.Errorf("[CartRedisRepository-3] GetCart: %v", err)
		return nil, err
	}

	return items, nil

}

// AddToCart implements CartRedisRepositoryInterface.
func (c *CartRedisRepository) AddToCart(ctx context.Context, userID int64, items []entity.CartItem) error {

	data, err := json.Marshal(items)
	if err != nil {
		log.Errorf("[CartRedisRepository-1] AddToCart: %v", err)
		return err
	}
	return c.Client.Set(ctx, fmt.Sprintf("cart: %v", userID), data, 0).Err()

}

func NewCartRedisRepository(client *redis.Client) CartRedisRepositoryInterface {
	return &CartRedisRepository{
		Client: client,
	}
}

package repository

import (
	"context"
	"encoding/json"
	"product-service/internal/core/domain/entity"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type CartRedisRepositoryInterface interface {
	AddToCart(ctx context.Context, key string, items []entity.CartItem) error
	GetCart(ctx context.Context, key string) ([]entity.CartItem, error)
	RemoveFromCart(ctx context.Context, key string, productID uuid.UUID) error
	RemoveAllCart(ctx context.Context, key string) error
}

type CartRedisRepository struct {
	Client *redis.Client
}

// RemoveAllCart implements CartRedisRepositoryInterface.
func (c *CartRedisRepository) RemoveAllCart(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// RemoveFromCart implements CartRedisRepositoryInterface.
func (c *CartRedisRepository) RemoveFromCart(ctx context.Context, key string, productID uuid.UUID) error {

	cart, err := c.GetCart(ctx, key)
	if err != nil {
		log.Errorf("[CartRedisRepository-1] RemoveFromCart: %v", err)
		return err
	}

	newCart := []entity.CartItem{}
	for _, item := range cart {
		if item.ProductID != productID {
			newCart = append(newCart, item)
		}
	}

	err = c.Client.Del(ctx, key).Err()
	if err != nil {
		log.Errorf("[CartRedisRepository-2] RemoveFromCart: %v", err)
		return err
	}

	return c.AddToCart(ctx, key, newCart)

}

// GetCart implements CartRedisRepositoryInterface.
func (c *CartRedisRepository) GetCart(ctx context.Context, key string) ([]entity.CartItem, error) {

	val, err := c.Client.Get(ctx, key).Result()

	if err == redis.Nil {
		log.Infof("[CartRedisRepository-1] GetCart: Cart not found")
		return []entity.CartItem{}, nil
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
func (c *CartRedisRepository) AddToCart(ctx context.Context, key string, items []entity.CartItem) error {

	data, err := json.Marshal(items)
	if err != nil {
		log.Errorf("[CartRedisRepository-1] AddToCart: %v", err)
		return err
	}
	return c.Client.Set(ctx, key, data, 0).Err()

}

func NewCartRedisRepository(client *redis.Client) CartRedisRepositoryInterface {
	return &CartRedisRepository{
		Client: client,
	}
}

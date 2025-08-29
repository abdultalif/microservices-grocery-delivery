package service

import (
	"context"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
)

type CartServiceInterface interface {
	AddToCart(ctx context.Context, userID int64, req entity.CartItem) error
}

type cartService struct {
	cartRepository repository.CartRedisRepositoryInterface
}

// AddToCart implements CartServiceInterface.
func (c *cartService) AddToCart(ctx context.Context, userID int64, req entity.CartItem) error {
	cart, err := c.cartRepository.GetCart(ctx, userID)
	if err != nil {
		log.Errorf("[CartService-1] AddToCart: %v", err)
		return err
	}

	found := false
	for i, item := range cart {
		if item.ProductID == req.ProductID {
			cart[i].Quantity += req.Quantity
			found = true
			break
		}
	}

	if !found {
		cart = append(cart, req)
	}

	return c.cartRepository.AddToCart(ctx, userID, cart)
}

func NewCartService(cartRepository repository.CartRedisRepositoryInterface) CartServiceInterface {
	return &cartService{
		cartRepository: cartRepository,
	}
}

package service

import (
	"context"
	"fmt"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type CartServiceInterface interface {
	AddToCart(ctx context.Context, userID int64, req entity.CartItem) ([]entity.CartItem, error)
	GetCartByUserID(ctx context.Context, userID int64) ([]entity.CartItem, error)
	RemoveFromCart(ctx context.Context, userID int64, productID uuid.UUID) error
	RemoveAllCart(ctx context.Context, userID int64) error
}

type cartService struct {
	cartRepository repository.CartRedisRepositoryInterface
}

// RemoveAllCart implements CartServiceInterface.
func (c *cartService) RemoveAllCart(ctx context.Context, userID int64) error {
	return c.cartRepository.RemoveAllCart(ctx, fmt.Sprintf("cart:%d", userID))
}

// RemoveFromCart implements CartServiceInterface.
func (c *cartService) RemoveFromCart(ctx context.Context, userID int64, productID uuid.UUID) error {
	return c.cartRepository.RemoveFromCart(ctx, fmt.Sprintf("cart:%d", userID), productID)
}

// GetCartByUserID implements CartServiceInterface.
func (c *cartService) GetCartByUserID(ctx context.Context, userID int64) ([]entity.CartItem, error) {
	key := fmt.Sprintf("cart:%d", userID)
	cart, err := c.cartRepository.GetCart(ctx, key)
	if err != nil {
		log.Errorf("[CartService-1] GetCartByUserID: %v", err)
		return nil, err
	}

	return cart, nil
}

// AddToCart implements CartServiceInterface.
func (c *cartService) AddToCart(ctx context.Context, userID int64, req entity.CartItem) ([]entity.CartItem, error) {

	key := fmt.Sprintf("cart:%d", userID)

	cart, err := c.cartRepository.GetCart(ctx, key)
	if err != nil {
		log.Errorf("[CartService-1] AddToCart: %v", err)
		return nil, err
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

	if err := c.cartRepository.AddToCart(ctx, key, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func NewCartService(cartRepository repository.CartRedisRepositoryInterface) CartServiceInterface {
	return &cartService{
		cartRepository: cartRepository,
	}
}

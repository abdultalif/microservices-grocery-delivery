package repository

import (
	"context"
	"order-service/internal/core/domain/entity"
	"order-service/internal/core/domain/model"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	CreateOrder(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error)
}
type OrderRepository struct {
	db *gorm.DB
}

// CreateOrder implements OrderRepositoryInterface.
func (o *OrderRepository) CreateOrder(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error) {

	orderDate, err := time.Parse("2006-01-02", req.OrderDate) // YY-MM-DD format
 	if err != nil {
		log.Errorf("[OrderRepository-1] CreateOrder: %v", err)
		return uuid.Nil, err
	}

	var orderItems []model.OrderItem
	for _, item := range req.OrderItems {
		orderItem := model.OrderItem{
			ProductID: item.ProductID,
			Quantity: item.Quantity,
		}
		orderItems = append(orderItems, orderItem)
	}

	modelOrder := model.Order{
		BuyerID:     req.BuyerID,
		OrderCode:   req.OrderCode,
		OrderDate:  orderDate,
		OrderTime:  req.OrderTime,
		Status: 	req.Status,
		TotalAmount: float64(req.TotalAmount),
		ShippingType: req.ShippingType,
		ShippingFee:  float64(req.ShippingFee),
		Remarks:     req.Remarks,
		OrderItems:  orderItems,
	}

	if err := o.db.Create(&modelOrder).Error; err != nil {
		log.Errorf("[OrderRepository-3] CreateOrder: %v", err)
		return uuid.Nil, err
	}

	return modelOrder.ID, nil

}

func NewOrderRepository(db *gorm.DB) OrderRepositoryInterface {
	return &OrderRepository{
		db: db,
	}
}

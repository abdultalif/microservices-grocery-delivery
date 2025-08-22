package repository

import (
	"context"
	"errors"
	"order-service/internal/core/domain/entity"
	errs "order-service/internal/core/domain/error"
	"order-service/internal/core/domain/model"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	CreateOrder(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error)
	GetByID(ctx context.Context, orderID uuid.UUID) (*entity.OrderEntity, error)
}
type OrderRepository struct {
	db *gorm.DB
}

// GetByID implements OrderRepositoryInterface.
func (o *OrderRepository) GetByID(ctx context.Context, orderID uuid.UUID) (*entity.OrderEntity, error) {

	modelOrder := model.Order{}

	if err := o.db.Preload("OrderItems").Where("id = ?", orderID).First(&modelOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("[OrderRepository-1] GetByID: %v", err)
			return nil, errs.ErrNotFoundOrder
		}
		log.Errorf("[OrderRepository-1] GetByID: %v", err)
		return nil, err
	}

	orderItemEntities := []entity.OrderItemEntity{}
	for _, item := range modelOrder.OrderItems {
		orderItemEntities = append(orderItemEntities, entity.OrderItemEntity{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &entity.OrderEntity{
		ID:           modelOrder.ID,
		BuyerID:      modelOrder.BuyerID,
		OrderCode:    modelOrder.OrderCode,
		Status:       modelOrder.Status,
		OrderDate:    modelOrder.OrderDate.Format("2006-01-02 15:04:05"),
		TotalAmount:  int64(modelOrder.TotalAmount),
		OrderItems:   orderItemEntities,
		Remarks:      modelOrder.Remarks,
		ShippingType: modelOrder.ShippingType,
		ShippingFee:  int64(modelOrder.ShippingFee),
	}, nil

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
			Quantity:  item.Quantity,
		}
		orderItems = append(orderItems, orderItem)
	}

	modelOrder := model.Order{
		BuyerID:      req.BuyerID,
		OrderCode:    req.OrderCode,
		OrderDate:    orderDate,
		OrderTime:    req.OrderTime,
		Status:       req.Status,
		TotalAmount:  float64(req.TotalAmount),
		ShippingType: req.ShippingType,
		ShippingFee:  float64(req.ShippingFee),
		Remarks:      req.Remarks,
		OrderItems:   orderItems,
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

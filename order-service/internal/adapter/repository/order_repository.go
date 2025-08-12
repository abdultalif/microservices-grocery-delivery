package repository

import (
	"context"
	"errors"
	"math"
	"order-service/internal/core/domain/entity"
	errs "order-service/internal/core/domain/error"
	"order-service/internal/core/domain/model"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.OrderEntity, error)
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

	orderTime, err := time.Parse("15:04:05", req.OrderTime) // HH:MM:SS format
	if err != nil {
		log.Errorf	("[OrderRepository-2] CreateOrder: %v", err)
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
		OrderTime:  orderTime,
		Status: 	req.Status,
		TotalAmount: req.TotalAmount,
		ShippingType: req.ShippingType,
		ShipingFee:  req.ShipingFee,
		Remarks:     req.Remarks,
		OrderItems:  orderItems,
	}

	if err := o.db.Create(&modelOrder).Error; err != nil {
		log.Errorf("[OrderRepository-3] CreateOrder: %v", err)
		return uuid.Nil, err
	}

	return modelOrder.ID, nil

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
		TotalAmount:  modelOrder.TotalAmount,
		OrderItems:   orderItemEntities,
		Remarks:      modelOrder.Remarks,
		ShippingType: modelOrder.ShippingType,
		ShipingFee:   modelOrder.ShipingFee,
	}, nil

}

// GetAll implements OrderRepositoryInterface.
func (o *OrderRepository) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {

	var modelOrders = []model.Order{}
	var countData int64

	offset := (query.Page - 1) * query.Limit

	sqlMain := o.db.Preload("OrderItems").Where("order_code ILIKE ? OR status ILIKE ?", "%"+query.Search+"%", "%"+query.Status+"%")
	if err := sqlMain.Model(&modelOrders).Count(&countData).Error; err != nil {
		log.Errorf("[Order-Repository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))
	if err := sqlMain.Order("order_date DESC").Limit(int(query.Limit)).Offset(int(offset)).Find(&modelOrders).Error; err != nil {
		log.Errorf("[OrderRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelOrders) == 0 {
		err := errs.ErrNotFoundOrder
		log.Errorf("[OrderRepository-3] GetAll: %v", err)
		return nil, 0, 0, err
	}

	entities := []entity.OrderEntity{}
	for _, val := range modelOrders {
		orderItemEntities := []entity.OrderItemEntity{}
		for _, item := range val.OrderItems {
			orderItemEntities = append(orderItemEntities, entity.OrderItemEntity{
				ID:        item.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
		entities = append(entities, entity.OrderEntity{
			ID:          val.ID,
			OrderCode:   val.OrderCode,
			Status:      val.Status,
			OrderDate:   val.OrderDate.Format("2006-01-02 15:04:05"),
			TotalAmount: val.TotalAmount,
			OrderItems:  orderItemEntities,
			BuyerID:     val.BuyerID,
		})
	}

	return entities, int64(countData), int64(totalPage), nil

}

func NewOrderRepository(db *gorm.DB) OrderRepositoryInterface {
	return &OrderRepository{
		db: db,
	}
}

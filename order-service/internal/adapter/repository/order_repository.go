package repository

import (
	"context"
	"math"
	"order-service/internal/core/domain/entity"
	errs "order-service/internal/core/domain/error"
	"order-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
}
type OrderRepository struct {
	db *gorm.DB
}

// GetAll implements OrderRepositoryInterface.
func (o *OrderRepository) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {
	
	var modelOrders = []model.Order{}
	var countData int64

	offset := (query.Page -1) * query.Limit

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
	for _, val := range modelOrders{
		orderItemEntities := []entity.OrderItemEntity{}
			for _, item := range val.OrderItems{
				orderItemEntities = append(orderItemEntities, entity.OrderItemEntity{
					ID: item.ID,
					ProductID: item.ProductID,
					Quatity: item.Quatity,

				})
			}
		entities = append(entities, entity.OrderEntity{
			ID: val.ID,
			OrderCode: val.OrderCode,
			Status: val.Status,
			OrderDate: val.OrderDate.Format("2006-01-02 15:04:05"),
			TotalAmount: val.TotalAmount,
			OrderItems: orderItemEntities,
		})
	}

	return entities, int64(countData), int64(totalPage), nil

}

func NewOrderRepository(db *gorm.DB) OrderRepositoryInterface {
	return &OrderRepository{
		db: db,
	}
}

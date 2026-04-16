package repository

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/model"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID uuid.UUID) (*entity.OrderEntity, error)
	CreateOrder(ctx context.Context, req entity.OrderEntity) (uuid.UUID, error)
	UpdateStatus(ctx context.Context, req entity.OrderEntity) (int64, string, string, error)
	DeleteOrderByID(ctx context.Context, orderID uuid.UUID) error
	GetOrderByOrderCode(ctx context.Context, orderCode string) (*entity.OrderEntity, error)
	GetByIDCustomer(ctx context.Context, orderID uuid.UUID, buyerID int64) (*entity.OrderEntity, error)
	GetProductFromSnapshoot(productID uuid.UUID) (*entity.ProductResponseEntity, error)
}
type OrderRepository struct {
	db                   *gorm.DB
	productSnapshootRepo ProductSnapshootRepositoryInterface
}

// GetProductFromSnapshoot implements OrderRepositoryInterface.
func (o *OrderRepository) GetProductFromSnapshoot(productID uuid.UUID) (*entity.ProductResponseEntity, error) {
	productSnapshot, err := o.productSnapshootRepo.GetByID(productID)
	if err != nil {
		log.Errorf("[OrderRepository] GetProductFromSnapshoot-1: %v", err)
		return nil, err
	}

	return &entity.ProductResponseEntity{
		ID:           productSnapshot.ID,
		ProductName:  productSnapshot.Name,
		ProductImage: productSnapshot.Image,
		SalePrice:    float64(productSnapshot.SalePrice),
		Unit:         productSnapshot.Unit,
		Weight:       productSnapshot.Weight,
		Stock:        productSnapshot.Stock,
	}, nil
}

// GetByIDCustomer implements OrderRepositoryInterface.
func (o *OrderRepository) GetByIDCustomer(ctx context.Context, orderID uuid.UUID, buyerID int64) (*entity.OrderEntity, error) {

	modelOrder := model.Order{}

	if err := o.db.Preload("OrderItems").First(&modelOrder, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("[OrderRepository-1] GetByID: %v", err)
			return nil, errs.ErrNotFoundOrder
		}
		return nil, err
	}

	if buyerID != 0 && modelOrder.BuyerID != buyerID {
		log.Warnf("[OrderRepository-2] Forbidden access, orderID=%s buyerID=%d tokenBuyerID=%d", orderID, modelOrder.BuyerID, buyerID)
		return nil, errs.ErrForbiddenOrder
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

// DeleteOrderByID implements OrderRepositoryInterface.
func (o *OrderRepository) DeleteOrderByID(ctx context.Context, orderID uuid.UUID) error {

	modelOrder := model.Order{}

	if err := o.db.Preload("OrderItems").Where("id = ?", orderID).First(&modelOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("[OrderRepository-1] DeleteOrderByID: Order not found")
			return errs.ErrNotFoundOrder
		}
		log.Errorf("[OrderRepository-2] DeleteOrderByID: %v", err)
		return err
	}

	if err := o.db.Select("OrderItems").Delete(&modelOrder).Error; err != nil {
		log.Errorf("[OrderRepository-3] DeleteOrderByID: %v", err)
		return err
	}

	return nil

}

// GetOrderByOrderCode implements OrderRepositoryInterface.
func (o *OrderRepository) GetOrderByOrderCode(ctx context.Context, orderCode string) (*entity.OrderEntity, error) {
	modelOrder := model.Order{}

	if err := o.db.Preload("OrderItems").Where("order_code = ?", orderCode).First(&modelOrder).Error; err != nil {
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
		BuyerName:    modelOrder.BuyerName,
		BuyerEmail:   modelOrder.BuyerEmail,
		BuyerPhone:   modelOrder.BuyerPhone,
		BuyerAddress: modelOrder.BuyerAddress,
		BuyerLat:     modelOrder.BuyerLat,
		BuyerLng:     modelOrder.BuyerLng,
	}, nil
}

// UpdateStatus implements OrderRepositoryInterface.
func (o *OrderRepository) UpdateStatus(ctx context.Context, req entity.OrderEntity) (int64, string, string, error) {

	modelOrder := model.Order{}

	if err := o.db.Select("id", "order_code", "status", "buyer_id", "remarks").First(&modelOrder, "id = ?", req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("[OrderRepository-1] UpdateStatus: Order not found")
			return 0, "", "", errs.ErrNotFoundOrder
		}
		log.Infof("[OrderRepository-2] UpdateStatus: %v", err)
		return 0, "", "", err
	}

	if modelOrder.Status == "Pending" && (req.Status != "Confirmed" && req.Status != "Cancelled") {
		log.Infof("[OrderRepository-3] UpdateStatus: Invalid status transition")
		return 0, "", "", errs.ErrInvalidStatus
	}

	if modelOrder.Status == "Confirmed" && (req.Status != "Process" && req.Status != "Cancelled") {
		log.Infof("[OrderRepository-4] UpdateStatus: Invalid status transition")
		return 0, "", "", errs.ErrInvalidStatus
	}

	if modelOrder.Status == "Process" && (req.Status != "Sending" && req.Status != "Cancelled") {
		log.Infof("[OrderRepository-5] UpdateStatus: Invalid status transition")
		return 0, "", "", errs.ErrInvalidStatus
	}

	if modelOrder.Status == "Sending" && (req.Status != "Done" && req.Status != "Cancelled") {
		log.Infof("[OrderRepository-6] UpdateStatus: Invalid status transition")
		return 0, "", "", errs.ErrInvalidStatus
	}

	modelOrder.Status = req.Status
	modelOrder.Remarks = req.Remarks

	if err := o.db.UpdateColumns(&modelOrder).Error; err != nil {
		log.Errorf("[OrderRepository-7] UpdateStatus: %v", err)
		return 0, "", "", err
	}

	return modelOrder.BuyerID, modelOrder.Status, modelOrder.OrderCode, nil

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
		BuyerName:    req.BuyerName,
		BuyerEmail:   req.BuyerEmail,
		BuyerPhone:   req.BuyerPhone,
		BuyerAddress: req.BuyerAddress,
		BuyerLat:     req.BuyerLat,
		BuyerLng:     req.BuyerLng,
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
		TotalAmount:  int64(modelOrder.TotalAmount),
		OrderItems:   orderItemEntities,
		Remarks:      modelOrder.Remarks,
		ShippingType: modelOrder.ShippingType,
		ShippingFee:  int64(modelOrder.ShippingFee),
		BuyerName:    modelOrder.BuyerName,
		BuyerEmail:   modelOrder.BuyerEmail,
		BuyerPhone:   modelOrder.BuyerPhone,
		BuyerAddress: modelOrder.BuyerAddress,
		BuyerLat:     modelOrder.BuyerLat,
		BuyerLng:     modelOrder.BuyerLng,
	}, nil

}

// GetAll implements OrderRepositoryInterface.
func (o *OrderRepository) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {

	var modelOrders = []model.Order{}
	var countData int64

	offset := (query.Page - 1) * query.Limit

	sqlMain := o.db.Preload("OrderItems").
		Where("order_code ILIKE ? OR status ILIKE ?", "%"+query.Search+"%", "%"+query.Status+"%")

	if query.BuyerID != 0 {
		sqlMain = sqlMain.Where("buyer_id = ?", query.BuyerID)
		log.Info("Filter by BuyerID:", query.BuyerID)
	}

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
			ID:           val.ID,
			OrderCode:    val.OrderCode,
			Status:       val.Status,
			OrderDate:    val.OrderDate.Format("2006-01-02 15:04:05"),
			TotalAmount:  int64(val.TotalAmount),
			OrderItems:   orderItemEntities,
			BuyerID:      val.BuyerID,
			BuyerName:    val.BuyerName,
			BuyerEmail:   val.BuyerEmail,
			BuyerPhone:   val.BuyerPhone,
			BuyerAddress: val.BuyerAddress,
			BuyerLat:     val.BuyerLat,
			BuyerLng:     val.BuyerLng,
		})
	}

	return entities, int64(countData), int64(totalPage), nil

}

func NewOrderRepository(db *gorm.DB, productSnapshootRepo ProductSnapshootRepositoryInterface) OrderRepositoryInterface {
	return &OrderRepository{
		db:                   db,
		productSnapshootRepo: productSnapshootRepo,
	}
}

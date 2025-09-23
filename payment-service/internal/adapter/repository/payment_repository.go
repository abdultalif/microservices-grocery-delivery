package repository

import (
	"context"
	"payment-service/internal/core/domain/entity"
	"payment-service/internal/core/domain/model"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type PaymentRepositoryInterface interface {
	CreatePayment(ctx context.Context, payment entity.PaymentEntity) error
	LogPayment(ctx context.Context, paymentID uuid.UUID, status string) error
	UpdateStatusByOrderCode(ctx context.Context, orderID uuid.UUID, status string) error
	GetByOrderID(ctx context.Context, orderID uuid.UUID) error
}
type PaymentRepository struct {
	db *gorm.DB
}

// GetByOrderID implements PaymentRepositoryInterface.
func (p *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) error {

	if err := p.db.WithContext(ctx).First(&model.Payment{}, "order_id = ?", orderID).Error; err != nil {
		log.Errorf("[PaymentRepository] GetByOrderID-1: %v", err)
		return err
	}

	return nil
}

// UpdateStatusByOrderCode implements PaymentRepositoryInterface.
func (p *PaymentRepository) UpdateStatusByOrderCode(ctx context.Context, orderID uuid.UUID, status string) error {
	modelPayment := model.Payment{}

	if err := p.db.WithContext(ctx).First(&modelPayment, "order_id = ?", orderID).Error; err != nil {
		log.Errorf("[PaymentRepository] UpdateStatusByOrderCode-1: %v", err)
		return err
	}

	modelPayment.PaymentStatus = status

	if err := p.db.WithContext(ctx).Save(&modelPayment).Error; err != nil {
		log.Errorf("[PaymentRepository] UpdateStatusByOrderCode-1: %v", err)
		return err
	}

	return nil

}

// LogPayment implements PaymentRepositoryInterface.
func (p *PaymentRepository) LogPayment(ctx context.Context, paymentID uuid.UUID, status string) error {

	logPayment := model.PaymentLog{
		PaymentID: paymentID,
		Status:    status,
	}

	if err := p.db.WithContext(ctx).Create(&logPayment).Error; err != nil {
		log.Errorf("[PaymentRepository-1] LogPayment: %v", err)
		return err
	}

	return nil

}

// CreatePayment implements PaymentRepositoryInterface.
func (p *PaymentRepository) CreatePayment(ctx context.Context, payment entity.PaymentEntity) error {

	modelPayment := model.Payment{
		OrderID:          payment.OrderID,
		UserID:           payment.UserID,
		PaymentMethod:    payment.PaymentMethod,
		PaymentStatus:    payment.PaymentStatus,
		PaymentGatewayID: &payment.PaymentGatewayID,
		GrossAmount:      payment.GrossAmount,
		PaymentURL:       &payment.PaymentURL,
	}

	if err := p.db.WithContext(ctx).Create(&modelPayment).Error; err != nil {
		log.Errorf("[PaymentRepository-1] create: %v", err)
		return err
	}

	return p.LogPayment(ctx, modelPayment.ID, modelPayment.PaymentStatus)

}

func NewPaymentRepository(db *gorm.DB) PaymentRepositoryInterface {
	return &PaymentRepository{
		db: db,
	}
}

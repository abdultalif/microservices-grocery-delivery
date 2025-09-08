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
	CreatePayment(ctx context.Context, payment *entity.PaymentEntity) error
	LogPayment(ctx context.Context, paymentID uuid.UUID, status string) error
}
type PaymentRepository struct {
	db gorm.DB
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
func (p *PaymentRepository) CreatePayment(ctx context.Context, payment *entity.PaymentEntity) error {

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

func NewPaymentRepository(db gorm.DB) PaymentRepositoryInterface {
	return &PaymentRepository{
		db: db,
	}
}

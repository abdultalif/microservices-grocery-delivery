package repository

import (
	"context"
	"errors"
	"math"

	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/entity"
	errs "github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/error"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/model"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepositoryInterface interface {
	CreatePayment(ctx context.Context, payment entity.PaymentEntity) error
	LogPayment(ctx context.Context, paymentID uuid.UUID, status string) error
	GetAll(ctx context.Context, req entity.PaymentQueryStringRequest) ([]entity.PaymentEntity, int64, int64, error)
	GetByOrderCode(ctx context.Context, orderCode string) error
	GetDetail(ctx context.Context, paymentID uuid.UUID) (*entity.PaymentEntity, error)

	GetByOrderCodeForUpdate(ctx context.Context, orderCode string) (*model.Payment, error)
	UpdateStatusByOrderCode(ctx context.Context, req entity.CancelTransaction, status string) error
}
type PaymentRepository struct {
	db *gorm.DB
}

func (p *PaymentRepository) UpdateStatusByOrderCode(ctx context.Context, req entity.CancelTransaction, status string) error {

	err := p.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("order_code = ? AND user_id = ?", req.OrderCode, req.UserID).
		Update("payment_status", status).Error

	if err != nil {
		log.Errorf("[PaymentRepository] UpdateStatusByGatewayID-1: %v", err)
		return errs.ErrNotFoundPayment
	}

	return nil
}

// GetByOrderCodeForUpdate implements PaymentRepositoryInterface.
func (p *PaymentRepository) GetByOrderCodeForUpdate(ctx context.Context, orderCode string) (*model.Payment, error) {

	var payment model.Payment

	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clauses(clause.Locking{Strength: "UPDATE"}) = SELECT ... FOR UPDATE
		// Row akan di-lock sampai transaksi ini commit/rollback
		return tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_code = ?", orderCode).
			First(&payment).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Belum ada payment → boleh proses
		}
		log.Errorf("[PaymentRepository] GetByOrderCodeForUpdate: %v", err)
		return nil, err
	}

	return &payment, nil

}

// GetDetail implements PaymentRepositoryInterface.
func (p *PaymentRepository) GetDetail(ctx context.Context, paymentID uuid.UUID) (*entity.PaymentEntity, error) {

	modelPayment := model.Payment{}

	if err := p.db.Where("id = ?", paymentID).First(&modelPayment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("[PaymentRepository-1] GetDetail: No payment found")
			return nil, errs.ErrNotFoundPayment
		}
		log.Errorf("[PaymentRepository-1] GetDetail: %v", err)
		return nil, err
	}

	return &entity.PaymentEntity{
		ID:               modelPayment.ID,
		OrderCode:        modelPayment.OrderCode,
		UserID:           modelPayment.UserID,
		PaymentMethod:    modelPayment.PaymentMethod,
		PaymentStatus:    modelPayment.PaymentStatus,
		PaymentGatewayID: *modelPayment.PaymentGatewayID,
		GrossAmount:      modelPayment.GrossAmount,
		PaymentURL:       *modelPayment.PaymentURL,
		PaymentAt:        modelPayment.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil

}

// GetAll implements PaymentRepositoryInterface.
func (p *PaymentRepository) GetAll(ctx context.Context, req entity.PaymentQueryStringRequest) ([]entity.PaymentEntity, int64, int64, error) {

	var countData int64
	modelPayments := []model.Payment{}

	offset := (req.Page - 1) * req.Limit

	sqlMain := p.db.WithContext(ctx).Where("payment_method ILIKE ? OR payment_status ILIKE ?", "%"+req.Search+"%", "%"+req.Status+"%")

	if req.UserID != 0 {
		sqlMain = sqlMain.Where("user_id = ?", req.UserID)
	}

	if err := sqlMain.Model(&modelPayments).Count(&countData).Error; err != nil {
		log.Errorf("[PaymentRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(req.Limit)))
	if err := sqlMain.Order("created_at DESC").Limit(int(req.Limit)).Offset(int(offset)).Find(&modelPayments).Error; err != nil {
		log.Errorf("[PaymentRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelPayments) == 0 {
		log.Infof("[PaymentRepository-3] GetAll: No payment found")
		return nil, 0, 0, errs.ErrNotFoundPayment
	}

	entities := []entity.PaymentEntity{}
	for _, val := range modelPayments {
		entities = append(entities, entity.PaymentEntity{
			ID:               val.ID,
			OrderCode:        val.OrderCode,
			UserID:           val.UserID,
			PaymentMethod:    val.PaymentMethod,
			PaymentStatus:    val.PaymentStatus,
			PaymentGatewayID: *val.PaymentGatewayID,
			GrossAmount:      val.GrossAmount,
			PaymentURL:       *val.PaymentURL,
		})
	}

	return entities, countData, int64(totalPage), nil

}

// GetByOrderCode implements PaymentRepositoryInterface.
func (p *PaymentRepository) GetByOrderCode(ctx context.Context, orderCode string) error {

	if err := p.db.WithContext(ctx).First(&model.Payment{}, "order_code = ?", orderCode).Error; err != nil {
		log.Errorf("[PaymentRepository] GetByOrderCode-1: %v", err)
		return err
	}

	return nil
}

// UpdateStatusByOrderCode implements PaymentRepositoryInterface.
// func (p *PaymentRepository) UpdateStatusByOrderCode(ctx context.Context, orderCode string, status string) error {
// 	modelPayment := model.Payment{}

// 	if err := p.db.WithContext(ctx).First(&modelPayment, "order_code = ?", orderCode).Error; err != nil {
// 		log.Errorf("[PaymentRepository] UpdateStatusByOrderCode-1: %v", err)
// 		return err
// 	}

// 	modelPayment.PaymentStatus = status

// 	if err := p.db.WithContext(ctx).Save(&modelPayment).Error; err != nil {
// 		log.Errorf("[PaymentRepository] UpdateStatusByOrderCode-1: %v", err)
// 		return err
// 	}

// 	return nil

// }

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
		OrderCode:        payment.OrderCode,
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

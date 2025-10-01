package repositories

import (
	"context"
	"fmt"
	"math"
	"notification-service/internal/domain/entity"
	errs "notification-service/internal/domain/error"
	"notification-service/internal/domain/model"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type NotifRepositoryInterface interface {
	GetAll(ctx context.Context, query entity.NotifyQuerySting) ([]entity.NotificationEntity, int64, int64, error)
	GetByID(ctx context.Context, notifID uuid.UUID) (*entity.NotificationEntity, error)
	MarkAsRead(ctx context.Context, notifID uuid.UUID) error
}

type NotifRepository struct {
	db *gorm.DB
}

// MarkAsRead implements NotifRepositoryInterface.
func (n *NotifRepository) MarkAsRead(ctx context.Context, notifID uuid.UUID) error {

	modelNotif := model.NotificationModel{}
	if err := n.db.First(&modelNotif, notifID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Errorf("[MarkAsRead-1] Record not found for notification ID %d", notifID)
			return errs.ErrNotFoundNotification
		}
		log.Errorf("[MarkAsRead-2] Failed to find notification by ID: %v", err)
		return err
	}

	now := time.Now()
	modelNotif.ReadAt = &now
	if err := n.db.UpdateColumns(&modelNotif).Error; err != nil {
		log.Errorf("[MarkAsRead-3] Failed to save notification: %v", err)
		return err
	}
	return nil

}

// GetByID implements NotifRepositoryInterface.
func (n *NotifRepository) GetByID(ctx context.Context, notifID uuid.UUID) (*entity.NotificationEntity, error) {
	modelNotif := model.NotificationModel{}

	if err := n.db.WithContext(ctx).Select("id", "subject", "status", "sent_at", "read_at", "message", "notification_type").First(&modelNotif, notifID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Errorf("[GetByID-1] Record not found for notification ID %d", notifID)
			err = errs.ErrNotFoundNotification
			return nil, err
		}
		log.Errorf("[GetByID-2] Failed to find notification by ID: %v", err)
		return nil, err
	}

	return &entity.NotificationEntity{
		ID:               modelNotif.ID,
		Subject:          modelNotif.Subject,
		Status:           modelNotif.Status,
		SentAt:           modelNotif.SentAt,
		ReadAt:           modelNotif.ReadAt,
		Message:          modelNotif.Message,
		NotificationType: modelNotif.NotificationType,
	}, nil
}

// GetAll implements NotifRepositoryInterface.
func (n *NotifRepository) GetAll(ctx context.Context, queryString entity.NotifyQuerySting) ([]entity.NotificationEntity, int64, int64, error) {
	modelNotifes := []model.NotificationModel{}

	var countData int64
	offset := (queryString.Page - 1) * queryString.Limit

	sqlMain := n.db.WithContext(ctx).
		Select("id", "subject", "status", "sent_at").
		Where("subject ILIKE ? OR message ILIKE ? OR status ILIKE ?", "%"+queryString.Search+"%", "%"+queryString.Search+"%", "%"+queryString.Status+"%")

	if queryString.UserID != 0 {
		sqlMain = sqlMain.Where("reciever_id = ?", queryString.UserID)
	}

	if queryString.IsRead {
		sqlMain = sqlMain.Where("read_at IS NOT NULL")
	}

	if err := sqlMain.Model(&modelNotifes).Count(&countData).Error; err != nil {
		log.Errorf("[NotificationRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	order := fmt.Sprintf("%s %s", queryString.OrderBy, queryString.OrderType)

	totalPage := int(math.Ceil(float64(countData) / float64(queryString.Limit)))
	if err := sqlMain.Order(order).Limit(int(queryString.Limit)).Offset(int(offset)).Find(&modelNotifes).Error; err != nil {
		log.Errorf("[NotificationRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelNotifes) == 0 {
		log.Infof("[NotificationRepository-3] GetAll: No notification found")
		return nil, 0, 0, errs.ErrNotFoundNotification
	}
	notifEntities := []entity.NotificationEntity{}
	for _, val := range modelNotifes {
		notifEntities = append(notifEntities, entity.NotificationEntity{
			ID:      val.ID,
			Subject: val.Subject,
			Status:  val.Status,
			SentAt:  val.SentAt,
		})
	}

	return notifEntities, countData, int64(totalPage), nil
}

func NewRepositoryNotif(db *gorm.DB) NotifRepositoryInterface {
	return &NotifRepository{db: db}
}

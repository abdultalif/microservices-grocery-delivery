package services

import (
	"context"
	"notification-service/internal/domain/entity"
	"notification-service/internal/repositories"

	"github.com/google/uuid"
)

type NotificationServiceInterface interface {
	GetAll(ctx context.Context, query entity.NotifyQuerySting) ([]entity.NotificationEntity, int64, int64, error)
	GetByID(ctx context.Context, notifID uuid.UUID) (*entity.NotificationEntity, error)
}

type NotificationService struct {
	repoNotification repositories.NotifRepositoryInterface
}

// GetByID implements NotificationServiceInterface.
func (n *NotificationService) GetByID(ctx context.Context, notifID uuid.UUID) (*entity.NotificationEntity, error) {
	return n.repoNotification.GetByID(ctx, notifID)
}

// GetAll implements NotificationServiceInterface.
func (n *NotificationService) GetAll(ctx context.Context, query entity.NotifyQuerySting) ([]entity.NotificationEntity, int64, int64, error) {
	return n.repoNotification.GetAll(ctx, query)
}

func NewServiceNotification(repoNotification repositories.NotifRepositoryInterface) NotificationServiceInterface {
	return &NotificationService{repoNotification: repoNotification}
}

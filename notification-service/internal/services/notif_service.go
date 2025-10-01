package services

import (
	"context"
	"notification-service/internal/domain/entity"
	"notification-service/internal/repositories"
)

type NotificationServiceInterface interface {
	GetAll(ctx context.Context, query entity.NotifyQuerySting) ([]entity.NotificationEntity, int64, int64, error)
}

type NotificationService struct {
	repoNotification repositories.NotifRepositoryInterface
}

// GetAll implements NotificationServiceInterface.
func (n *NotificationService) GetAll(ctx context.Context, query entity.NotifyQuerySting) ([]entity.NotificationEntity, int64, int64, error) {
	return n.repoNotification.GetAll(ctx, query)
}

func NewServiceNotification(repoNotification repositories.NotifRepositoryInterface) NotificationServiceInterface {
	return &NotificationService{repoNotification: repoNotification}
}

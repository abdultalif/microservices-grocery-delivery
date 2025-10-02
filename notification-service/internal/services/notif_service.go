package services

import (
	"context"
	"notification-service/internal/domain/entity"
	"notification-service/internal/pkg"
	"notification-service/internal/repositories"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type NotificationServiceInterface interface {
	GetAll(ctx context.Context, query entity.NotifyQuerySting) ([]entity.NotificationEntity, int64, int64, error)
	GetByID(ctx context.Context, notifID uuid.UUID) (*entity.NotificationEntity, error)
	MarkAsRead(ctx context.Context, notifID uuid.UUID) error
	SendPushNotification(ctx context.Context, notification entity.NotificationEntity)
}

type NotificationService struct {
	repoNotification repositories.NotifRepositoryInterface
}

// SendPushNotification implements NotificationServiceInterface.
func (n *NotificationService) SendPushNotification(ctx context.Context, notification entity.NotificationEntity) {

	if notification.ReceiverID == nil {
		return
	}
	conn := pkg.GetWebSocketConn(*notification.ReceiverID)
	if conn == nil {
		log.Errorf("[SendPushNotification-1] WebSocket connection not found for user %d", *notification.ReceiverID)
		return
	}

	msg := map[string]interface{}{
		"type":    notification.NotificationType,
		"subject": notification.Subject,
		"message": notification.Message,
		"sent_at": notification.SentAt,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Errorf("[SendPushNotification-2] Failed to send WebSocket notification: %v", err)
	}

	if err := n.repoNotification.MarkAsSent(notification.ID); err != nil {
		log.Errorf("[SendPushNotification-3] Failed to mark notification as sent: %v", err)
	}

}

// MarkAsRead implements NotificationServiceInterface.
func (n *NotificationService) MarkAsRead(ctx context.Context, notifID uuid.UUID) error {
	return n.repoNotification.MarkAsRead(ctx, notifID)
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

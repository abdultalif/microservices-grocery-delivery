package worker

import (
	"context"
	"encoding/json"
	"notification-service/internal/config"
	"notification-service/internal/domain/entity"
	"notification-service/internal/infrastructure/messaging"
	"notification-service/internal/repositories"
	"notification-service/internal/services"

	"github.com/labstack/gommon/log"
)

type ConsumeRabbitMQInterface interface {
	ConsumeMessage(queueName string) error
	SendNotification(notificationEntity entity.NotificationEntity)
}

type ConsumeRabbitMQ struct {
	emailService        messaging.MessageEmailInterface
	notifRepository     repositories.NotifRepositoryInterface
	notificationService services.NotificationServiceInterface
}

// ConsumeMessage implements ConsumeRabbitMQInterface.
func (c *ConsumeRabbitMQ) ConsumeMessage(queueName string) error {
	cfg := config.NewConfig()

	rabbitMQ := messaging.NewRabbitMQ(cfg.RabbitMQ)
	conn, err := rabbitMQ.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel: %v", err)
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Errorf("Failed to register a consumer: %v", err)
		return err
	}

	for msg := range msgs {
		var notification entity.NotificationEntity
		if err := json.Unmarshal(msg.Body, &notification); err != nil {
			log.Errorf("Failed to unmarshal: %v", err)
			continue
		}

		notification.Status = "PENDING"
		if notification.NotificationType == "EMAIL" {
			notification.Status = "SENT"
		}

		err = c.notifRepository.CreateNotification(context.Background(), notification)
		if err != nil {
			log.Errorf("Failed to create notification: %v", err)
			continue
		}

		go c.ConsumeMessage(notification.NotificationType)
	}

	return nil
}

func (c *ConsumeRabbitMQ) SendNotification(notificationEntity entity.NotificationEntity) {
	switch notificationEntity.NotificationType {
	case "EMAIL":
		err := c.emailService.SendEmailNotification(*notificationEntity.Email, *notificationEntity.Subject, notificationEntity.Message)
		if err != nil {
			log.Errorf("Failed to send email notification: %v", err)
		}
	case "PUSH":
		c.notificationService.SendPushNotification(context.Background(), notificationEntity)
	}
}

func NewConsumeRabbitMQ(emailService messaging.MessageEmailInterface, notifRepository repositories.NotifRepositoryInterface) ConsumeRabbitMQInterface {
	return &ConsumeRabbitMQ{emailService: emailService, notifRepository: notifRepository}
}

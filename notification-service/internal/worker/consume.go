package worker

import (
	"context"
	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/config"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/infrastructure/messaging"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/repositories"
	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/services"

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
		log.Infof("Received message from queue %s: %s", queueName, string(msg.Body))

		var notification entity.NotificationEntity
		if err := json.Unmarshal(msg.Body, &notification); err != nil {
			log.Errorf("Failed to unmarshal: %v", err)
			continue
		}

		log.Infof("Parsed notification: %+v", notification)

		// Validasi field yang diperlukan
		if notification.ReceiverEmail == nil || *notification.ReceiverEmail == "" {
			log.Errorf("ReceiverEmail is empty for notification")
			continue
		}

		if notification.Message == "" {
			log.Errorf("Message is empty for notification")
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

		log.Infof("Notification created successfully, calling SendNotification")
		go c.SendNotification(notification)
	}

	return nil
}

func (c *ConsumeRabbitMQ) SendNotification(notificationEntity entity.NotificationEntity) {
	switch notificationEntity.NotificationType {
	case "EMAIL":
		err := c.emailService.SendEmailNotification(*notificationEntity.ReceiverEmail, *notificationEntity.Subject, notificationEntity.Message)
		if err != nil {
			log.Errorf("Failed to send email notification: %v", err)
		}
	case "PUSH":
		c.notificationService.SendPushNotification(context.Background(), notificationEntity)
	}
}

func NewConsumeRabbitMQ(emailService messaging.MessageEmailInterface, notifRepository repositories.NotifRepositoryInterface, notificationService services.NotificationServiceInterface) ConsumeRabbitMQInterface {
	return &ConsumeRabbitMQ{
		emailService:        emailService,
		notifRepository:     notifRepository,
		notificationService: notificationService,
	}
}

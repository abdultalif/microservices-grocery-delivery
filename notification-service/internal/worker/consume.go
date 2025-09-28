package worker

import (
	"encoding/json"
	"notification-service/internal/config"
	"notification-service/internal/domain/entity"
	"notification-service/internal/infrastructure/messaging"

	"github.com/labstack/gommon/log"
)

type ConsumeRabbitMQInterface interface {
	ConsumeMessage(queueName string) error
}

type ConsumeRabbitMQ struct {
	emailService messaging.MessageEmailInterface
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
		log.Infof("Got notification: %+v", notification)
		c.emailService.SendEmailNotification(notification.Email, queueName, notification.Message)
	}

	return nil
}

func NewConsumeRabbitMQ(emailService messaging.MessageEmailInterface) ConsumeRabbitMQInterface {
	return &ConsumeRabbitMQ{emailService: emailService}
}

package message

import (
	"encoding/json"
	"order-service/config"
	"order-service/internal/core/domain/entity"
	"order-service/utils"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type PublishRabbitMQInterface interface {
	PublishUpdateStock(productID uuid.UUID, quantity int64)
	PublishOrderToQueue(order entity.OrderEntity) error
	PublishSendEmailUpdateStatus(email, message, queuename string, userID int64) error
}

type PublishRabbitMQ struct {
	cfg *config.Config
}

// PublishSendEmailUpdateStatus implements PublishRabbitMQInterface.
func (p *PublishRabbitMQ) PublishSendEmailUpdateStatus(email, message, queuename string, userID int64) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	queue, err := ch.QueueDeclare(
		queuename,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-3] Failed to declare a queue: %v", err)
		return err
	}

	notifType := "EMAIL"
	if queuename == utils.PUSH_NOTIF {
		notifType = "PUSH"
	}

	notification := map[string]interface{}{
		"receiver_email":    email,
		"message":           message,
		"subject":           "Update Status Order",
		"type":              "UPDATE_STATUS",
		"receiver_id":       userID,
		"notification_type": notifType,
	}

	body, err := json.Marshal(notification)
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-4] Failed to marshal JSON: %v", err)
		return err
	}

	return ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// PublishOrderToQueue implements PublishRabbitMQInterface.
func (p *PublishRabbitMQ) PublishOrderToQueue(order entity.OrderEntity) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishOrderToQueue-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishOrderToQueue-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		p.cfg.Publisher.OrderPublishName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[PublishOrderToQueue-3] Failed to declare queue: %v", err)
		return err
	}

	data, _ := json.Marshal(order)
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil {
		log.Errorf("[PublishOrderToQueue-4] Failed to publish message: %v", err)
		return err
	}

	return nil
}

// PublishUpdateStock implements PublishRabbitMQInterface.
func (p *PublishRabbitMQ) PublishUpdateStock(productID uuid.UUID, quantity int64) {

	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishUpdateStock-1] Failed to connect to RabbitMQ: %v", err)
		return
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishUpdateStock-2] Failed to open a channel: %v", err)
		return
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(p.cfg.Publisher.ProductUpdateStock, true, false, false, false, nil)
	if err != nil {
		log.Errorf("[PublishUpdateStock-3] Failed to declare a queue: %v", err)
		return
	}

	order := entity.PublishOrderItemEntity{
		ProductID: productID,
		Quantity:  quantity,
	}

	data, err := json.Marshal(order)
	if err != nil {
		log.Errorf("[PublishUpdateStock-4] Failed to marshal order: %v", err)
		return
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil {
		log.Errorf("[PublishUpdateStock-5] Failed to publish message: %v", err)
		return
	}

	log.Info("Pesan order dikirim ke RabbitMQ")

}

func NewPublisherRabbitMQ(cfg *config.Config) PublishRabbitMQInterface {
	return &PublishRabbitMQ{
		cfg: cfg,
	}
}

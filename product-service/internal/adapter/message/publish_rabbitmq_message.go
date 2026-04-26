package message

import (
	"encoding/json"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/utils"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type Event struct {
	EventID   string               `json:"event_id"`
	EventType string               `json:"event_type"`
	Source    string               `json:"source"`
	Data      entity.ProductEntity `json:"data"`
	CreatedAt time.Time            `json:"created_at"`
}

func PublishProduct(eventType string, routingKey string, payload entity.ProductEntity) error {

	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[Publisher-1] Failed connect RMQ: %v", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[Publisher-2] Failed open channel: %v", err)
		return err
	}
	defer ch.Close()

	exchangeName := utils.PRODUCT_EXCHANGE

	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[Publisher-3] Failed declare exchange: %v", err)
		return err
	}

	event := Event{
		EventID:   uuid.NewString(),
		EventType: eventType,
		Source:    "product-service",
		Data:      payload,
		CreatedAt: time.Now(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Errorf("[Publisher-4] Marshal error: %v", err)
		return err
	}

	err = ch.Publish(
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		log.Errorf("[Publisher-5] Publish error: %v", err)
		return err
	}

	log.Infof("[Event Published] %s (%s)", eventType, routingKey)
	return nil
}

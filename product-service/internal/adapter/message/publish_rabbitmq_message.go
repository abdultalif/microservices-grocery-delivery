package message

import (
	"encoding/json"
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type PublishRabbitMQInterface interface {
	PublishProductToQueue(product entity.ProductEntity) error
	DeleteProductFromQueue(productID uuid.UUID) error
	PublishProductToOrderService(product interface{}) error
}

type PublishRabbitMQ struct {
	cfg *config.Config
}

func NewPublishRabbitMQ(cfg *config.Config) PublishRabbitMQInterface {
	return &PublishRabbitMQ{cfg: cfg}
}

func (p *PublishRabbitMQ) PublishProductToOrderService(product interface{}) error {

	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishProductToOrderService-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishProductToOrderService-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		p.cfg.PublisherName.ProductToOrder,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[PublishProductToOrderService-3] Failed to declare queue: %v", err)
		return err
	}

	data, _ := json.Marshal(product)
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
		log.Errorf("[DeleteProductFromQueue-4] Failed to publish message: %v", err)
		return err
	}

	return nil

}

// DeleteProductFromQueue implements PublishRabbitMQInterface.
func (p *PublishRabbitMQ) DeleteProductFromQueue(productID uuid.UUID) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[DeleteProductFromQueue-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[DeleteProductFromQueue-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()
	q, err := ch.QueueDeclare(
		p.cfg.PublisherName.ProductDelete,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[DeleteProductFromQueue-3] Failed to declare queue: %v", err)
		return err
	}

	data, _ := json.Marshal(map[string]string{"ProductID": fmt.Sprintf("%d", productID)})
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
		log.Errorf("[DeleteProductFromQueue-4] Failed to publish message: %v", err)
		return err
	}

	return nil
}

func (p *PublishRabbitMQ) PublishProductToQueue(product entity.ProductEntity) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishProductToQueue-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishProductToQueue-2] Failed to open a channel: %v", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		p.cfg.PublisherName.ProductPublish,
		true, false, false, false, nil,
	)
	if err != nil {
		log.Errorf("[PublishProductToQueue-3] Failed to declare queue: %v", err)
		return err
	}

	childMsgs := make([]entity.ProductChildEntity, len(product.Child))
	for i, c := range product.Child {
		childMsgs[i] = entity.ProductChildEntity{
			ID:           c.ID,
			Image:        c.Image,
			Weight:       c.Weight,
			Stock:        c.Stock,
			RegulerPrice: c.RegulerPrice,
			SalePrice:    c.SalePrice,
		}
	}

	var parentID *uuid.UUID
	if product.ParentID != nil {
		str := product.ParentID
		parentID = str
	}

	message := entity.ProductEntity{
		ID:           product.ID,
		CategorySlug: product.CategorySlug,
		ParentID:     parentID,
		Name:         product.Name,
		Image:        product.Image,
		Description:  product.Description,
		RegulerPrice: product.RegulerPrice,
		SalePrice:    product.SalePrice,
		Unit:         product.Unit,
		Weight:       product.Weight,
		Stock:        product.Stock,
		Variant:      product.Variant,
		Status:       product.Status,
		CategoryName: product.CategoryName,
		Child:        childMsgs,
		CreatedAt:    product.CreatedAt,
	}

	data, _ := json.Marshal(message)

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		log.Errorf("[PublishProductToQueue-4] Failed to publish message: %v", err)
		return err
	}

	log.Infof("[PublishProductToQueue] Message published to queue: %v", string(data))
	return nil
}

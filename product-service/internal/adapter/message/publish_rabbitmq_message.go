package message

import (
	"encoding/json"
	"product-service/config"
	"product-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

func PublishProductToQueue(product entity.ProductEntity) error {
	cfg := config.NewConfig()

	conn, err := cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("RabbitMQ connection failed: %v", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Channel creation failed: %v", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("elasticsearch_indexing", true, false, false, false, nil)
	if err != nil {
		log.Errorf("Queue declare failed: %v", err)
		return err
	}

	body, err := json.Marshal(product)
	if err != nil {
		log.Errorf("Marshal failed: %v", err)
		return err
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Errorf("Publish failed: %v", err)
		return err
	}

	log.Infof("Published product ID: %s", product.ID.String())
	return nil
}

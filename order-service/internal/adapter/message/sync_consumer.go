package message

import (
	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	"github.com/labstack/gommon/log"
)

func ConsumeUserUpdated() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumeUserUpdated-1] Failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeUserUpdated-2] Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"user.updated",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[ConsumeUserUpdated-3] Failed to declare queue: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("[ConsumeUserUpdated-4] Failed to register consumer: %v", err)
		return
	}

	db, err := config.NewConfig().ConnectionPostgres()
	if err != nil {
		log.Fatalf("[ConsumeUserUpdated-5] Failed to connect DB: %v", err)
		return
	}

	localDataRepo := repository.NewLocalDataRepository(db.DB)

	log.Info("RabbitMQ Consumer user.updated started...")

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			var buyer entity.CustomerResponseEntity
			if err := json.Unmarshal(msg.Body, &buyer); err != nil {
				log.Errorf("[ConsumeUserUpdated-6] Failed to parse message: %v", err)
				msg.Nack(false, false)
				continue
			}

			if err := localDataRepo.UpsertBuyer(nil, buyer); err != nil {
				log.Errorf("[ConsumeUserUpdated-7] Failed to upsert buyer ID %d: %v", buyer.ID, err)
				msg.Nack(false, true) // requeue
				continue
			}

			log.Infof("[ConsumeUserUpdated-8] Buyer %d synced to local DB", buyer.ID)
			msg.Ack(false)
		}
	}()

	log.Info("[ConsumeUserUpdated] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func ConsumeProductUpdated() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumeProductUpdated-1] Failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeProductUpdated-2] Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"product.updated",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[ConsumeProductUpdated-3] Failed to declare queue: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("[ConsumeProductUpdated-4] Failed to register consumer: %v", err)
		return
	}

	db, err := config.NewConfig().ConnectionPostgres()
	if err != nil {
		log.Fatalf("[ConsumeProductUpdated-5] Failed to connect DB: %v", err)
		return
	}

	localDataRepo := repository.NewLocalDataRepository(db.DB)

	log.Info("RabbitMQ Consumer product.updated started...")

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			var product entity.ProductResponseEntity
			if err := json.Unmarshal(msg.Body, &product); err != nil {
				log.Errorf("[ConsumeProductUpdated-6] Failed to parse message: %v", err)
				msg.Nack(false, false)
				continue
			}

			if err := localDataRepo.UpsertProduct(nil, product); err != nil {
				log.Errorf("[ConsumeProductUpdated-7] Failed to upsert product %s: %v", product.ID, err)
				msg.Nack(false, true) // requeue
				continue
			}

			log.Infof("[ConsumeProductUpdated-8] Product %s synced to local DB", product.ID)
			msg.Ack(false)
		}
	}()

	log.Info("[ConsumeProductUpdated] Waiting for messages. To exit press CTRL+C")
	<-forever
}

package message

import (
	"context"
	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/payment-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/entity"
	"github.com/labstack/gommon/log"
)

func ConsumeUserEvents() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumeUserEvents-1] RMQ error: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeUserEvents-2] Channel error: %v", err)
		return
	}
	defer ch.Close()

	exchangeName := "user.events"
	queueName := "user.events.payment"

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
		log.Fatalf("[ConsumeUserEvents-3] Exchange error: %v", err)
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[ConsumeUserEvents-4] Queue error: %v", err)
	}

	err = ch.QueueBind(
		q.Name,
		"user.*",
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[ConsumeUserEvents-5] Bind error: %v", err)
	}

	err = ch.Qos(
		10,
		0,
		false,
	)
	if err != nil {
		log.Errorf("[ConsumeUserEvents-6] QoS error: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[ConsumeUserEvents-7] Consume error: %v", err)
	}

	db, _ := config.NewConfig().ConnectionPostgres()
	localRepo := repository.NewLocalDataRepository(db.DB)

	log.Infof("Consumer %s started...", queueName)

	go func() {
		for msg := range msgs {

			var event struct {
				EventType string                        `json:"event_type"`
				Data      entity.CustomerResponseEntity `json:"data"`
			}

			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Errorf("[ConsumeUserEvents-8] Unmarshal error: %v", err)
				msg.Nack(false, false)
				continue
			}

			switch event.EventType {

			case "user.created", "user.updated":
				err := localRepo.UpsertBuyer(context.Background(), event.Data)
				if err != nil {
					log.Errorf("[ConsumeUserEvents-9] Upsert error: %v", err)
					msg.Nack(false, true)
					continue
				}

			case "user.deleted":
				err := localRepo.DeleteBuyer(context.Background(), event.Data.ID)
				if err != nil {
					log.Errorf("[ConsumeUserEvents-10] Delete error: %v", err)
					msg.Nack(false, true)
					continue
				}

			default:
				log.Warnf("[ConsumeUserEvents] Unknown event: %s", event.EventType)
			}

			msg.Ack(false)
		}
	}()

	forever := make(chan bool)
	<-forever
}

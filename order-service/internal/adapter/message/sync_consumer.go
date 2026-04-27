package message

import (
	"context"
	"encoding/json"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/utils"
	"github.com/labstack/gommon/log"
)

type Event struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	Data      entity.ProductSnapshot `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

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

	q, err := ch.QueueDeclare(
		"user.events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[ConsumeUserEvents-3] Queue error: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("[ConsumeUserEvents-4] Consume error: %v", err)
	}

	db, _ := config.NewConfig().ConnectionPostgres()
	localRepo := repository.NewLocalDataRepository(db.DB)

	log.Info("Consumer user.events started...")

	go func() {
		for msg := range msgs {

			var event struct {
				EventType string                        `json:"event_type"`
				Data      entity.CustomerResponseEntity `json:"data"`
			}

			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Errorf("[ConsumeUserEvents-5] Unmarshal error: %v", err)
				msg.Nack(false, false)
				continue
			}

			switch event.EventType {

			case "user.created", "user.updated":
				err := localRepo.UpsertBuyer(context.Background(), event.Data)
				if err != nil {
					msg.Nack(false, true)
					continue
				}

			case "user.deleted":
				err := localRepo.DeleteBuyer(context.Background(), event.Data.ID)
				if err != nil {
					msg.Nack(false, true)
					continue
				}
			}

			msg.Ack(false)
		}
	}()

	forever := make(chan bool)
	<-forever
}

func ConsumeProductEvents() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumeProductEvents-1] Failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeProductEvents-2] Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		utils.PRODUCT_EXCHANGE,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-3] Failed declare exchange: %v", err)
		return
	}

	q, err := ch.QueueDeclare(
		utils.ORDER_PRODUCT_QUEUE,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-4] Failed to declare queue: %v", err)
		return
	}

	err = ch.QueueBind(
		q.Name,
		"product.*",
		utils.PRODUCT_EXCHANGE,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-5] Failed bind queue: %v", err)
		return
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-6] Failed set QoS: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-7] Failed to register consumer: %v", err)
		return
	}

	db, _ := config.NewConfig().ConnectionPostgres()
	localDataRepo := repository.NewLocalDataRepository(db.DB)

	log.Infof("[ConsumeProductEvents-8] Listening on queue: %s", q.Name)

	for msg := range msgs {

		var event Event

		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Errorf("[ConsumeProductEvents-9] Failed to unmarshal message: %v", err)
			msg.Ack(false)
			continue
		}

		var processErr error

		switch msg.RoutingKey {

		case utils.PRODUCT_CREATED_RK, utils.PRODUCT_UPDATED_RK:

			parent := entity.ProductSnapshot{
				ID:           event.Data.ID,
				Name:         event.Data.Name,
				Stock:        event.Data.Stock,
				Image:        event.Data.Image,
				RegulerPrice: event.Data.RegulerPrice,
				SalePrice:    event.Data.SalePrice,
				Unit:         event.Data.Unit,
				Weight:       event.Data.Weight,
				CreatedAt:    event.Data.CreatedAt,
			}

			processErr = localDataRepo.UpsertProduct(context.Background(), parent)
			if processErr != nil {
				break
			}

			for _, child := range event.Data.Child {

				childData := entity.ProductSnapshot{
					ID:           child.ID,
					ParentID:     &event.Data.ID,
					Name:         event.Data.Name,
					Stock:        child.Stock,
					Image:        child.Image,
					RegulerPrice: child.RegulerPrice,
					SalePrice:    child.SalePrice,
					Unit:         child.Unit,
					Weight:       child.Weight,
					CreatedAt:    child.CreatedAt,
				}

				processErr = localDataRepo.UpsertProduct(context.Background(), childData)
				if processErr != nil {
					break
				}
			}

		case utils.PRODUCT_DELETED_RK:

			processErr = localDataRepo.DeleteProduct(context.Background(), event.Data.ID)
		}

		if processErr != nil {
			log.Errorf("[ConsumeProductEvents-10] Process error: %v", processErr)

			msg.Ack(false)
			continue
		}

		log.Infof("[ConsumeProductEvents-11] Product %s processed (%s)", event.Data.ID, msg.RoutingKey)
		msg.Ack(false)
	}
}
